package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/pkg/errors"
)

type Config struct {
	Cameras []struct{
		Name string `yaml:"name"`
		Source string `yaml:"source"`
	} `yaml:"cameras"`
}

func parseConfig(fn string) (*Config, error) {
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
