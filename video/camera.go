package video

import (
	"github.com/pkg/errors"
	"time"
	"log"
	"github.com/thenrich/go-surv/config"
)

// CameraStreamer defines the behavior for camera handlers
type CameraStreamer interface {
	Camera(name string) *Camera
	StartStreams()
}

type Camera struct {
	// Name of camera
	Name string

	// Latest still image in JPEG format
	LatestImage []byte

	// SourceURL defines the video source
	SourceURL string

	// interval to record
	recordInterval time.Duration

	// writers
	writers []Writer
}

// AddWriter adds packet writers to this camera
func (c *Camera) AddWriter(w Writer) {
	c.writers = append(c.writers, w)
}

// NewCamera creates a new camera instance
func NewCamera(name string, source string, recordInterval time.Duration) *Camera {
	return &Camera{Name: name, SourceURL: source, recordInterval: recordInterval}
}

type CameraHandler struct {
	cfg *config.Config

	// cameras we're monitoring
	cameras map[string]*Camera

	// streams for all cameras we're monitoring, indexed by
	// camera name
	streams map[string]*Stream
}

func (ch *CameraHandler) AddCamera(cam *Camera) {
	log.Printf("Add camera %s %s", cam.Name, cam.SourceURL)
	ch.cameras[cam.Name] = cam
}

func (ch *CameraHandler) Camera(name string) *Camera {
	if _, ok := ch.cameras[name]; ok {
		return ch.cameras[name]
	}

	return nil
}

// StartStreams sets up the streams for each camera and begins reading
// camera data.
func (ch *CameraHandler) StartStreams() {
	// Create the streams and add writers
	ch.setupStreams()

	ch.stream()

}

// iterate over all of our streams and start each one
func (ch *CameraHandler) stream() {
	for _, stream := range ch.streams {
		stream.Start()
	}

}

func (ch *CameraHandler) setupStreams() {
	// Setup streams for each camera,
	for _, cam := range ch.cameras {
		log.Printf("Setup stream for %s", cam.Name)
		stream := NewStream(cam)

		// open camera streams
		if err := stream.Open(); err != nil {
			log.Fatal(errors.Wrap(err, "error opening stream"))
		}


		// setup still writer
		// @TODO should the Stills channel be on a stream or the writer?
		still, err := NewStillWriter(stream.Stills(), stream.demuxer)
		if err != nil {
			log.Println(err)
			continue
		}
		if err = still.Open(); err != nil {
			log.Println(errors.Wrap(err, "error opening writer"))
		}

		stream.AddWriter(still)
		//// Add all of the camera writers to the stream
		//for _, w := range cam.writers {
		//	stream.AddWriter(w)
		//}
		//
		ch.streams[cam.Name] = stream
		//
		go func(stream *Stream, cam *Camera) {
			for {
				select {
				case s := <-stream.Stills():
					cam.LatestImage = s.imgData
				}
			}
		}(stream, cam)
	}
}

func (ch *CameraHandler) CloseStreams() {
	for name, stream := range ch.streams {
		log.Printf("Cleaning up stream for %s\n", name)
		stream.Cleanup()
	}
}

func NewCameraHandler(cfg *config.Config) *CameraHandler {
	return &CameraHandler{cfg, make(map[string]*Camera), make(map[string]*Stream)}
}
