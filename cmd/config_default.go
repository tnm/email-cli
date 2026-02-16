package cmd

import (
	"fmt"

	"github.com/tnm/email-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func configDefaultCommand() *cli.Command {
	return &cli.Command{
		Name:      "default",
		Usage:     "Set default provider",
		ArgsUsage: "<name>",
		Action:    runConfigDefault,
	}
}

func runConfigDefault(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("usage: email-cli config default <name>")
	}
	name := c.Args().First()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Providers[name]; !exists {
		return fmt.Errorf("provider %q not found", name)
	}

	cfg.DefaultProvider = name

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Default provider set to %q.\n", name)
	return nil
}
