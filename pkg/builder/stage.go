package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/engine"
	"github.com/maxlaverse/image-builder/pkg/fileutils"
	"github.com/maxlaverse/image-builder/pkg/template"
	log "github.com/sirupsen/logrus"
)

const (
	dockerIgnoreName = ".dockerignore"
)

// StageImageStatus represents the status of the image corresponding to a
// stage (e.g. absent, cached, pulled)
type StageImageStatus string

const (
	// Initialized is unknown
	Initialized StageImageStatus = "initialize"

	// ImageAbsent is for images not present locally not remotely
	ImageAbsent StageImageStatus = "absent"

	// ImageCached is for images that are found in the Application or Builder's cache
	ImageCached StageImageStatus = "present-in-cache"

	// ImagePulled is for images that were absent but could be pulled
	ImagePulled StageImageStatus = "pulled"

	// ImageBuilt is for images that have been built
	ImageBuilt StageImageStatus = "built"
)

// BuildStage represents a individual stage which can be built
type BuildStage interface {
	Build(engineBuild engine.BuildEngine) error
	ComputeContentHash() error
	ContentHash() string
	Dockerfile() string
	Dockerignore() []string
	GetRequiredStages() []string
	GetTagAliases() []string
	ImageTag() (string, error)
	ImageURL() string
	Name() string
	Render() error
	SetImageURL(source string)
	SetSourceImageURL(source string)
	SetStatus(status StageImageStatus)
	SourceImageURL() string
	Status() StageImageStatus
}

// buildStage represents a individual stage which can be built
type buildStage struct {
	extraIgnorePatterns []string
	contentHash         string
	dockerfile          template.Dockerfile
	imageURL            string
	name                string
	sourceImageURL      string
	status              StageImageStatus
}

// NewBuildStage returns a individual stage
func NewBuildStage(name string, dockerfile template.Dockerfile, extraIgnorePatterns []string) BuildStage {
	return &buildStage{
		extraIgnorePatterns: append(extraIgnorePatterns, dockerIgnoreName),
		dockerfile:          dockerfile,
		name:                name,
		status:              Initialized,
	}
}

func (b *buildStage) SetSourceImageURL(source string) {
	b.sourceImageURL = source
}

func (b *buildStage) SetImageURL(source string) {
	b.imageURL = source
}

func (b *buildStage) SetStatus(status StageImageStatus) {
	b.status = status
}

// Build writes a Dockerfile and .dockerignore and calls the engine's build command
func (b *buildStage) Build(engineBuild engine.BuildEngine) error {
	log.Infof("Build context for '%s' is '%s'", b.Name(), b.dockerfile.GetBuildContext())
	dockerfilePath, err := writeDockerfile(b.dockerfile.GetContent())
	if err != nil {
		return fmt.Errorf("error writing 'Dockerfile' file: %w", err)
	}
	defer os.Remove(dockerfilePath)

	dockerignorePath := path.Join(b.dockerfile.GetBuildContext(), dockerIgnoreName)
	err = writeDockerIgnore(dockerignorePath, b.Dockerignore())
	if err != nil {
		return fmt.Errorf("error writing '.dockerignore' file: %w", err)
	}
	defer os.Remove(dockerignorePath)

	return engineBuild.Build(dockerfilePath, b.imageURL, b.dockerfile.GetBuildContext())
}

func (b *buildStage) ComputeContentHash() error {
	log.Tracef("Context directory is '%s'", b.dockerfile.GetBuildContext())
	contentHash, err := fileutils.ContentHashing(b.dockerfile.GetContentWithoutIgnoredLines(), b.dockerfile.GetBuildContext(), b.Dockerignore())
	if err != nil {
		return fmt.Errorf("error computing ContentHash: %w", err)
	}
	b.contentHash = contentHash
	return nil
}

func (b *buildStage) ContentHash() string {
	if len(b.contentHash) == 0 {
		b.ComputeContentHash()
	}
	return b.contentHash
}

func (b *buildStage) Dockerfile() string {
	return b.dockerfile.GetContent()
}

func (b *buildStage) Dockerignore() []string {
	return append(b.extraIgnorePatterns, b.dockerfile.GetDockerIgnores()...)
}

func (b *buildStage) GetRequiredStages() []string {
	return b.dockerfile.GetRequiredStages()
}

func (b *buildStage) GetTagAliases() []string {
	return b.dockerfile.GetTagAliases()
}

func (b *buildStage) ImageURL() string {
	return b.imageURL
}

func (b *buildStage) Render() error {
	return b.dockerfile.Render()
}

func (b *buildStage) ImageTag() (string, error) {
	if len(b.contentHash) == 0 {
		b.ComputeContentHash()
	}
	if len(b.dockerfile.GetFriendlyTag()) > 0 {
		return fmt.Sprintf("%s-%s", b.dockerfile.GetFriendlyTag(), b.contentHash), nil
	}
	return b.contentHash, nil
}

func (b *buildStage) Name() string {
	return b.name
}

func (b *buildStage) Status() StageImageStatus {
	return b.status
}

func (b *buildStage) SourceImageURL() string {
	return b.sourceImageURL
}

func writeDockerfile(content string) (string, error) {
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return f.Name(), err
}

func writeDockerIgnore(path string, ignorePatterns []string) error {
	content := strings.Join(ignorePatterns, "\n") + "\n"
	return ioutil.WriteFile(path, []byte(content), 0644)
}
