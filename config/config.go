package config

import (
	"time"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/pkg/errors"
)

type Config struct {
	// Storage medium for recorded video
	Storage string `yaml:"storage"`

	// Length of recorded segments for storage
	StorageInterval time.Duration `yaml:"storageInterval"`

	// AWS configuration
	AWS AWSConfig `yaml:"aws"`

	// Camera configuration
	Cameras []struct {
		Name   string `yaml:"name"`
		Source string `yaml:"source"`
	} `yaml:"cameras"`
}

type AWSConfig struct {
	// S3 bucket for storage
	S3Bucket string `yaml:"s3bucket"`

	// AWS Region
	Region string `yaml:"region"`

	// AWS creds
	AccessKey       string `yaml:"accessKey"`
	SecretAccessKey string `yaml:"secretAccessKey"`
}

func (a *AWSConfig) Ready() bool {
	if a.S3Bucket == "" {
		return false
	}

	if a.Region == "" {
		return false
	}

	if a.AccessKey == "" {
		return false
	}

	if a.SecretAccessKey == "" {
		return false
	}

	return true
}

func ParseConfig(fn string) (*Config, error) {
	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading %s", fn)
	}

	var cc Config

	err = yaml.Unmarshal(bytes, &cc)
	if err != nil {
		return nil, errors.Wrap(err, "error deserializing configuration")
	}

	return &cc, nil

}
