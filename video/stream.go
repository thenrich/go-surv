package video

import (
	"github.com/nareix/joy4/av/avutil"
	"log"
	"github.com/nareix/joy4/av"
	"github.com/pkg/errors"
	"io"
	"github.com/nareix/joy4/format"
)

func init() {
	format.RegisterAll()
}

// Still defines an object for holding bytes for still images
type Still struct {
	imgData []byte
}

// Stream supports reading from a Camera and writing to one
// or more writers. This is also reponsible for handling a channel
// that transfers still image data for viewing via HTTP.
type Stream struct {
	// Camera for this stream
	cam *Camera

	// video source
	demuxer av.DemuxCloser

	// outputs
	writers []Writer

	// streams
	streams []av.CodecData

	stills chan *Still
}

// NewStream creates a new stream for a Camera
func NewStream(cam *Camera) *Stream {
	return &Stream{cam: cam, stills: make(chan *Still, 100)}
}

// AddWriter adds a new writer to the stream and opens it
func (s *Stream) AddWriter(w Writer) {
	if err := w.Open(s.streams); err != nil {
		log.Println(errors.Wrap(err, "error opening still writer"))
	}

	s.writers = append(s.writers, w)
}

// Stills returns the channel used for communicating still images
func (s *Stream) Stills() chan *Still {
	return s.stills
}

// openStream opens the source stream
func (s *Stream) openStream() error {
	demux, err := avutil.Open(s.cam.SourceURL)
	if err != nil {
		return errors.Wrapf(err, "error opening %s", s.cam.SourceURL)
	}

	s.demuxer = demux

	return nil
}

// Open camera stream and return the available stream data.
func (s *Stream) Open() ([]av.CodecData, error) {
	s.openStream()

	// Get a reference to the incoming video stream
	var streams []av.CodecData
	var err error
	if streams, err = s.demuxer.Streams(); err != nil {
		return nil, errors.Wrap(err, "error getting streams")
	}

	s.streams = streams

	return s.streams, nil
}

// Read reads a packet from the stream.
func (s *Stream) Read() error {
	// read packets
	var err error
	var pkt av.Packet
	if pkt, err = s.demuxer.ReadPacket(); err != nil {
		if err == io.EOF {
			return err
		}
		return errors.Wrap(err, "error reading packet")
	}

	// Write packet to each writer
	for _, w := range s.writers {
		if err := w.Write(pkt); err != nil {
			log.Println(errors.Wrapf(err, "error writing packet to %s", w))
		}
	}

	return nil

}

// Cleanup closes streams and calls the Close method on each writer
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
