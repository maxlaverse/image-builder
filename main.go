package main

import (
	"math"
	"os"
	"path"

	"github.com/maxlaverse/image-builder/pkg/cmd"
	"github.com/maxlaverse/image-builder/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	conf := config.NewDefaultConfiguration()
	conf.Load(getConfigurationPath())

	verbose := 0
	command := &cobra.Command{
		Use:              "image-builder",
		Long:             "Build container images for many different application types",
		TraverseChildren: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.Level(math.Min(float64(verbose+4), 6.0)))
			log.SetFormatter(&log.TextFormatter{})
		},
	}
	command.PersistentFlags().IntVarP(&verbose, "verbose", "v", 0, "Be verbose on log output")

	command.AddCommand(cmd.NewBuildCmd(conf))
	command.AddCommand(cmd.NewPreBuildCmd(conf))
	command.AddCommand(cmd.NewConfigCmd(conf))

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func getConfigurationPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return path.Join(home, ".image-builder", "config.yaml")
}
