package cmd

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/maxlaverse/image-builder/pkg/config"
	"github.com/spf13/cobra"
)

// NewConfigSetCmd returns a Cobra Command to set value
func NewConfigSetCmd(conf *config.CliConfiguration) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name> <value>",
		Short: "Sets an individual value in a configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Not enough argument")
			}
			return setConfigKey(conf, args[0], args[1])
		},
	}
	return cmd
}

func setConfigKey(conf *config.CliConfiguration, key, value string) error {
	for i := 0; i < reflect.TypeOf(conf).Elem().NumField(); i++ {
		field := reflect.TypeOf(conf).Elem().Field(i)
		if string(field.Tag.Get("yaml")) == key {
			r := reflect.ValueOf(conf)
			v := reflect.Indirect(r).FieldByName(field.Name)
			if field.Type.Kind() == reflect.String {
				v.SetString(value)
			} else if field.Type.Kind() == reflect.Bool {
				b, err := strconv.ParseBool(value)
				if err != nil {
					return err
				}
				v.SetBool(b)
			} else {
				return fmt.Errorf("Unsupported type for field '%s'", field.Name)
			}
		}
	}

	err := conf.Save()
	if err != nil {
		return err
	}
	return nil
}
