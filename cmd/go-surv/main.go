package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/thenrich/go-surv/cloud"
	"github.com/thenrich/go-surv/config"
	ghttp "github.com/thenrich/go-surv/http"
	"github.com/thenrich/go-surv/video"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	for _, cfgCam := range cfg.Cameras {
		camera := video.NewCamera(cfgCam.Name, cfgCam.Source, cfg.StorageInterval)
		if cfg.AWS.Ready() && cfg.Storage == "s3" {
			//s3storage := gaws.NewS3Storage(cfg.AWS, cfg.AWS.S3Bucket)
			//camera.AddWriter(video.NewCloudStorage(camera.Name, cfg.StorageInterval, cfg, s3storage))
		}
		ch.AddCamera(camera)
	}

	log.Println("Start streams")
	go ch.StartStreams()

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

	http.ListenAndServe(":8080", ghttp.NewHandler(ch))
}
