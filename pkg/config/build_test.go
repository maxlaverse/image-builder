package config

import (
	"testing"
)

func TestReadingEmptyConfigurationDoesntCrash(t *testing.T) {
	conf := BuildConfiguration{
		path: "somewhere",
		data: map[string]interface{}{},
	}

	conf.BuilderCache()
	conf.BuilderLocation()
	conf.BuilderName()
	conf.IncludePatterns("")
	conf.IsBuilderCacheSet()
	conf.SpecAttribute("", "")
	conf.SpecAttributeNames("")
}
