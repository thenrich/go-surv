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
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/thenrich/go-surv/cloud"
	"github.com/thenrich/go-surv/config"
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

	conf := flag.String("conf", "", "config file")
	flag.Parse()

	cfg, err := config.ParseConfig(*conf)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.AWS.AccessKey != "" && cfg.AWS.SecretAccessKey != "" {
		creds := credentials.NewStaticCredentials(cfg.AWS.AccessKey, cfg.AWS.SecretAccessKey, "")
		awsCfg := aws.NewConfig().WithCredentials(creds).WithRegion(cfg.AWS.Region)
		cloud.ConfigureS3(awsCfg)
	}

	ch := video.NewCameraHandler(cfg)
	for _, cam := range cfg.Cameras {
		ch.AddCamera(video.NewCamera(cam.Name, cam.Source, cfg.StorageInterval))
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
