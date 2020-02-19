package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/maxlaverse/image-builder/pkg/builder"
	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/maxlaverse/image-builder/pkg/engine"
	"github.com/maxlaverse/image-builder/pkg/executor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type buildCommandOptions struct {
	buildConfiguration string
	cacheImagePush     bool
	cacheImagePull     bool
	dryRun             bool
	engine             string
	targetImage        string
	targetStages       []string
	extraTags          map[string][]string
}

// NewBuildCmd returns a Cobra Command to build images
func NewBuildCmd() *cobra.Command {
	var opts buildCommandOptions
	var extraTagArray []string
	cmd := &cobra.Command{
		Use:              "build [options] <buildContext>",
		Short:            "Build an image from a Build Definition file",
		TraverseChildren: true,
		SilenceUsage:     true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Wrong number of argument")
			}
			opts.extraTags = map[string][]string{}
			for _, v := range extraTagArray {
				parts := strings.Split(v, "=")
				if len(parts) < 2 {
					return fmt.Errorf("Invalid extra tag format")
				}
				if opts.extraTags[parts[0]] == nil {
					opts.extraTags[parts[0]] = []string{}
				}
				opts.extraTags[parts[0]] = append(opts.extraTags[parts[0]], parts[1])
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildStageApp(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.buildConfiguration, "config", "c", "build.yaml", "Configuration file of the application")
	cmd.Flags().BoolVarP(&opts.cacheImagePull, "cache-image-pull", "", true, "Pull cache images from the registry")
	cmd.Flags().BoolVarP(&opts.cacheImagePush, "cache-image-push", "", true, "Push cache images to the registry")
	cmd.Flags().BoolVarP(&opts.dryRun, "dry-run", "", false, "Only display the generated Dockerfiles")
	cmd.Flags().StringVarP(&opts.engine, "engine", "", "docker", "Engine to use for building images")
	cmd.Flags().StringVarP(&opts.targetImage, "target-image", "t", "", "Specifies the name which will be assigned to the resulting image if the build process completes successfully")
	cmd.Flags().StringArrayVarP(&extraTagArray, "extra-tag", "", []string{}, "Extra tag if the stage was built (format: <stage>=<tag>")
	cmd.Flags().StringArrayVarP(&opts.targetStages, "target-stages", "s", []string{"release"}, "Specifies the stages to build")

	return cmd
}

func buildStageApp(opts buildCommandOptions, buildContext string) error {
	buildConf, err := config.ReadBuildConfiguration(opts.buildConfiguration)
	if err != nil {
		return err
	}

	if len(opts.targetImage) == 0 {
		opts.cacheImagePush = false
		opts.targetImage = generatedTargetName()
		log.Infof("No target image name has been provided. Using '%s'", opts.targetImage)
	}

	buildContext, err = filepath.Abs(path.Dir(buildContext))
	if err != nil {
		return err
	}
	return buildStageGeneric(opts, opts.targetStages, buildConf, buildContext)
}

func buildStageGeneric(opts buildCommandOptions, stages []string, buildConf config.BuildConfiguration, buildContext string) error {
	builderDef, err := config.NewBuilderDef(buildConf.Builder.Location, buildConf.Builder.Name)
	if err != nil {
		return err
	}

	engineCli, err := engine.New(opts.engine, executor.New())
	if err != nil {
		return err
	}

	b := builder.NewBuild(engineCli, executor.New(), buildConf, builderDef, opts.dryRun, opts.cacheImagePull, opts.cacheImagePush, opts.targetImage, buildContext)
	orderedStages, err := b.GetStageBuildOrder(stages)
	if err != nil {
		return err
	}

	log.Infof("The stages will be build in the following order: %s\n", strings.Join(orderedStages, ", "))
	imageURLs := []string{}
	for _, stage := range orderedStages {
		// Build stage
		buildStatus, err := b.BuildStage(stage)
		if err != nil {
			return err
		}

		// Add extra tags
		if buildStatus.Built {
			for _, j := range opts.extraTags[stage] {
				err = engineCli.Tag(buildStatus.ImageURL, opts.targetImage+":"+j)
				if err != nil {
					return err
				}
			}
		}

		// Compute image URLs to display in the summary
		if len(opts.extraTags[stage]) > 0 {
			imageURLs = append(imageURLs, fmt.Sprintf("%s [built:%v] (extra tag: %s)", buildStatus.ImageURL, buildStatus.Built, strings.Join(opts.extraTags[stage], ", ")))
		} else {
			imageURLs = append(imageURLs, fmt.Sprintf("%s [built:%v]", buildStatus.ImageURL, buildStatus.Built))
		}
	}

	log.Info("Build finished! The following images have been pulled or built:")
	for _, image := range imageURLs {
		log.Infof("* %s\n", image)
	}

	return nil
}

func generatedTargetName() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "panic"
	}

	return fmt.Sprintf("generated-%s", filepath.Base(dir))
}
