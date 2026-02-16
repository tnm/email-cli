package cmd

import (
	"fmt"

	"github.com/tnm/email-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func configRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a provider configuration",
		ArgsUsage: "<name>",
		Action:    runConfigRemove,
	}
}

func runConfigRemove(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("usage: email-cli config remove <name>")
	}
	name := c.Args().First()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Providers[name]; !exists {
		return fmt.Errorf("provider %q not found", name)
	}

	delete(cfg.Providers, name)

	if cfg.DefaultProvider == name {
		cfg.DefaultProvider = ""
		for n := range cfg.Providers {
			cfg.DefaultProvider = n
			break
		}
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Provider %q removed.\n", name)
	return nil
}
