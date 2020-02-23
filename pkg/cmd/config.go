package cmd

import (
	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/spf13/cobra"
)

// NewConfigCmd returns a Cobra command to configure the tool
func NewConfigCmd(conf *config.CliConfiguration) *cobra.Command {
	cmd := &cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(NewConfigViewCmd(conf))
	cmd.AddCommand(NewConfigSetCmd(conf))

	return cmd
}
