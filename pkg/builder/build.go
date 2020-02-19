package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/maxlaverse/image-builder/pkg/engine"
	"github.com/maxlaverse/image-builder/pkg/executor"
	"github.com/maxlaverse/image-builder/pkg/fileutils"
	"github.com/maxlaverse/image-builder/pkg/registry"
	"github.com/maxlaverse/image-builder/pkg/template"
	log "github.com/sirupsen/logrus"
)

// Build transform BuildConfigurations into Docker images
type Build struct {
	engine         engine.BuildEngine
	buildDef       *config.BuilderDef
	exec           executor.Executor
	images         map[string]string
	dryRun         bool
	localContext   string
	cacheImagePull bool
	cacheImagePush bool
	targetImage    string
	buildConf      config.BuildConfiguration
}

// BuildStatus represent the status of an image built
type BuildStatus struct {
	ImageURL string
	Built    bool
}

// NewBuild returns a new instance of Build
func NewBuild(e engine.BuildEngine, exec executor.Executor, buildConf config.BuildConfiguration, r *config.BuilderDef, dryRun, cacheImagePull, cacheImagePush bool, targetImage string, localContext string) *Build {
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
		exec:           exec,
	}
}

// GetStageBuildOrder returns in which order the stages should be build
func (b *Build) GetStageBuildOrder(stages []string) ([]string, error) {
	stageGraph := NewGraph()
	for _, stage := range b.buildDef.GetStages() {
		builderPath := b.buildDef.GetStagePath(stage)
		data := template.NewMinimalBuildData(b.buildConf, b.localContext)
		dockerfile, err := template.RenderDockerfile(path.Join(builderPath, "Dockerfile"), data, b.exec)
		if err != nil {
			log.Errorf("Could not render the Dockerfile: %v", err)
			return nil, err
		}

		stageGraph.AddNode(stage, dockerfile.GetRequiredStages()...)
	}

	return stageGraph.ResolveUpTo(stages)
}

// BuildStage pulls or build an image and push it
func (b *Build) BuildStage(stage string) (BuildStatus, error) {
	log.Debugf("Building stage '%s'", stage)
	buildStatus, err := b.pullOrBuildStage(b.targetImage, stage)
	if err != nil {
		log.Debugf("Unable to pull or build stage '%s': %v", stage, err)
		return buildStatus, err
	}
	b.images[stage] = buildStatus.ImageURL
	return buildStatus, nil
}

// pullOrBuildStage lookup the builder cache for an already existing image for the given content and build it otherwise
func (b *Build) pullOrBuildStage(dockerImage, stage string) (BuildStatus, error) {
	builderPath := b.buildDef.GetStagePath(stage)
	data := template.NewBuildData(b.buildConf, b.localContext, b.images)
	dockerfile, err := template.RenderDockerfile(path.Join(builderPath, "Dockerfile"), data, b.exec)
	if err != nil {
		return BuildStatus{}, err
	}

	if b.dryRun {
		log.Infof("Generated Dockerfile:\n%s", dockerfile.GetContent())
	}

	dockerfilePath, err := writeDockerFile(dockerfile.GetContent())
	if err != nil {
		return BuildStatus{}, err
	}
	defer os.Remove(dockerfilePath)

	var buildContext string
	if dockerfile.UseBuilderContext() {
		buildContext = builderPath
	} else {
		buildContext = b.localContext
	}

	dockerignorePath := path.Join(buildContext, ".dockerignore")
	err = writeDockerIgnore(dockerignorePath, b.getFilesToIgnore(dockerfile, stage))
	if err != nil {
		return BuildStatus{}, err
	}
	defer os.Remove(dockerignorePath)

	tag, err := b.computeImageTag(dockerfile, buildContext, stage)
	if err != nil {
		log.Errorf("Error while computing the tag for the image: %v", err)
		return BuildStatus{}, err
	}

	var dockerImageWithTag string
	if b.cacheImagePull && b.buildConf.IsBuilderCacheSet() {
		// Try to pull a pre build image
		dockerImageWithTag = b.buildConf.Builder.ImageCache + "/" + b.buildConf.Builder.Name + ":" + stage + "-" + tag
		exists, err := registry.ImageExists(dockerImageWithTag)
		if err != nil {
			log.Errorf("Error while computing the tag for the image: %v", err)
			return BuildStatus{}, err
		}

		if exists {
			log.Infof("An image for '%s' was found in the builder's cache", dockerImageWithTag)
			return BuildStatus{ImageURL: dockerImageWithTag, Built: false}, nil
		}
		log.Debugf("No image found for '%s' in the Builder's cache", dockerImageWithTag)
	}

	// Try to pull a cached image
	//TODO: Normalize Docker URL
	dockerImageWithTag = dockerImage + ":" + stage + "-" + tag
	if b.cacheImagePull {
		exists, err := registry.ImageExists(dockerImageWithTag)
		if err != nil {
			log.Errorf("Error while computing the tag for the image: %v", err)
			return BuildStatus{}, err
		}
		if exists {
			log.Debugf("An image for '%s' was found in the application's cache", dockerImageWithTag)
			return BuildStatus{ImageURL: dockerImageWithTag, Built: false}, nil
		}
	}

	if b.dryRun {
		log.Infof("Skipping build of '%s' because of dry run", dockerImageWithTag)
		return BuildStatus{ImageURL: dockerImageWithTag, Built: false}, nil
	}

	log.Infof("A new image needs to be built for '%s'", dockerImageWithTag)
	err = b.engine.Build(dockerfilePath, dockerImageWithTag, buildContext)
	if err != nil {
		log.Errorf("The image build failed for '%s'", dockerImageWithTag)
		return BuildStatus{}, err
	}

	if b.cacheImagePush {
		log.Infof("Pushing image '%s'", dockerImageWithTag)
		err := b.engine.Push(dockerImageWithTag)
		if err != nil {
			return BuildStatus{}, err
		}
		for _, t := range dockerfile.GetTagAliases() {
			log.Infof("Tagging image '%s' as '%s'", dockerImageWithTag, t)
			err := registry.TagImage(dockerImageWithTag, t)
			if err != nil {
				return BuildStatus{}, err
			}
		}
	}

	return BuildStatus{ImageURL: dockerImageWithTag, Built: true}, nil
}

func (b *Build) getFilesToIgnore(dockerfile template.Dockerfile, stage string) []string {
	res := append(dockerfile.GetDockerIgnores(), b.buildConf.DockerignoreForStage(stage)...)
	res = append(res, ".dockerignore", "build.yaml")
	return res
}

func (b *Build) computeImageTag(dockerfile template.Dockerfile, context string, stage string) (string, error) {
	hash, err := fileutils.ContentHashing(dockerfile.GetContentWithoutIgnoredLines(), context, b.getFilesToIgnore(dockerfile, stage))
	if err != nil {
		log.Errorf("Error while computing content hashing")
		return "", err
	}
	if len(dockerfile.GetFriendlyTag()) > 0 {
		return fmt.Sprintf("%s-%s", dockerfile.GetFriendlyTag(), hash), nil
	}
	return hash, nil
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
