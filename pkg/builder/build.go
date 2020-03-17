package builder

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/BurntSushi/locker"
	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/maxlaverse/image-builder/pkg/engine"
	"github.com/maxlaverse/image-builder/pkg/executor"
	"github.com/maxlaverse/image-builder/pkg/registry"
	"github.com/maxlaverse/image-builder/pkg/template"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// BuildOptions holds the options for a build
type BuildOptions struct {
	// Check for already build images
	CacheImagePull bool

	// Push built stages for reuse
	CacheImagePush bool

	// DryRun disables any actual image build
	DryRun bool

	// BuildConcurrency indicates how many concurrent image build are allowed
	BuildConcurrency int64
}

// Build transform BuildConfigurations into Docker images
type Build struct {
	buildConf    config.BuildConfiguration
	buildDef     Definition
	buildStages  sync.Map
	engine       engine.BuildEngine
	exec         executor.Executor
	localContext string
	opts         BuildOptions
	targetImage  string
	locker       *locker.Locker
	sem          *semaphore.Weighted
}

// NewBuild returns a new instance of Build
func NewBuild(e engine.BuildEngine, exec executor.Executor, buildDef Definition, buildConf config.BuildConfiguration, opts BuildOptions, targetImage string, localContext string) *Build {
	if opts.BuildConcurrency < 1 {
		opts.BuildConcurrency = 1
	}
	return &Build{
		buildConf:    buildConf,
		buildDef:     buildDef,
		buildStages:  sync.Map{},
		engine:       e,
		exec:         exec,
		localContext: localContext,
		opts:         opts,
		targetImage:  targetImage,
		locker:       locker.NewLocker(),
		sem:          semaphore.NewWeighted(opts.BuildConcurrency),
	}
}

// PrepareStages builds a set of stages
func (b *Build) PrepareStages(stageNames []string) ([]BuildStage, error) {
	log.Infof("Rendering Dockerfiles")
	for _, stageName := range stageNames {
		stage, err := b.prepareStage(stageName)
		if err != nil {
			return nil, err
		}
		b.buildStages.Store(stageName, stage)
	}

	var err error
	b.buildStages.Range(func(stageName, stage interface{}) bool {
		log.Debugf("Final rendering of template '%s' to resolve stage references", stageName)
		if err = stage.(BuildStage).Render(); err != nil {
			return false
		}
		log.Debugf(".dockerignore for stage '%s' is:\n%s", stageName, strings.Join(stage.(BuildStage).Dockerignore(), "\n"))
		log.Debugf("Dockerfile for stage '%s' is:\n%s", stageName, stage.(BuildStage).Dockerfile())
		return true
	})
	return b.getBuildStages(), err
}

// BuildStages builds a set of stages
func (b *Build) BuildStages(stageNames []string) ([]BuildStage, error) {
	_, err := b.PrepareStages(stageNames)
	if err != nil {
		return nil, fmt.Errorf("error while preparing some stages: %w", err)
	}

	if b.opts.DryRun {
		return b.getBuildStages(), nil
	}

	log.Infof("Starting build")
	g, _ := errgroup.WithContext(context.Background())
	for _, stageName := range stageNames {
		stage, ok := b.buildStages.Load(stageName)
		if !ok {
			return nil, fmt.Errorf("stage '%s' was not prepared", stageName)
		}

		g.Go(func() error { return b.buildStage(stage.(BuildStage)) })
	}

	if err := g.Wait(); err != nil {
		return b.getBuildStages(), err
	}

	return b.getBuildStages(), nil
}

// templateStageResolver is called by the renderer to replace a stage
// reference with its imageURL. It's used to recursively prepare stages
func (b *Build) templateStageResolver(stageName string) (string, error) {
	stage, err := b.prepareStage(stageName)
	if err != nil {
		return "error-while-resoving-stage", err
	}
	b.buildStages.Store(stageName, stage)
	return stage.ImageURL(), nil
}

