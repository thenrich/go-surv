package http

import (
	"net/http"
	"regexp"
	"log"
	"github.com/thenrich/go-surv/video"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexHandler struct {
	routes []*route
}

func (rh *RegexHandler) Handle(pattern *regexp.Regexp, handler http.Handler) {
	rh.routes = append(rh.routes, &route{pattern, handler})
}

func (rh *RegexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range rh.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

func NewRegexHandler() *RegexHandler {
	return &RegexHandler{}
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

func NewDashHandler() *DashHandler {
	return &DashHandler{}
}

type DashHandler struct {}
func (dh *DashHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	html := `
<html>
 <style>
 img {
    width: 30%;
 }
 </style>
 <body>
  <div><img src="/cameras/front_door"></div>
  <div><img src="/cameras/back_door"></div>
 </body>
</html>
`
	w.Write([]byte(html))
}

func NewHandler(cs video.CameraStreamer) http.Handler {
	h := NewRegexHandler()
	h.Handle(regexp.MustCompile("cameras/"), NewCameraHandler(cs))
	h.Handle(regexp.MustCompile("dash"), NewDashHandler())

	return h
}