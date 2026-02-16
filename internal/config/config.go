package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tnm/email-cli/internal/keychain"
)

type ProviderType string

const (
	ProviderGoogle    ProviderType = "google"
	ProviderProton    ProviderType = "proton"
	ProviderSMTP      ProviderType = "smtp"
	ProviderAgentMail ProviderType = "agentmail"
)

type GoogleConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenExpiry  string `json:"token_expiry,omitempty"`
}

type ProtonConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseTLS   bool   `json:"use_tls"`
}

type AgentMailConfig struct {
	APIKey  string `json:"api_key"`
	InboxID string `json:"inbox_id"`
}

type ProviderConfig struct {
	Type      ProviderType     `json:"type"`
	Name      string           `json:"name"`
	From      string           `json:"from"`
	Google    *GoogleConfig    `json:"google,omitempty"`
	Proton    *ProtonConfig    `json:"proton,omitempty"`
	SMTP      *SMTPConfig      `json:"smtp,omitempty"`
	AgentMail *AgentMailConfig `json:"agentmail,omitempty"`
}

type Config struct {
	DefaultProvider string                    `json:"default_provider"`
	Providers       map[string]ProviderConfig `json:"providers"`
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "email-cli"), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{
			Providers: make(map[string]ProviderConfig),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	if name == "" {
		name = c.DefaultProvider
	}
	if name == "" {
		return nil, fmt.Errorf("no provider specified and no default set")
	}

	provider, ok := c.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", name)
	}

	return &provider, nil
}

// ResolveSecrets resolves any keychain references in the provider config.
// Returns a copy with secrets resolved (original is not modified).
func (p *ProviderConfig) ResolveSecrets() (*ProviderConfig, error) {
	resolved := *p

	switch p.Type {
	case ProviderSMTP:
		if p.SMTP != nil {
			smtpCfg := *p.SMTP
			if keychain.IsKeychainRef(smtpCfg.Password) {
				secret, err := keychain.Resolve(smtpCfg.Password)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve SMTP password: %w", err)
				}
				smtpCfg.Password = secret
			}
			resolved.SMTP = &smtpCfg
		}

	case ProviderProton:
		if p.Proton != nil {
			protonCfg := *p.Proton
			if keychain.IsKeychainRef(protonCfg.Password) {
				secret, err := keychain.Resolve(protonCfg.Password)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve Proton password: %w", err)
				}
				protonCfg.Password = secret
			}
			resolved.Proton = &protonCfg
		}

	case ProviderGoogle:
		if p.Google != nil {
			googleCfg := *p.Google
			if keychain.IsKeychainRef(googleCfg.ClientSecret) {
				secret, err := keychain.Resolve(googleCfg.ClientSecret)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve Google client secret: %w", err)
				}
				googleCfg.ClientSecret = secret
			}
			if keychain.IsKeychainRef(googleCfg.AccessToken) {
				secret, err := keychain.Resolve(googleCfg.AccessToken)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve Google access token: %w", err)
				}
				googleCfg.AccessToken = secret
			}
			if keychain.IsKeychainRef(googleCfg.RefreshToken) {
				secret, err := keychain.Resolve(googleCfg.RefreshToken)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve Google refresh token: %w", err)
				}
				googleCfg.RefreshToken = secret
			}
			resolved.Google = &googleCfg
		}

	case ProviderAgentMail:
		if p.AgentMail != nil {
			agentMailCfg := *p.AgentMail
			if keychain.IsKeychainRef(agentMailCfg.APIKey) {
				secret, err := keychain.Resolve(agentMailCfg.APIKey)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve AgentMail API key: %w", err)
				}
				agentMailCfg.APIKey = secret
			}
			resolved.AgentMail = &agentMailCfg
		}
	}

	return &resolved, nil
}