// prepareStage renders all the required Dockerfiles and verifies image some
// stages can be pulled from remote registries
func (b *Build) prepareStage(stageName string) (BuildStage, error) {
	v, ok := b.buildStages.Load(stageName)
	if ok {
		if v.(BuildStage).Status() == Initialized {
			return v.(BuildStage), fmt.Errorf("stage '%s' is already being built - possible loop in the stage dependencies", stageName)
		}
		return v.(BuildStage), nil
	}

	dockerfile, err := template.NewDockerfileFromFile(b.buildDef.GetStageDockerfile(stageName), stageName, b.buildConf, b.localContext, b.buildDef.GetStageDirectory(stageName), b.templateStageResolver, b.exec)
	if err != nil {
		return nil, fmt.Errorf("failed to read the Dockerfile template: %w", err)
	}

	stage := NewBuildStage(stageName, dockerfile, b.buildConf.IgnorePatterns(stageName))
	b.buildStages.Store(stageName, stage)

	err = stage.Render()
	if err != nil {
		return stage, err
	}

	tag, err := stage.ImageTag()
	if err != nil {
		return stage, err
	}

	// TODO: Normalize Docker URL
	stage.SetImageURL(b.targetImage + ":" + stageName + "-" + tag)
	stage.SetSourceImageURL(b.targetImage + ":" + stageName + "-" + tag)
	if b.opts.CacheImagePull && b.buildConf.IsBuilderCacheSet() {
		cachedDockerImageWithTag := b.buildConf.Builder.ImageCache + "/" + b.buildConf.Builder.Name + ":" + stageName + "-" + tag
		exists, err := registry.ImageExists(cachedDockerImageWithTag)
		if err != nil {
			return stage, fmt.Errorf("error while verifying if image '%s' exists: %w", cachedDockerImageWithTag, err)
		}

		if exists {
			stage.SetStatus(ImageCached)
			stage.SetSourceImageURL(cachedDockerImageWithTag)
			return stage, nil
		}
	}

	if b.opts.CacheImagePull {
		exists, err := registry.ImageExists(stage.ImageURL())
		if err != nil {
			return stage, fmt.Errorf("error while verifying if image '%s' exists: %w", stage.ImageURL(), err)
		}

		if exists {
			stage.SetStatus(ImageCached)
			return stage, nil
		}
	}

	log.Debugf("No existing image for stage '%s' was found!", stageName)
	stage.SetStatus(ImageAbsent)
	return stage, nil
}

