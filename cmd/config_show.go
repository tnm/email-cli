package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/tnm/email-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func configShowCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show provider config as JSON (for agents)",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "show-secrets", Usage: "Include passwords/tokens in output"},
		},
		Action: runConfigShow,
	}
}

func runConfigShow(c *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	showSecrets := c.Bool("show-secrets")
	var output interface{}
	if c.Args().Len() == 0 {
		if showSecrets {
			output = cfg
		} else {
			output = redactConfig(cfg)
		}
	} else if c.Args().Len() == 1 {
		name := c.Args().First()
		p, exists := cfg.Providers[name]
		if !exists {
			return fmt.Errorf("provider %q not found", name)
		}
		if showSecrets {
			output = p
		} else {
			output = redactProviderConfig(p)
		}
	} else {
		return fmt.Errorf("usage: email-cli config show [name]")
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func redactConfig(cfg *config.Config) *config.Config {
	redacted := &config.Config{
		DefaultProvider: cfg.DefaultProvider,
		Providers:       make(map[string]config.ProviderConfig, len(cfg.Providers)),
	}
	for name, providerCfg := range cfg.Providers {
		redacted.Providers[name] = redactProviderConfig(providerCfg)
	}
	return redacted
}

func redactProviderConfig(providerCfg config.ProviderConfig) config.ProviderConfig {
	redacted := providerCfg

	if redacted.SMTP != nil {
		smtpCfg := *redacted.SMTP
		smtpCfg.Password = "[REDACTED]"
		redacted.SMTP = &smtpCfg
	}

	if redacted.Proton != nil {
		protonCfg := *redacted.Proton
		protonCfg.Password = "[REDACTED]"
		redacted.Proton = &protonCfg
	}

	if redacted.Google != nil {
		googleCfg := *redacted.Google
		googleCfg.ClientSecret = "[REDACTED]"
		if googleCfg.AccessToken != "" {
			googleCfg.AccessToken = "[REDACTED]"
		}
		if googleCfg.RefreshToken != "" {
			googleCfg.RefreshToken = "[REDACTED]"
		}
		redacted.Google = &googleCfg
	}

	return redacted
}
