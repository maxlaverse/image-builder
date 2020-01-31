package registry

import (
	"fmt"

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
