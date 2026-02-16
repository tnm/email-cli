package cmd

import (
	"fmt"

	"github.com/tnm/email-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func configListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List configured providers",
		Action: runConfigList,
	}
}

func runConfigList(c *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Providers) == 0 {
		fmt.Println("No providers configured. Use 'email-cli config add <name>' to add one.")
		return nil
	}

	fmt.Println("Configured providers:")
	fmt.Println()
	for name, p := range cfg.Providers {
		defaultMark := ""
		if name == cfg.DefaultProvider {
			defaultMark = " (default)"
		}
		fmt.Printf("  %s%s\n", name, defaultMark)
		fmt.Printf("    Type: %s\n", p.Type)
		fmt.Printf("    From: %s\n", p.From)
	}

	return nil
}
