package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// PrebuildConfiguration lists a combination of parameter
// in order to cache stages that are widely used
type PrebuildConfiguration struct {
	BuilderSource string                       `yaml:"builderSource"`
	BuilderCache  string                       `yaml:"builderCache"`
	BasePreBuild  []extendedBuildConfiguration `yaml:"images"`
}

type extendedBuildConfiguration struct {
	BuilderName string `yaml:"builderName"`
	Spec        map[string]interface{}
	Stages      []string `yaml:"stages"`
}

// ReadPrebuildConfiguration unserialize the build configuration
func ReadPrebuildConfiguration(filepath string) (PrebuildConfiguration, error) {
	conf := PrebuildConfiguration{}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}
