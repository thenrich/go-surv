package http

import (
	"net/http"
	"regexp"
	"log"
	"github.com/thenrich/go-surv/video"
}


func NewCameraHandler(cs video.CameraStreamer) *CameraHandler {
	return &CameraHandler{cs}
}

type CameraHandler struct {
	cameras video.CameraStreamer
}

func (ch *CameraHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	re, err := regexp.Compile("cameras/(?P<Camera>[a-zA-Z0-9_]+)")
	if err != nil {
		log.Fatal(err)
	}
	f := re.FindStringSubmatch(r.URL.Path)

	if len(f) != 2 {
		http.NotFound(w, r)
		return
	}

	cam := ch.cameras.Camera(f[1])
	if cam == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(cam.LatestImage)
	return

}



func NewHandler(cs video.CameraStreamer) http.Handler {
	h := NewRegexHandler()
	h.Handle(regexp.MustCompile("cameras/"), NewCameraHandler(cs))
	h.Handle(regexp.MustCompile("dash$"), NewDashHandler())

	return h
}
