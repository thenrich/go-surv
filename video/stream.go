package video

import (
	"github.com/nareix/joy4/av/avutil"
	"log"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/pkg/errors"
	"io"
)

type Still struct {
	imgData []byte
}

type Stream struct {
	cam *Camera

	demuxer      av.DemuxCloser
	writers      []Writer
	videoDecoder *ffmpeg.VideoDecoder

	stills chan *Still
}

func NewStream(cam *Camera) *Stream {
	return &Stream{cam: cam, stills: make(chan *Still, 100)}
}

func (s *Stream) AddWriter(w Writer) {
	s.writers = append(s.writers, w)
}

func (s *Stream) Stills() chan *Still {
	return s.stills
}

func (s *Stream) openStream() error {
	demux, err := avutil.Open(s.cam.SourceURL)
	if err != nil {
		return errors.Wrapf(err, "error opening %s", s.cam.SourceURL)
	}

	s.demuxer = demux

	return nil
}

func (s *Stream) Stream() error {
	// Open the camera source
	s.openStream()

	// Get a reference to the incoming video stream
	var streams []av.CodecData
	var err error
	if streams, err = s.demuxer.Streams(); err != nil {
		return errors.Wrap(err, "error getting streams")
	}

	// Setup still writer
	still, err := NewStillWriter(streams, s.stills)
	if err != nil {
		log.Println(errors.Wrap(err, "error creating still writer"))
	}
	// add still writer to our slice of writers
	s.writers = append(s.writers, still)

	for id := range s.writers {
		if err := s.writers[id].Open(streams); err != nil {
			log.Println(errors.Wrap(err, "error opening still writer"))
		}
	}

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

		// Write packet to each writer
		for id := range s.writers {
			if err := s.writers[id].Write(pkt); err != nil {
				log.Println(errors.Wrapf(err, "error writing packet to %s", s.writers[id]))
			}
		}

	}

	s.Cleanup()
	return nil

}

func (s *Stream) Cleanup() {
	for id := range s.writers {
		if err := s.writers[id].Close(); err != nil {
			log.Println(errors.Wrapf(err, "error closing %s", s.writers[id]))
		}
	}

	if err := s.demuxer.Close(); err != nil {
		log.Println(err)
	}
}
