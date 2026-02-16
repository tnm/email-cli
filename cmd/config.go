package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

func configCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage email provider configuration",
		Subcommands: []*cli.Command{
			configAddCommand(),
			configListCommand(),
			configShowCommand(),
			configRemoveCommand(),
			configDefaultCommand(),
			configPathCommand(),
			configSetCommand(),
		},
	}
}

// prompt reads a line from the reader with a label.
func prompt(reader *bufio.Reader, label string) string {
	fmt.Printf("%s: ", label)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// promptDefault reads a line with a default value shown.
func promptDefault(reader *bufio.Reader, label, defaultVal string) string {
	fmt.Printf("%s [%s]: ", label, defaultVal)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	return text
}

// isNonInteractive returns true if flags indicate non-interactive mode.
func isNonInteractive(c *cli.Context) bool {
	return c.String("type") != ""
}
