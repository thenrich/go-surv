package http

import "net/http"

func NewDashHandler() *DashHandler {
	return &DashHandler{}
}

type DashHandler struct{}

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