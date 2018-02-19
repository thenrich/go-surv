package main

import (
	"net/http"
	"regexp"
	"log"
	"github.com/thenrich/gosurveil/video"
)

func NewHttpHandler(cs video.CameraStreamer) *HttpHandler {
	return &HttpHandler{cs}
}

type HttpHandler struct {
	cameras video.CameraStreamer
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	re, err := regexp.Compile("cameras/(?P<Camera>[a-zA-Z0-9_]+)")
	if err != nil {
		log.Fatal(err)
	}
	f := re.FindStringSubmatch(r.URL.Path)

	if len(f) != 2 {
		http.NotFound(w, r)
		return
	}

	cam := h.cameras.Camera(f[1])
	if cam == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(cam.LatestImage)
	return

}