// buildStage builds a specific stage
func (b *Build) buildStage(stage BuildStage) error {
	log.Debugf("Trying to acquire lock on '%s'", stage.Name())
	b.locker.Lock(stage.Name())
	defer func() {
		b.locker.Unlock(stage.Name())
		log.Debugf("Releasing lock on '%s'", stage.Name())
	}()
	log.Debugf("Got lock on '%s'", stage.Name())

	// Log and exit of the nothing need to be done
	if stage.Status() == ImageCached {
		log.Infof("Image for stage '%s' (hash: '%s') is cached", stage.Name(), stage.ContentHash())
		return nil
	} else if stage.Status() == ImagePulled {
		log.Infof("Image for stage '%s' (hash: '%s') has been pulled", stage.Name(), stage.ContentHash())
		return nil
	} else if stage.Status() == ImageBuilt {
		log.Infof("Image for stage '%s' (hash: '%s') was built", stage.Name(), stage.ContentHash())
		return nil
	} else if stage.Status() != ImageAbsent {
		// e.g ImageInitialized
		return fmt.Errorf("image for stage '%s' (hash: '%s') has an invalid status: %v", stage.Name(), stage.ContentHash(), stage.Status())
	}

	log.Infof("Image for stage '%s' (hash: '%s') needs to be build", stage.Name(), stage.ContentHash())
	requiredStages := stage.GetRequiredStages()
	if len(requiredStages) > 0 {
		log.Debugf("Stage '%s' requires: %v", stage.Name(), stage.GetRequiredStages())
	} else {
		log.Debugf("Stage '%s' doesn't depend on any other stage", stage.Name())
	}

	// Build dependencies
	g, _ := errgroup.WithContext(context.Background())
	for _, s := range requiredStages {
		stageDep, ok := b.buildStages.Load(s)
		if !ok {
			return fmt.Errorf("stage '%s' dependency of '%s' was not prepared", s, stage.Name())
		}
		g.Go(func() error {
			log.Infof("Preparing build of stage '%s' as dependency of '%s'", stageDep.(BuildStage).Name(), stage.Name())
			if err := b.ensureDependencyPresence(stageDep.(BuildStage)); err != nil {
				return fmt.Errorf("error while ensuring presence of image for stage '%s' dependency of stage '%s': %w", stageDep.(BuildStage).Name(), stage.Name(), err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Build image
	log.Debugf("Trying to acquire concurrency semaphore for '%s'", stage.Name())
	if err := b.sem.Acquire(context.Background(), 1); err != nil {
		return err
	}
	log.Debugf("Acquired concurrency semaphore for '%s'", stage.Name())

	err := func() error {
		defer func() {
			b.sem.Release(1)
			log.Debugf("Releasing concurrency semaphore on '%s'", stage.Name())
		}()

		log.Infof("Building stage '%s'", stage.Name())
		return stage.Build(b.engine)
	}()
	if err != nil {
		return fmt.Errorf("error while building stage '%s': %w", stage.Name(), err)
	}

	// Eventually push the image
	stage.SetStatus(ImageBuilt)
	log.Infof("Stage '%s' successfully built!", stage.Name())
	if b.opts.CacheImagePush {
		if err := b.pushStage(stage); err != nil {
			return fmt.Errorf("error while pusing image for stage '%s': %w", stage.Name(), err)
		}
	}
	return nil
}

// ensureDependencyPresence pull the dependencies for a Stage and retags them to match the expected name
// or build them
func (b *Build) ensureDependencyPresence(stage BuildStage) error {
	if stage.Status() == ImageAbsent {
		err := b.buildStage(stage)
		if err != nil {
			return fmt.Errorf("error while building dependency '%s' required for stage '%s': %w", stage.SourceImageURL(), stage.Name(), err)
		}
		return nil
	} else if stage.Status() == ImageCached {
		err := b.engine.Pull(stage.SourceImageURL())
		if err != nil {
			return fmt.Errorf("error while pulling image '%s' required for stage '%s': %w", stage.SourceImageURL(), stage.Name(), err)
		}

		err = b.engine.Tag(stage.SourceImageURL(), stage.ImageURL())
		if err != nil {
			return fmt.Errorf("error while tagging image '%s' required for stage '%s': %w", stage.SourceImageURL(), stage.Name(), err)
		}
		stage.SetStatus(ImagePulled)
	}
	return nil
}

// pushStage push stages
func (b *Build) pushStage(stage BuildStage) error {
	log.Infof("Pushing image '%s'", stage.(BuildStage).ImageURL())
	if err := b.engine.Push(stage.(BuildStage).ImageURL()); err != nil {
		return fmt.Errorf("error while pushing image for stage '%s': %w", stage.Name(), err)
	}

	for _, tag := range stage.GetTagAliases() {
		log.Infof("Tagging image '%s' as '%s'", stage.ImageURL(), tag)
		if err := registry.TagImage(stage.ImageURL(), tag); err != nil {
			return fmt.Errorf("error while tagging image for stage '%s' with '%s': %w", stage.Name(), tag, err)
		}
	}
	return nil
}

// getBuildStages returns in which order the stages should be build
func (b *Build) getBuildStages() []BuildStage {
	stages := []BuildStage{}
	b.buildStages.Range(func(key, value interface{}) bool {
		stages = append(stages, value.(BuildStage))
		return true
	})
	return stages
}
