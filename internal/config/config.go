package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ProviderType string

const (
	ProviderGoogle ProviderType = "google"
	ProviderProton ProviderType = "proton"
	ProviderSMTP   ProviderType = "smtp"
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

type ProviderConfig struct {
	Type   ProviderType  `json:"type"`
	Name   string        `json:"name"`
	From   string        `json:"from"`
	Google *GoogleConfig `json:"google,omitempty"`
	Proton *ProtonConfig `json:"proton,omitempty"`
	SMTP   *SMTPConfig   `json:"smtp,omitempty"`
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
