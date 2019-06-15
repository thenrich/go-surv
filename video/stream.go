package video

import (
	"github.com/pkg/errors"
	"io"
	"log"

	"github.com/3d0c/gmf"
)

func init() {

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
	demuxer *demuxer

	// channel of packet data
	data chan []*gmf.Frame

	// outputs
	writers []Writer

	// streams
	//streams []av.CodecData

	// Channel for sending stills
	stills chan *Still
}

// NewStream creates a new stream for a Camera
func NewStream(cam *Camera) *Stream {
	return &Stream{
		cam: cam,
		stills: make(chan *Still, 100),
		data: make(chan []*gmf.Frame),
	}
}

// AddWriter adds a new writer to the stream and opens it
func (s *Stream) AddWriter(w Writer) {
	//if err := w.Open(s.streams); err != nil {
	//	log.Println(errors.Wrap(err, "error opening still writer"))
	//}

	s.writers = append(s.writers, w)
}

// Stills returns the channel used for communicating still images
func (s *Stream) Stills() chan *Still {
	return s.stills
}

// openStream opens the source stream
func (s *Stream) openStream() error {
	// Open video file

	s.demuxer = NewDemuxer(s.cam.SourceURL)

	if err := s.demuxer.open(); err != nil {
		return errors.Wrap(err, "error opening demuxer")
	}

	return nil
}

// Open camera stream and return the available stream data.
func (s *Stream) Open() error {
	if err := s.openStream(); err != nil {
		return errors.Wrap(err, "error opening stream")
	}

	return nil
}

// Start the stream.
//
// Here we start the writers goroutine which starts reading from the
// data channel and writes the packet to all writers when one is received.
// Then we start the reader to read the packets from the demuxer buffer.
func (s *Stream) Start() {
	go s.startWriters()
	go s.startReader()
}

func (s *Stream) startWriters() {
	for {
		select {
		case frames := <-s.data:
			for _, w := range s.writers {
				if err := w.Write(frames); err != nil {
					log.Println(errors.Wrapf(err, "error writing packet to %s", w))
				}
			}

			for i := range frames {
				frames[i].Free()
			}


		}
	}
}

func (s *Stream) startReader() {
	for {
		// read packets
		var err error
		var frames []*gmf.Frame
		if frames, err = s.demuxer.ReadFrames(); err != nil {
			if err == io.EOF {
				break
			}
			log.Println(errors.Wrap(err, "error reading packet"))
			break
		}


		s.data <- frames

		//// Write packet to each writer
		//for _, w := range s.writers {
		//	if err := w.Write(pkt); err != nil {
		//		log.Println(errors.Wrapf(err, "error writing packet to %s", w))
		//	}
		//}
	}
}

// Cleanup closes streams and calls the Close method on each writer
func (s *Stream) Cleanup() {
	//for id := range s.writers {
	//	if err := s.writers[id].Close(); err != nil {
	//		log.Println(errors.Wrapf(err, "error closing %s", s.writers[id]))
	//	}
	//}

	if err := s.demuxer.Close(); err != nil {
		log.Println(err)
	}
}
