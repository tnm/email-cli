package cmd

import (
	"fmt"
	"sort"

	"github.com/tnm/email-cli/internal/config"
	"github.com/tnm/email-cli/internal/keychain"
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

	p, exists := cfg.Providers[name]
	if !exists {
		return fmt.Errorf("provider %q not found", name)
	}

	// Clean up any keychain entries for this provider
	if keychain.IsSupported() {
		cleanupKeychainSecrets(name, &p)
	}

	delete(cfg.Providers, name)

	if cfg.DefaultProvider == name {
		cfg.DefaultProvider = ""
		// Pick alphabetically first provider for deterministic behavior
		if len(cfg.Providers) > 0 {
			names := make([]string, 0, len(cfg.Providers))
			for n := range cfg.Providers {
				names = append(names, n)
			}
			sort.Strings(names)
			cfg.DefaultProvider = names[0]
		}
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Provider %q removed.\n", name)
	return nil
}

// cleanupKeychainSecrets removes any keychain entries associated with a provider
func cleanupKeychainSecrets(name string, p *config.ProviderConfig) {
	// Check each possible secret field for keychain references
	var secretsToDelete []string

	switch p.Type {
	case config.ProviderAgentMail:
		if p.AgentMail != nil && keychain.IsKeychainRef(p.AgentMail.APIKey) {
			secretsToDelete = append(secretsToDelete, name+"/api-key")
		}
	case config.ProviderSMTP:
		if p.SMTP != nil && keychain.IsKeychainRef(p.SMTP.Password) {
			secretsToDelete = append(secretsToDelete, name+"/password")
		}
	case config.ProviderProton:
		if p.Proton != nil && keychain.IsKeychainRef(p.Proton.Password) {
			secretsToDelete = append(secretsToDelete, name+"/password")
		}
	case config.ProviderGoogle:
		if p.Google != nil {
			if keychain.IsKeychainRef(p.Google.ClientSecret) {
				secretsToDelete = append(secretsToDelete, name+"/client-secret")
			}
			if keychain.IsKeychainRef(p.Google.AccessToken) {
				secretsToDelete = append(secretsToDelete, name+"/access-token")
			}
			if keychain.IsKeychainRef(p.Google.RefreshToken) {
				secretsToDelete = append(secretsToDelete, name+"/refresh-token")
			}
		}
	}

	for _, account := range secretsToDelete {
		_ = keychain.Delete(account) // Ignore errors, best effort cleanup
	}
}
