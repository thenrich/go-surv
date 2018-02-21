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

	"os"
	"github.com/thenrich/go-surv/config"
	"strings"
	"io"
	"image"
)

// Writer defines the interface for writing video packets
type Writer interface {
	Open(streams []av.CodecData) error
	Write(writer av.Packet) error
	Close() error
}

// CloudWriter defines the interface for writing files to cloud providers
type CloudWriter interface {
	UploadFile(r io.ReadCloser, key string) error
}

type CloudStorage struct {
	*LocalWriter
	writer CloudWriter

	nextUploadFile string
	nextUploadTime time.Time
}

func (cs *CloudStorage) Write(pkt av.Packet) error {
	if err := cs.LocalWriter.dst.WritePacket(pkt); err != nil {
		return errors.Wrap(err, "error writing packet to local writer")
	}

	if cs.nextRotation.Before(time.Now().UTC()) {
		log.Println("Time to rotate")
		// Set file to upload
		cs.nextUploadFile = cs.outfile
		cs.nextUploadTime = cs.now
		if err := cs.rotate(); err != nil {
			log.Println(errors.Wrap(err, "error rotating"))
		}
	}

	return nil
}

// Rotate closes the local writer and reopens it at the current time
func (cs *CloudStorage) rotate() error {
	if err := cs.Close(); err != nil {
		return errors.Wrap(err, "rotate: error closing local writer")
	}

	if err := cs.Open(nil); err != nil {
		return errors.Wrap(err, "rotate: error opening new local writer")
	}

	cs.nextRotation = time.Now().UTC().Add(cs.duration)

	return nil
}

// Close begins uploading completed file to S3
func (cs *CloudStorage) Close() error {
	if err := cs.LocalWriter.Close(); err != nil {
		log.Println(errors.Wrap(err, "error closing file in s3 writer"))
	}

	if cs.nextUploadFile != "" {
		go func() {
			f, err := os.Open(cs.nextUploadFile)
			if err != nil {
				log.Println(err)
				return
			}
			defer f.Close()
			key := fmt.Sprintf("%d-%d-%d/%s", cs.nextUploadTime.Year(), cs.nextUploadTime.Month(), cs.nextUploadTime.Day(), strings.Replace(cs.nextUploadFile, "/tmp", "", 1))
			err = cs.writer.UploadFile(f, key)
			if err != nil {
				log.Println(err)
			}
			// Copy our reference to the filename so we can clear nextUploadFile
			deleteFile := cs.nextUploadFile
			cs.nextUploadFile = ""
			if err := os.Remove(deleteFile); err != nil {
				log.Println(err)
				return
			}

		}()
	}

	return nil
}

func NewCloudStorage(name string, interval time.Duration, cfg *config.Config, cloud CloudWriter) *CloudStorage {
	localPath := fmt.Sprintf("/tmp/%s", name)
	return &CloudStorage{LocalWriter: NewLocalWriter(localPath, interval), writer: cloud}
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
	return fmt.Sprintf("%s-%d-%d-%d-%d-%d.mp4", name, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
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

	defer frame.Free()

	// get packet time
	if sw.lastStill == 0 {
		sw.lastStill = pkt.Time
	}

	if pkt.Time-sw.lastStill < time.Duration(1*time.Second) {
		return nil
	}

	sw.lastStill = pkt.Time
	img := frame.Image

	go sw.encodeStill(&img)

	return nil
}

func (sw *StillWriter) encodeStill(img image.Image) {
	var b bytes.Buffer
	jpeg.Encode(&b, img, nil)

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
