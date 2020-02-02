package config

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// BuildConfiguration represents the build configuration of an application
type BuildConfiguration struct {
	BuilderCache  string                            `yaml:"builderCache"`
	BuilderName   string                            `yaml:"builderName"`
	BuilderSource string                            `yaml:"builderSource"`
	ImageSpec     map[string]interface{}            `yaml:"imageSpec"`
	StageSpec     map[string]map[string]interface{} `yaml:"stageSpec"`
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

// DockerignoreForStage returns the Dockerfiles to ignore
func (c *BuildConfiguration) DockerignoreForStage(stageName string) []string {
	result := []string{}
	if v, ok := c.ImageSpec["dockerIgnores"]; ok {
		for _, v3 := range v.([]interface{}) {
			result = append(result, v3.(string))
		}
	}
	if v, ok := c.StageSpec[stageName]; ok {
		if v2, ok := v["dockerIgnores"]; ok {
			for _, v3 := range v2.([]interface{}) {
				result = append(result, v3.(string))
			}
		}
	}
	return result
}

func (c *BuildConfiguration) normalize() {
	if strings.HasSuffix(c.BuilderCache, "/") {
		c.BuilderCache = strings.TrimSuffix(c.BuilderCache, "/")
	}
}
