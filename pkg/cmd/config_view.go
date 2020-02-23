package cmd

import (
	"fmt"

	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// NewConfigViewCmd returns a Cobra Command to configuration the tool
func NewConfigViewCmd(conf *config.CliConfiguration) *cobra.Command {
	cmd := &cobra.Command{
		Use: "view",
		RunE: func(cmd *cobra.Command, args []string) error {
			return viewConfig(conf)
		},
	}

	return cmd
}

func viewConfig(conf *config.CliConfiguration) error {
	data, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}
	fmt.Printf("%v", string(data))
	return nil
}
