package main

import (
	"flag"
	"github.com/thenrich/go-surv/config"
	"log"
	"os"
	"os/signal"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/thenrich/go-surv/cloud"
	"github.com/thenrich/go-surv/video"
	ghttp "github.com/thenrich/go-surv/http"
	"net/http"
)

func main() {
	conf := flag.String("conf", "", "config file")
	flag.Parse()

	if *conf == "" {
		flag.Usage()
		os.Exit(1)
	}

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

	http.ListenAndServe(":8080", ghttp.NewHttpHandler(ch))
}
