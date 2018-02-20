package video

import (
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/nareix/joy4/av"
	"github.com/pkg/errors"
	"bytes"
	"image/jpeg"
	"fmt"
	"github.com/nareix/joy4/av/avutil"
	"time"
	"log"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/thenrich/go-surv/cloud"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/thenrich/go-surv/config"
	"strings"
)

type Writer interface {
	Open(streams []av.CodecData) error
	Write(writer av.Packet) error
	Close() error
}

type S3Writer struct {
	*LocalWriter
	S3     *s3manager.Uploader
	Bucket string

	nextUploadFile string
	nextUploadTime time.Time
}

func (s3 *S3Writer) Write(pkt av.Packet) error {
	if err := s3.LocalWriter.dst.WritePacket(pkt); err != nil {
		return errors.Wrap(err, "error writing packet to local writer")
	}

	if s3.nextRotation.Before(time.Now().UTC()) {
		log.Println("Time to rotate")
		// Set file to upload
		s3.nextUploadFile = s3.outfile
		s3.nextUploadTime = s3.now
		if err := s3.rotate(); err != nil {
			log.Println(errors.Wrap(err, "error rotating"))
		}
	}

	return nil
}

// Rotate closes the local writer and reopens it at the current time
func (s3 *S3Writer) rotate() error {
	if err := s3.Close(); err != nil {
		return errors.Wrap(err, "rotate: error closing local writer")
	}

	if err := s3.Open(nil); err != nil {
		return errors.Wrap(err, "rotate: error opening new local writer")
	}

	s3.nextRotation = time.Now().UTC().Add(s3.duration)

	return nil
}

// Close begins uploading completed file to S3
func (s3 *S3Writer) Close() error {
	if err := s3.LocalWriter.Close(); err != nil {
		log.Println(errors.Wrap(err, "error closing file in s3 writer"))
	}

	if s3.nextUploadFile != "" {
		go func() {
			f, err := os.Open(s3.nextUploadFile)
			if err != nil {
				log.Println(err)
				return
			}
			defer f.Close()
			key := fmt.Sprintf("%d-%s-%d/%s", s3.nextUploadTime.Year(), s3.nextUploadTime.Month(), s3.nextUploadTime.Day(), strings.Replace(s3.nextUploadFile, "/tmp", "", 1))
			result, err := s3.S3.Upload(&s3manager.UploadInput{
				Bucket: aws.String(s3.Bucket),
				Key:    aws.String(key),
				Body:   f,
			})
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(result.Location)
			// Copy our reference to the filename so we can clear nextUploadFile
			deleteFile := s3.nextUploadFile
			s3.nextUploadFile = ""
			if err := os.Remove(deleteFile); err != nil {
				log.Println(err)
				return
			}

		}()
	}

	return nil
}

func NewS3Writer(name string, interval time.Duration, cfg *config.Config) *S3Writer {
	localPath := fmt.Sprintf("/tmp/%s", name)
	return &S3Writer{LocalWriter: NewLocalWriter(localPath, interval), S3: cloud.Uploader(), Bucket: cfg.AWS.S3Bucket}
}

type LocalWriter struct {
	name         string
	outfile      string
	dst          av.MuxCloser
	duration     time.Duration
	nextRotation time.Time
	now          time.Time
	streams      []av.CodecData
}

func (lw *LocalWriter) Open(streams []av.CodecData) error {
	// Create timestamp for the current time
	lw.now = time.Now().UTC()
	lw.outfile = lw.filename(lw.name, lw.now)

	// Set streams
	if streams != nil {
		lw.streams = streams
	}

	dst, err := avutil.Create(lw.outfile)
	if err != nil {
		return errors.Wrap(err, "error creating dest")
	}
	lw.dst = dst

	if err := lw.dst.WriteHeader(lw.streams); err != nil {
		return errors.Wrap(err, "error writing header for local writer")
	}

	return nil
}

// Rotate closes the local writer and reopens it at the current time
func (lw *LocalWriter) rotate() error {
	if err := lw.Close(); err != nil {
		return errors.Wrap(err, "rotate: error closing local writer")
	}

	if err := lw.Open(nil); err != nil {
		return errors.Wrap(err, "rotate: error opening new local writer")
	}

	lw.nextRotation = time.Now().UTC().Add(lw.duration)

	return nil
}

func (lw *LocalWriter) filename(name string, t time.Time) string {
	return fmt.Sprintf("%s-%d-%s-%d-%d-%d.mp4", name, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

func (lw *LocalWriter) Close() error {
	if err := lw.dst.WriteTrailer(); err != nil {
		return errors.Wrap(err, "error writing trailer to local writer")
	}
	if err := lw.dst.Close(); err != nil {
		return errors.Wrap(err, "error closing local writer")
	}

	return nil
}

func (lw *LocalWriter) Write(pkt av.Packet) error {
	if err := lw.dst.WritePacket(pkt); err != nil {
		return errors.Wrap(err, "error writing packet to local writer")
	}

	if lw.nextRotation.Before(time.Now().UTC()) {
		log.Println("Time to rotate")
		if err := lw.rotate(); err != nil {
			log.Println(errors.Wrap(err, "error rotating"))
		}
	}

	return nil
}

func NewLocalWriter(name string, interval time.Duration) *LocalWriter {
	return &LocalWriter{name: name, duration: interval, nextRotation: time.Now().UTC().Add(interval)}

}

type StillWriter struct {
	// FFmpeg decoder
	videoDecoder *ffmpeg.VideoDecoder

	// Still channel
	stills chan *Still

	lastStill time.Duration
}

func (sw *StillWriter) Open(streams []av.CodecData) error {
	return nil
}

func (sw *StillWriter) Close() error {
	return nil
}

func (sw *StillWriter) Write(pkt av.Packet) error {
	frame, err := sw.videoDecoder.Decode(pkt.Data)
	if err != nil {
		return errors.Wrap(err, "error decoding packet data")
	}

	if frame == nil {
		return nil
	}

	// get packet time
	if sw.lastStill == 0 {
		sw.lastStill = pkt.Time
	}

	if pkt.Time-sw.lastStill < time.Duration(1*time.Second) {
		return nil
	}

	sw.lastStill = pkt.Time

	go sw.encodeStill(frame)

	return nil
}

func (sw *StillWriter) encodeStill(frame *ffmpeg.VideoFrame) {
	defer frame.Free()

	var b bytes.Buffer
	jpeg.Encode(&b, &frame.Image, nil)

	sw.stills <- &Still{b.Bytes()}

}

func NewStillWriter(streams []av.CodecData, ch chan *Still) (*StillWriter, error) {
	// get video stream from streams
	vstream, err := extractVideoStream(streams)
	if err != nil {
		return nil, errors.Wrap(err, "error reading stream")
	}
	decoder, err := ffmpeg.NewVideoDecoder(vstream)
	if err != nil {
		return nil, errors.Wrap(err, "error creating video decoder")
	}

	if err := decoder.Setup(); err != nil {
		return nil, errors.Wrap(err, "error setting up video decoder")
	}

	return &StillWriter{videoDecoder: decoder, stills: ch}, nil
}
