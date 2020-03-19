package config

import (
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type CliConfiguration struct {
	DefaultLocation          string `yaml:"default-builder-location"`
	DefaultBuildConcurrency  int64  `yaml:"default-build-concurrency"`
	DefaultPullConcurrency   int64  `yaml:"default-pull-concurrency"`
	DefaultBuilderImageCache string `yaml:"default-builder-image-cache"`
	DefaultCacheImagePush    bool   `yaml:"default-cache-image-push"`
	DefaultCacheImagePull    bool   `yaml:"default-cache-image-pull"`
	DefaultEngine            string `yaml:"default-engine"`
	filepath                 string
}

func (c *CliConfiguration) Load(filepath string) error {
	c.filepath = filepath
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return err
	}
	return nil
}

func (c *CliConfiguration) Save() error {
	bytes, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(c.filepath), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.filepath, bytes, 0700)
	if err != nil {
		return err
	}
	return nil
}

func NewDefaultConfiguration() *CliConfiguration {
	return &CliConfiguration{
		DefaultCacheImagePull:   true,
		DefaultCacheImagePush:   true,
		DefaultEngine:           "docker",
		DefaultBuildConcurrency: 1,
		DefaultPullConcurrency:  1,
	}
}
