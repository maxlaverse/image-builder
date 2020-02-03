package registry

import (
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func ImageWithDigest(ref string) (string, error) {
	desc, err := getManifest(ref)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s@%s", desc.Ref.Context(), desc.Digest.String()), nil
}

func ImageAge(ref string) (time.Duration, error) {
	desc, err := getManifest(ref)
	if err != nil {
		return time.Duration(0), err
	}
	img, err := desc.Image()
	if err != nil {
		return time.Duration(0), err
	}
	cfg, err := img.ConfigFile()
	if err != nil {
		return time.Duration(0), err
	}
	time.Now().Sub(cfg.Created.Time)
	return time.Now().Sub(cfg.Created.Time), nil
}

func TagImage(source, dest string) error {
	ref, err := name.ParseReference(source)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %v", source, err)
	}
	desc, err := remote.Get(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return fmt.Errorf("fetching %q: %v", source, err)
	}

	dst := ref.Context().Tag(dest)

	return remote.Tag(dst, desc, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}

func ImageExists(ref string) (bool, error) {
	_, err := getManifest(ref)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func getManifest(r string) (*remote.Descriptor, error) {
	ref, err := name.ParseReference(r)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %v", r, err)
	}
	return remote.Get(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}
