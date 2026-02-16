package cmd

import (
	"fmt"
	"io"
	"os"
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
		cleanupKeychainSecrets(&p)
	}

	delete(cfg.Providers, name)

	if cfg.DefaultProvider == name {
		cfg.DefaultProvider = selectNewDefault(cfg.Providers)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Provider %q removed.\n", name)
	return nil
}

// selectNewDefault picks a deterministic replacement default provider.
func selectNewDefault(providers map[string]config.ProviderConfig) string {
	if len(providers) == 0 {
		return ""
	}

	names := make([]string, 0, len(providers))
	for n := range providers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names[0]
}

// cleanupKeychainSecrets removes any keychain entries associated with a provider.
func cleanupKeychainSecrets(p *config.ProviderConfig) {
	cleanupKeychainSecretsWithDeleter(p, keychain.Delete, os.Stderr)
}

func cleanupKeychainSecretsWithDeleter(p *config.ProviderConfig, deleteFn func(string) error, errWriter io.Writer) {
	for _, account := range keychainAccountsToDelete(p) {
		if err := deleteFn(account); err != nil {
			fmt.Fprintf(errWriter, "Warning: failed to remove keychain entry %q: %v\n", account, err)
		}
	}
}

func keychainAccountsToDelete(p *config.ProviderConfig) []string {
	if p == nil {
		return nil
	}

	// Parse actual keychain references from config values.
	var secretsToDelete []string

	switch p.Type {
	case config.ProviderAgentMail:
		if p.AgentMail != nil && keychain.IsKeychainRef(p.AgentMail.APIKey) {
			secretsToDelete = append(secretsToDelete, keychain.ParseKeychainRef(p.AgentMail.APIKey))
		}
	case config.ProviderSMTP:
		if p.SMTP != nil && keychain.IsKeychainRef(p.SMTP.Password) {
			secretsToDelete = append(secretsToDelete, keychain.ParseKeychainRef(p.SMTP.Password))
		}
	case config.ProviderProton:
		if p.Proton != nil && keychain.IsKeychainRef(p.Proton.Password) {
			secretsToDelete = append(secretsToDelete, keychain.ParseKeychainRef(p.Proton.Password))
		}
	case config.ProviderGoogle:
		if p.Google != nil {
			if keychain.IsKeychainRef(p.Google.ClientSecret) {
				secretsToDelete = append(secretsToDelete, keychain.ParseKeychainRef(p.Google.ClientSecret))
			}
			if keychain.IsKeychainRef(p.Google.AccessToken) {
				secretsToDelete = append(secretsToDelete, keychain.ParseKeychainRef(p.Google.AccessToken))
			}
			if keychain.IsKeychainRef(p.Google.RefreshToken) {
				secretsToDelete = append(secretsToDelete, keychain.ParseKeychainRef(p.Google.RefreshToken))
			}
		}
	}

	return secretsToDelete
}
