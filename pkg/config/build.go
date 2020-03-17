package config

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// BuildConfiguration represents the build configuration of an application
type BuildConfiguration struct {
	Builder    BuildBuilderConfiguration         `yaml:"builder"`
	GlobalSpec map[string]interface{}            `yaml:"globalSpec"`
	StageSpec  map[string]map[string]interface{} `yaml:"stageSpec"`
	path       string
}

// BuildBuilderConfiguration represents the settings for the Builder to use
type BuildBuilderConfiguration struct {
	Name       string `yaml:"name"`
	Location   string `yaml:"location"`
	ImageCache string `yaml:"imageCache"`
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
	conf.path = filepath
	conf.normalize()
	return conf, nil
}

// IsBuilderCacheSet returns wether a buildercache is specified
func (c *BuildConfiguration) IsBuilderCacheSet() bool {
	return len(c.Builder.ImageCache) > 0
}

// IgnorePatterns returns the Dockerfiles to ignore
func (c *BuildConfiguration) IgnorePatterns(stageName string) []string {
	result := []string{}
	if v, ok := c.GlobalSpec["dockerIgnores"]; ok {
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
	result = append(result, c.path)
	return result
}

// GetSpecAttribute returns the Dockerfiles to ignore
func (c *BuildConfiguration) GetSpecAttribute(stageName, attrName string) (interface{}, bool) {
	if v, ok := c.StageSpec[stageName][attrName]; ok {
		return v, true
	}
	if v, ok := c.GlobalSpec[attrName]; ok {
		return v, true
	}
	return nil, false
}

func (c *BuildConfiguration) normalize() {
	if strings.HasSuffix(c.Builder.ImageCache, "/") {
		c.Builder.ImageCache = strings.TrimSuffix(c.Builder.ImageCache, "/")
	}
}
