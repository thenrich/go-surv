package http

import (
	"net/http"
	"regexp"
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