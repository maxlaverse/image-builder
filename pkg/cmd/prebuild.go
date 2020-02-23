package cmd

import (
	"path"
	"path/filepath"

	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/spf13/cobra"
)

// NewPreBuildCmd returns a Cobra command to prebuild combination of images
func NewPreBuildCmd(conf *config.CliConfiguration) *cobra.Command {
	var opts buildCommandOptions
	cmd := &cobra.Command{
		Use:              "prebuild [options]",
		Short:            "Builds a combination of images",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildStageBase(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.buildConfiguration, "prebuild-config", "c", "prebuild.yaml", "Configuration file of the application")
	cmd.Flags().BoolVarP(&opts.cacheImagePull, "cache-image-pull", "", conf.DefaultCacheImagePull, "Pull cache images from the registry")
	cmd.Flags().BoolVarP(&opts.cacheImagePush, "cache-image-push", "", conf.DefaultCacheImagePush, "Push cache images to the registry")
	cmd.Flags().BoolVarP(&opts.dryRun, "dry-run", "", false, "Only display the generated Dockerfiles")
	cmd.Flags().StringVarP(&opts.engine, "engine", "", conf.DefaultEngine, "Engine to use for building images")
	cmd.Flags().StringVarP(&opts.targetImage, "target-image", "t", "", "Specifies the name which will be assigned to the resulting image if the build process completes successfully")

	return cmd
}

func buildStageBase(opts buildCommandOptions) error {
	prebuildConf, err := config.ReadPrebuildConfiguration(opts.buildConfiguration)
	if err != nil {
		return err
	}

	source, err := filepath.Abs(path.Dir(opts.buildConfiguration))
	if err != nil {
		return err
	}

	for _, conf := range prebuildConf.BasePreBuild {
		buildConf := config.BuildConfiguration{
			Builder: config.BuildBuilderConfiguration{
				Location:   source,
				ImageCache: prebuildConf.BuilderCache,
				Name:       conf.BuilderName,
			},
			ImageSpec: conf.Spec,
		}

		if len(conf.BuilderName) == 0 {
			panic("No builder name")
		}
		opts.targetImage = prebuildConf.BuilderCache + "/" + conf.BuilderName

		err := buildStageGeneric(opts, conf.Stages, buildConf, path.Join(source, conf.BuilderName))
		if err != nil {
			return err
		}
	}
	return nil
}
