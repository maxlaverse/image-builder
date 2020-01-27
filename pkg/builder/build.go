package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/maxlaverse/image-builder/pkg/engine"
	"github.com/maxlaverse/image-builder/pkg/fileutils"
	"github.com/maxlaverse/image-builder/pkg/registry"
	"github.com/maxlaverse/image-builder/pkg/template"
	log "github.com/sirupsen/logrus"
)

// Build transform BuildConfigurations into Docker images
type Build struct {
	engine         engine.BuildEngine
	buildDef       *config.BuilderDef
	images         map[string]string
	dryRun         bool
	localContext   string
	cacheImagePull bool
	cacheImagePush bool
	targetImage    string
	buildConf      config.BuildConfiguration
}

// NewBuild returns a new instance of Build
func NewBuild(e engine.BuildEngine, buildConf config.BuildConfiguration, r *config.BuilderDef, dryRun, cacheImagePull, cacheImagePush bool, targetImage string, localContext string) *Build {
	return &Build{
		engine:         e,
		buildDef:       r,
		images:         map[string]string{},
		dryRun:         dryRun,
		localContext:   localContext,
		cacheImagePush: cacheImagePush,
		cacheImagePull: cacheImagePull,
		targetImage:    targetImage,
		buildConf:      buildConf,
	}
}

// GetStageBuildOrder returns in which order the stages should be build
func (b *Build) GetStageBuildOrder(finalStage string) ([]string, error) {
	stageGraph := NewGraph()
	for _, stage := range b.buildDef.GetStages() {
		builderPath := b.buildDef.GetStagePath(stage)
		tmplContext := template.NewMinimalTemplateContext(b.buildConf)
		dockerfile, err := template.RenderDockerfile(path.Join(builderPath, "Dockerfile"), tmplContext)
		if err != nil {
			log.Errorf("Could not render the Dockerfile: %v", err)
			return nil, err
		}

		stageGraph.AddNode(stage, dockerfile.GetRequiredStages()...)
	}

	return stageGraph.ResolveUpTo(finalStage)
}

// BuildStage pulls or build an image and push it
func (b *Build) BuildStage(stage string) error {
	log.Debugf("Building stage '%s'", stage)
	dockerImageWithTag, err := b.pullOrBuildStage(b.targetImage, stage)
	if err != nil {
		log.Debugf("Unable to pull or build stage '%s': %v", stage, err)
		return err
	}
	if len(dockerImageWithTag) > 0 {
		b.images[stage] = dockerImageWithTag
	}
	return nil
}

// pullOrBuildStage lookup the builder cache for an already existing image for the given content and build it otherwise
func (b *Build) pullOrBuildStage(dockerImage, stage string) (string, error) {
	builderPath := b.buildDef.GetStagePath(stage)
	tmplContext := template.NewTemplateContext(b.buildConf, b.images)
	dockerfile, err := template.RenderDockerfile(path.Join(builderPath, "Dockerfile"), tmplContext)
	if err != nil {
		return "", err
	}

	if b.dryRun {
		log.Infof("Generated Dockerfile:\n%s", dockerfile.GetContent())
	}

	dockerfilePath, err := writeDockerFile(dockerfile.GetContent())
	if err != nil {
		return "", err
	}
	defer os.Remove(dockerfilePath)

	var buildContext string
	if dockerfile.UseBuilderContext() {
		buildContext = builderPath
	} else {
		buildContext = b.localContext
	}

	dockerignorePath := path.Join(buildContext, ".dockerignore")
	err = writeDockerIgnore(dockerignorePath, b.getFilesToIgnore(dockerfile))
	if err != nil {
		return "", err
	}
	defer os.Remove(dockerignorePath)

	tag, err := b.computeImageTag(dockerfile, buildContext)
	if err != nil {
		log.Errorf("Error while computing the tag for the image: %v", err)
		return "", err
	}

	var dockerImageWithTag string
	if b.cacheImagePull && b.buildConf.HasBuilderCache() {
		// Try to pull a pre build image
		dockerImageWithTag = b.buildConf.BuilderCache + "/" + b.buildConf.BuilderName + ":" + stage + "-" + tag
		if registry.ImageExists(dockerImageWithTag) {
			log.Infof("An image for '%s' was found in the builder's cache", dockerImageWithTag)
			return dockerImageWithTag, nil
		}
		log.Debugf("No image found for '%s' in the Builder's cache", dockerImageWithTag)
	}

	// Try to pull a cached image
	//TODO: Normalize Docker URL
	dockerImageWithTag = dockerImage + ":" + stage + "-" + tag
	if b.cacheImagePull && registry.ImageExists(dockerImageWithTag) {
		log.Debugf("An image for '%s' was found in the application's cache", dockerImageWithTag)
		return dockerImageWithTag, nil
	}

	if b.dryRun {
		log.Infof("Skipping build of '%s' because of dry run", dockerImageWithTag)
		return dockerImageWithTag, nil
	}

	log.Infof("A new image needs to be built for '%s'", dockerImageWithTag)
	err = b.engine.Build(dockerfilePath, dockerImageWithTag, buildContext)
	if err != nil {
		log.Errorf("The image build failed for '%s'", dockerImageWithTag)
		return dockerImageWithTag, err
	}

	if b.cacheImagePush {
		log.Infof("Pushing image '%s'", dockerImageWithTag)
		err := b.engine.Push(dockerImageWithTag)
		if err != nil {
			return dockerImageWithTag, err
		}
	}

	return dockerImageWithTag, nil
}

func (b *Build) getFilesToIgnore(dockerfile *template.Dockerfile) []string {
	return append(dockerfile.GetDockerIgnores(), ".dockerignore", "build.yaml", ".image-builder-info")
}

func (b *Build) computeImageTag(dockerfile *template.Dockerfile, context string) (string, error) {
	hash, err := fileutils.ContentHashing(dockerfile.GetFilteredContent(), context, b.getFilesToIgnore(dockerfile))
	if err != nil {
		log.Errorf("Error while computing content hashing")
		return "", err
	}
	if len(dockerfile.GetFriendlyTag()) > 0 {
		return fmt.Sprintf("%s-%s", dockerfile.GetFriendlyTag(), hash), nil
	}
	return hash, nil
}

// Images returns the images the builder had to pull or build
func (b *Build) Images() map[string]string {
	return b.images
}

func writeDockerFile(dockerfile string) (string, error) {
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.WriteString(dockerfile)
	return f.Name(), err
}

func writeDockerIgnore(dockerfilePath string, filesToIgnore []string) error {
	dockerIgnoreContent := strings.Join(filesToIgnore, "\n") + "\n"
	return ioutil.WriteFile(dockerfilePath, []byte(dockerIgnoreContent), 0644)
}
