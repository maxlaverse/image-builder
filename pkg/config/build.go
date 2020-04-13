package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/utils"
	"gopkg.in/yaml.v2"
)

// BuildConfiguration represents the build configuration of an application
type BuildConfiguration struct {
	data map[string]interface{}
	path string
}

// ReadBuildConfiguration unserialize the build configuration
func ReadBuildConfiguration(filepath string) (BuildConfiguration, error) {
	conf := BuildConfiguration{}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal([]byte(data), &conf.data)
	if err != nil {
		return conf, err
	}
	conf.path = filepath
	return conf, nil
}

// BuilderName returns the builder's name
func (c *BuildConfiguration) BuilderName() string {
	return utils.KeyValueOrEmpty(c.data, "builderName")
}

// BuilderLocation returns the builder's location
func (c *BuildConfiguration) BuilderLocation() string {
	return strings.TrimSuffix(utils.KeyValueOrEmpty(c.data, "builderLocation"), "/")
}

// BuilderCache returns the builder's image cache
func (c *BuildConfiguration) BuilderCache() string {
	return utils.KeyValueOrEmpty(c.data, "builderCache")
}

// IsBuilderCacheSet returns wether a buildercache is specified
func (c *BuildConfiguration) IsBuilderCacheSet() bool {
	return len(c.BuilderCache()) > 0
}

// IncludePatterns returns the files to include in the Docker context
func (c *BuildConfiguration) IncludePatterns(stageName string) []string {
	return c.MergedStringSpecAttribute(stageName, "contextInclude")
}

// SpecAttribute returns the Dockerfiles to ignore
func (c *BuildConfiguration) SpecAttribute(stageName, attrName string) (interface{}, bool) {
	if v, ok := c.data[fmt.Sprintf("%sSpec", stageName)]; ok {
		if v, ok := v.(map[interface{}]interface{})[attrName]; ok {
			return v, true
		}
	}

	if v, ok := c.data["globalSpec"].(map[interface{}]interface{})[attrName]; ok {
		return v, true
	}
	return nil, false
}

// MergedStringSpecAttribute returns the Dockerfiles to ignore
func (c *BuildConfiguration) MergedStringSpecAttribute(stageName, attrName string) []string {
	result := []string{}
	if v, ok := c.data["globalSpec"].(map[interface{}]interface{})[attrName]; ok {
		for _, v3 := range v.([]interface{}) {
			result = append(result, v3.(string))
		}
	}
	if v, ok := c.data[fmt.Sprintf("%sSpec", stageName)]; ok {
		if v2, ok := v.(map[interface{}]interface{})[attrName]; ok {
			for _, v3 := range v2.([]interface{}) {
				result = append(result, v3.(string))
			}
		}
	}
	return result
}

// SpecAttributeNames returns the Dockerfiles to ignore
func (c *BuildConfiguration) SpecAttributeNames(stageName string) []string {
	keyList := []string{}
	if v, ok := c.data[fmt.Sprintf("%sSpec", stageName)]; ok {
		keyList = utils.MapKeys(v.(map[string]interface{}))
	}
	if v, ok := c.data["globalSpec"]; ok {
		keyList = append(keyList, utils.MapKeys(v.(map[string]interface{}))...)
	}
	return keyList
}
