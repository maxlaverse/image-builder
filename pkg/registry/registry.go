package registry

import (
	"net/url"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	log "github.com/sirupsen/logrus"
)

func ImageExists(imagename string) bool {
	if !isRemoteImage(imagename) {
		return false
	}
	log.Infof("Checking existence of '%s'", imagename)
	u, err := url.Parse("https://" + imagename)
	if err != nil {
		log.Errorf("Registry error: %v", err)
		return false
	}
	part := strings.Split(u.EscapedPath(), ":")
	url := "https://" + u.Hostname()
	username := "" // anonymous
	password := "" // anonymous
	hub, err := registry.New(url, username, password)
	if err != nil {
		log.Errorf("Registry error: %v", err)
		return false
	}

	hub.Logf = func(format string, args ...interface{}) {
		log.Debugf(format, args)
	}

	tags, err := hub.Tags(strings.TrimPrefix(part[0], "/"))
	if err != nil {
		log.Errorf("Registry error: %v", err)
		return false
	}
	for _, t := range tags {
		if t == part[1] {
			return true
		}
	}
	return false
}

func isRemoteImage(imagename string) bool {
	return strings.Count(imagename, "/") != 0
}
