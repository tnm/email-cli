package cmd

import (
	"fmt"

	"github.com/tnm/email-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func configPathCommand() *cli.Command {
	return &cli.Command{
		Name:   "path",
		Usage:  "Show config file path",
		Action: runConfigPath,
	}
}

func runConfigPath(c *cli.Context) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}
