package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

func Execute() {
	app := &cli.App{
		Name:  "email-cli",
		Usage: "A CLI for sending emails via multiple providers",
		Description: "email-cli supports sending emails through:\n" +
			"  - Google Workspace (OAuth2)\n" +
			"  - Proton Mail (via Bridge)\n" +
			"  - Generic SMTP\n\n" +
			"Perfect for automation, scripts, and AI agents.",
		Commands: []*cli.Command{
			sendCommand(),
			configCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		// Keep behavior similar to cobra root Execute(): print error and exit 1.
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

