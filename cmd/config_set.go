package cmd

import (
	"fmt"
	"strconv"

	"github.com/tnm/email-cli/internal/config"
	"github.com/tnm/email-cli/internal/keychain"
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
			"  email-cli config set agent api-key \"am_...\"\n" +
			"  email-cli config set --use-keychain agent api-key \"am_...\"",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "use-keychain", Usage: "Store secret in macOS Keychain"},
		},
		Action: runConfigSet,
	}
}

func runConfigSet(c *cli.Context) error {
	if c.Args().Len() != 3 {
		return fmt.Errorf("usage: email-cli config set <name> <key> <value>")
	}
	name, key, value := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)

	useKeychain := c.Bool("use-keychain")
	if useKeychain && !keychain.IsSupported() {
		return fmt.Errorf("--use-keychain is only supported on macOS")
	}

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
			if useKeychain || keychain.IsKeychainRef(p.SMTP.Password) {
				if err := keychain.Set(name+"/password", value); err != nil {
					return fmt.Errorf("failed to store password in keychain: %w", err)
				}
				p.SMTP.Password = keychain.KeychainRef(name, "password")
			} else {
				p.SMTP.Password = value
			}
		case config.ProviderProton:
			if p.Proton == nil {
				return fmt.Errorf("proton config missing for %q", name)
			}
			if useKeychain || keychain.IsKeychainRef(p.Proton.Password) {
				if err := keychain.Set(name+"/password", value); err != nil {
					return fmt.Errorf("failed to store password in keychain: %w", err)
				}
				p.Proton.Password = keychain.KeychainRef(name, "password")
			} else {
				p.Proton.Password = value
			}
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
		if useKeychain || keychain.IsKeychainRef(p.Google.ClientSecret) {
			if err := keychain.Set(name+"/client-secret", value); err != nil {
				return fmt.Errorf("failed to store client secret in keychain: %w", err)
			}
			p.Google.ClientSecret = keychain.KeychainRef(name, "client-secret")
		} else {
			p.Google.ClientSecret = value
		}

	case "access-token":
		if p.Type != config.ProviderGoogle {
			return fmt.Errorf("key %q only valid for Google provider", key)
		}
		if p.Google == nil {
			return fmt.Errorf("google config missing for %q", name)
		}
		if useKeychain || keychain.IsKeychainRef(p.Google.AccessToken) {
			if err := keychain.Set(name+"/access-token", value); err != nil {
				return fmt.Errorf("failed to store access token in keychain: %w", err)
			}
			p.Google.AccessToken = keychain.KeychainRef(name, "access-token")
		} else {
			p.Google.AccessToken = value
		}

	case "refresh-token":
		if p.Type != config.ProviderGoogle {
			return fmt.Errorf("key %q only valid for Google provider", key)
		}
		if p.Google == nil {
			return fmt.Errorf("google config missing for %q", name)
		}
		if useKeychain || keychain.IsKeychainRef(p.Google.RefreshToken) {
			if err := keychain.Set(name+"/refresh-token", value); err != nil {
				return fmt.Errorf("failed to store refresh token in keychain: %w", err)
			}
			p.Google.RefreshToken = keychain.KeychainRef(name, "refresh-token")
		} else {
			p.Google.RefreshToken = value
		}

	case "api-key":
		if p.Type != config.ProviderAgentMail {
			return fmt.Errorf("key %q only valid for AgentMail provider", key)
		}
		if p.AgentMail == nil {
			return fmt.Errorf("agentmail config missing for %q", name)
		}
		if useKeychain || keychain.IsKeychainRef(p.AgentMail.APIKey) {
			if err := keychain.Set(name+"/api-key", value); err != nil {
				return fmt.Errorf("failed to store API key in keychain: %w", err)
			}
			p.AgentMail.APIKey = keychain.KeychainRef(name, "api-key")
		} else {
			p.AgentMail.APIKey = value
		}

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
