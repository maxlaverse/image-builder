package main

import (
	"fmt"
	"math"
	"os"

	"github.com/maxlaverse/image-builder/pkg/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
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

	command.AddCommand(cmd.NewBuildCmd())
	command.AddCommand(cmd.NewPreBuildCmd())
	command.AddCommand(cmd.NewConfigCmd())

	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
