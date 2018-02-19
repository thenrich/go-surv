package video

import (
	"time"
	"log"
	"github.com/pkg/errors"
)

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
}

type CameraHandler struct {
	// cameras we're monitoring
	cameras map[string]*Camera

	// streams for all cameras we're monitoring
	streams map[string]*Stream
}

func (ch *CameraHandler) AddCamera(cam *Camera) {
	ch.cameras[cam.Name] = cam
}

func (ch *CameraHandler) Camera(name string) *Camera {
	if _, ok := ch.cameras[name]; ok {
		return ch.cameras[name]
	}

	return nil
}

func (ch *CameraHandler) StartStreams() {
	for _, cam := range ch.cameras {
		log.Printf("Starting stream for %s", cam.Name)
		stream, err := NewStream(cam.SourceURL, time.Minute).WithDst("out.mp4")
		if err != nil {
			log.Println(errors.Wrapf(err, "error starting stream for %s", cam))
			continue
		}
		ch.streams[cam.Name] = stream

		go func() {
			err := stream.Stream()
			if err != nil {
				log.Println(errors.Wrapf(err, "error from stream"))
			}
		}()
		go func() {
			for {
				select {
				case s := <-stream.Stills():
					cam.LatestImage = s.imgData
				}
			}
		}()

	}
}

func (ch *CameraHandler) CloseStreams() {
	for name, stream := range ch.streams {
		log.Printf("Cleaning up stream for %s\n", name)
		stream.Cleanup()
	}
}

func NewCameraHandler() *CameraHandler {
	return &CameraHandler{make(map[string]*Camera), make(map[string]*Stream)}
}
