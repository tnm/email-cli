package cmd

import (
	"fmt"
	"strconv"

	"github.com/tnm/email-cli/internal/config"
	"github.com/urfave/cli/v2"
)

func configSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set a config value (for agents)",
		ArgsUsage: "<name> <key> <value>",
		Description: "Set a specific configuration value.\n\n" +
			"Keys for AgentMail:\n" +
			"  api-key, inbox-id\n\n" +
			"Keys for SMTP/Proton:\n" +
			"  from, host, port, username, password, tls\n\n" +
			"Keys for Google:\n" +
			"  from, client-id, client-secret, access-token, refresh-token\n\n" +
			"Examples:\n" +
			"  email-cli config set mymail password \"new-password\"\n" +
			"  email-cli config set mymail host smtp.newserver.com\n" +
			"  email-cli config set agent api-key \"am_...\"",
		Action: runConfigSet,
	}
}

func runConfigSet(c *cli.Context) error {
	if c.Args().Len() != 3 {
		return fmt.Errorf("usage: email-cli config set <name> <key> <value>")
	}
	name, key, value := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	p, exists := cfg.Providers[name]
	if !exists {
		return fmt.Errorf("provider %q not found", name)
	}

	switch key {
	case "from":
		p.From = value

	case "host":
		switch p.Type {
		case config.ProviderSMTP:
			if p.SMTP == nil {
				return fmt.Errorf("smtp config missing for %q", name)
			}
			p.SMTP.Host = value
		case config.ProviderProton:
			if p.Proton == nil {
				return fmt.Errorf("proton config missing for %q", name)
			}
			p.Proton.Host = value
		default:
			return fmt.Errorf("key %q not valid for provider type %s", key, p.Type)
		}

	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		switch p.Type {
		case config.ProviderSMTP:
			if p.SMTP == nil {
				return fmt.Errorf("smtp config missing for %q", name)
			}
			p.SMTP.Port = port
		case config.ProviderProton:
			if p.Proton == nil {
				return fmt.Errorf("proton config missing for %q", name)
			}
			p.Proton.Port = port
		default:
			return fmt.Errorf("key %q not valid for provider type %s", key, p.Type)
		}

	case "username":
		switch p.Type {
		case config.ProviderSMTP:
			if p.SMTP == nil {
				return fmt.Errorf("smtp config missing for %q", name)
			}
			p.SMTP.Username = value
		case config.ProviderProton:
			if p.Proton == nil {
				return fmt.Errorf("proton config missing for %q", name)
			}
			p.Proton.Username = value
		default:
			return fmt.Errorf("key %q not valid for provider type %s", key, p.Type)
		}

	case "password":
		switch p.Type {
		case config.ProviderSMTP:
			if p.SMTP == nil {
				return fmt.Errorf("smtp config missing for %q", name)
			}
			p.SMTP.Password = value
		case config.ProviderProton:
			if p.Proton == nil {
				return fmt.Errorf("proton config missing for %q", name)
			}
			p.Proton.Password = value
		default:
			return fmt.Errorf("key %q not valid for provider type %s", key, p.Type)
		}

	case "tls":
		if p.Type != config.ProviderSMTP {
			return fmt.Errorf("key %q only valid for SMTP provider", key)
		}
		if p.SMTP == nil {
			return fmt.Errorf("smtp config missing for %q", name)
		}
		p.SMTP.UseTLS = value == "true" || value == "1" || value == "yes"

	case "client-id":
		if p.Type != config.ProviderGoogle {
			return fmt.Errorf("key %q only valid for Google provider", key)
		}
		if p.Google == nil {
			return fmt.Errorf("google config missing for %q", name)
		}
		p.Google.ClientID = value

	case "client-secret":
		if p.Type != config.ProviderGoogle {
			return fmt.Errorf("key %q only valid for Google provider", key)
		}
		if p.Google == nil {
			return fmt.Errorf("google config missing for %q", name)
		}
		p.Google.ClientSecret = value

	case "access-token":
		if p.Type != config.ProviderGoogle {
			return fmt.Errorf("key %q only valid for Google provider", key)
		}
		if p.Google == nil {
			return fmt.Errorf("google config missing for %q", name)
		}
		p.Google.AccessToken = value

	case "refresh-token":
		if p.Type != config.ProviderGoogle {
			return fmt.Errorf("key %q only valid for Google provider", key)
		}
		if p.Google == nil {
			return fmt.Errorf("google config missing for %q", name)
		}
		p.Google.RefreshToken = value

	case "api-key":
		if p.Type != config.ProviderAgentMail {
			return fmt.Errorf("key %q only valid for AgentMail provider", key)
		}
		if p.AgentMail == nil {
			return fmt.Errorf("agentmail config missing for %q", name)
		}
		p.AgentMail.APIKey = value

	case "inbox-id":
		if p.Type != config.ProviderAgentMail {
			return fmt.Errorf("key %q only valid for AgentMail provider", key)
		}
		if p.AgentMail == nil {
			return fmt.Errorf("agentmail config missing for %q", name)
		}
		p.AgentMail.InboxID = value

	default:
		return fmt.Errorf("unknown key %q", key)
	}

	cfg.Providers[name] = p

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Updated %s.%s\n", name, key)
	return nil
}
