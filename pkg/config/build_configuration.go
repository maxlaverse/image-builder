package config

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// BuildConfiguration represents the build configuration of an application
type BuildConfiguration struct {
	BuilderCache  string                 `yaml:"builderCache"`
	BuilderName   string                 `yaml:"builderName"`
	BuilderSource string                 `yaml:"builderSource"`
	ImageSpec     map[string]interface{} `yaml:"imageSpec"`
}

// ReadBuildConfiguration unserialize the build configuration
func ReadBuildConfiguration(filepath string) (BuildConfiguration, error) {
	conf := BuildConfiguration{}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		return conf, err
	}
	conf.normalize()
	return conf, nil
}

// IsBuilderCacheSet returns wether a buildercache is specified
func (c *BuildConfiguration) IsBuilderCacheSet() bool {
	return true
}

func (c *BuildConfiguration) normalize() {
	if strings.HasSuffix(c.BuilderCache, "/") {
		c.BuilderCache = strings.TrimSuffix(c.BuilderCache, "/")
	}
}
