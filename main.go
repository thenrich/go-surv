package main

import (
	"github.com/nareix/joy4/format"
	"flag"
	_ "image/jpeg"
	"net/http"
	"github.com/thenrich/go-surv/video"
	"log"
	"os"
	"os/signal"
)

func init() {
	format.RegisterAll()
}

//type Config struct {
//	// prefix for a directory of still images every second
//	prefix string
//
//	// single file updated every second
//	snapfile string
//}

func main() {

	config := flag.String("conf", "", "config file")
	//dstfile := flag.String("dst", "output.mp4", "Output file")
	//prefix := flag.String("snapdir", "stills", "Output directory for snapshots")
	//snapfile := flag.String("snapfile", "out.jpg", "Single output snapshot")
	//max := flag.Int("max", 5, "Max seconds")
	flag.Parse()

	//cfg := &Config{prefix: *prefix, snapfile: *snapfile}

	cfg, err := parseConfig(*config)
	if err != nil {
		log.Fatal(err)
	}

	ch := video.NewCameraHandler()
	for _, cam := range cfg.Cameras {
		ch.AddCamera(&video.Camera{
			Name:      cam.Name,
			SourceURL: cam.Source,
		})
	}

	log.Println("Start streams")
	ch.StartStreams()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				ch.CloseStreams()
				log.Fatal("Terminating...")
			}
		}
	}()

	http.ListenAndServe(":8080", NewHttpHandler(ch))

}
