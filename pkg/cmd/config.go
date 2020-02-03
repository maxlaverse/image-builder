package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewConfigCmd returns a Cobra Command to configuration the tool
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			//TODO: Implement
			return fmt.Errorf("Not implemented")
		},
	}

	cmdSet := &cobra.Command{
		Use: "set",
		RunE: func(cmd *cobra.Command, args []string) error {
			//TODO: Implement
			return fmt.Errorf("Not implemented")
		},
	}

	cmdGet := &cobra.Command{
		Use: "get",
		RunE: func(cmd *cobra.Command, args []string) error {
			//TODO: Implement
			return fmt.Errorf("Not implemented")
		},
	}

	cmdView := &cobra.Command{
		Use: "view",
		RunE: func(cmd *cobra.Command, args []string) error {
			//TODO: Implement
			return fmt.Errorf("Not implemented")
		},
	}

	cmd.AddCommand(cmdGet)
	cmd.AddCommand(cmdSet)
	cmd.AddCommand(cmdView)

	return cmd
}
