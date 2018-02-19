package video

import (
	"github.com/nareix/joy4/av/avutil"
	"log"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"time"
	"github.com/pkg/errors"
	"io"
	"image/jpeg"
	"bytes"
)

type Still struct {
	imgData []byte
}

type Stream struct {
	source string
	max    time.Duration

	demuxer      av.Demuxer
	dst          av.MuxCloser
	videoDecoder *ffmpeg.VideoDecoder

	stills chan *Still
}

func NewStream(url string, max time.Duration) *Stream {
	return &Stream{source: url, max: max, stills: make(chan *Still, 100)}
}

func (s *Stream) WithDest(dst string) (*Stream, error) {
	dest, err := avutil.Create(dst)
	if err != nil {
		return nil, errors.Wrap(err, "error creating dest")
	}
	s.dst = dest
	return s, nil
}

func (s *Stream) Stills() chan *Still {
	return s.stills
}

func (s *Stream) openStream() error {
	demux, err := avutil.Open(s.source)
	if err != nil {
		return errors.Wrapf(err, "error opening %s", s.source)
	}

	s.demuxer = demux

	return nil
}

func (s *Stream) copy() error {
	// create a video decoder for still images
	lastStillTime := time.Duration(0)

	// read packets
	for {
		var err error
		var pkt av.Packet
		if pkt, err = s.demuxer.ReadPacket(); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return errors.Wrap(err, "error reading packet")
		}

		if s.max > 0 && pkt.Time >= s.max {
			return nil
		}

		// Write to mp4
		if err := s.dst.WritePacket(pkt); err != nil {
			return errors.Wrap(err, "error writing packet")
		}

		frame, err := s.videoDecoder.Decode(pkt.Data)
		if err != nil {
			return errors.Wrap(err, "error decoding packet data")
		}

		if frame == nil {
			continue
		}

		// get packet time
		if lastStillTime == 0 {
			lastStillTime = pkt.Time
		}

		if pkt.Time-lastStillTime < time.Duration(1*time.Second) {
			continue
		}

		lastStillTime = pkt.Time

		go encodeStill(frame, s.stills)

	}

	return nil
}

func (s *Stream) getVideoStreams() ([]av.CodecData, av.CodecData, error) {
	var streams []av.CodecData
	var err error
	if streams, err = s.demuxer.Streams(); err != nil {
		return nil, nil, errors.Wrap(err, "error getting streams")
	}

	vstream, err := extractVideoStream(streams)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error locating video stream")
	}

	return streams, vstream, nil
}

func (s *Stream) Stream() error {
	s.openStream()
	allStreams, videoStream, err := s.getVideoStreams()
	if err != nil {
		return errors.Wrap(err, "error getting video stream")
	}

	// create a video decoder using the video stream
	decoder, err := ffmpeg.NewVideoDecoder(videoStream)
	if err != nil {
		log.Fatal(err)
	}

	s.videoDecoder = decoder
	if err := s.videoDecoder.Setup(); err != nil {
		return errors.Wrap(err, "error setting up video decoder")
	}

	if err := s.dst.WriteHeader(allStreams); err != nil {
		return errors.Wrap(err, "error writing header")
	}

	if err = s.copy(); err != nil {
		log.Fatal(err)
	}

	s.Cleanup()
	return nil

}

func (s *Stream) Cleanup() {
	if err := s.dst.WriteTrailer(); err != nil {
		log.Fatal(err)
	}
}

func encodeStill(frame *ffmpeg.VideoFrame, stills chan *Still) {
	defer frame.Free()

	var b bytes.Buffer
	jpeg.Encode(&b, &frame.Image, nil)

	stills <- &Still{b.Bytes()}

}

func extractVideoStream(streams []av.CodecData) (av.CodecData, error) {
	for _, stream := range streams {
		if stream.Type() == av.H264 {
			return stream, nil
		}
	}

	return nil, errors.New("no h264 stream")
}
