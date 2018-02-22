package video

import (
	"time"
	"log"
	"github.com/pkg/errors"
	"github.com/thenrich/go-surv/config"
	"github.com/thenrich/go-surv/cloud/aws"
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
}

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

func (ch *CameraHandler) StartStreams() {
	for _, cam := range ch.cameras {
		log.Printf("Starting stream for %s", cam.Name)
		stream := NewStream(cam)

		if ch.cfg.Storage == "s3" {
			if !ch.cfg.AWS.Ready() {
				log.Fatal("Missing AWS configuration")
			}
			cw := aws.NewS3Storage(ch.cfg.AWS, ch.cfg.AWS.S3Bucket)
			stream.AddWriter(NewCloudStorage(cam.Name, cam.recordInterval, ch.cfg, cw))
		}

		ch.streams[cam.Name] = stream

		go func(stream *Stream) {
			err := stream.Stream()
			if err != nil {
				log.Println(errors.Wrapf(err, "error from stream"))
			}
		}(stream)

		go func(stream *Stream, cam *Camera) {
			for {
				select {
				case s := <-stream.Stills():
					log.Printf("send still for %s", cam.Name)
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
