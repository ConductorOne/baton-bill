package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	Username        string   `mapstructure:"username"`
	Password        string   `mapstructure:"password"`
	OrganizationIds []string `mapstructure:"organizationIds"`
	DeveloperKey    string   `mapstructure:"developerKey"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.Username == "" {
		return fmt.Errorf("username is missing")
	}

	if cfg.Password == "" {
		return fmt.Errorf("password is missing")
	}

	if cfg.OrganizationIds == nil || len(cfg.OrganizationIds) == 0 {
		return fmt.Errorf("organizationIds are missing")
	}

	if cfg.DeveloperKey == "" {
		return fmt.Errorf("developerKey is missing")
	}

	return nil
}

// cmdFlags sets the cmdFlags required for the connector.
func cmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("username", "", "The Bill username used to connect to the Bill API. ($BATON_BILL_USERNAME)")
	cmd.PersistentFlags().String("password", "", "The Bill password used to connect to the Bill API. ($BATON_BILL_PASSWORD)")
	cmd.PersistentFlags().StringSlice("organizationIds", []string{}, "The Bill organizationIds used to connect to the Bill API. ($BATON_BILL_ORGANIZATION_IDS)")
	cmd.PersistentFlags().String("developerKey", "", "The Bill developerKey used to connect to the Bill API. ($BATON_BILL_DEVELOPER_KEY)")
}
