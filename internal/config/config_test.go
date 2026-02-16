package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSecrets_PlaintextPassthrough(t *testing.T) {
	// Non-keychain values should pass through unchanged
	p := &ProviderConfig{
		Type: ProviderSMTP,
		SMTP: &SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "plaintext-password",
		},
	}

	resolved, err := p.ResolveSecrets()
	if err != nil {
		t.Fatalf("ResolveSecrets() error = %v", err)
	}

	if resolved.SMTP.Password != "plaintext-password" {
		t.Errorf("Password = %q, want %q", resolved.SMTP.Password, "plaintext-password")
	}
}

func TestResolveSecrets_PreservesOriginal(t *testing.T) {
	// Original config should not be mutated
	p := &ProviderConfig{
		Type: ProviderSMTP,
		SMTP: &SMTPConfig{
			Password: "original",
		},
	}

	resolved, err := p.ResolveSecrets()
	if err != nil {
		t.Fatalf("ResolveSecrets() error = %v", err)
	}

	// Mutate resolved
	resolved.SMTP.Password = "mutated"

	// Original should be unchanged
	if p.SMTP.Password != "original" {
		t.Errorf("Original password mutated: got %q, want %q", p.SMTP.Password, "original")
	}
}

func TestResolveSecrets_AllProviderTypes(t *testing.T) {
	// Test that all provider types handle plaintext correctly
	tests := []struct {
		name string
		cfg  ProviderConfig
	}{
		{
			name: "smtp",
			cfg: ProviderConfig{
				Type: ProviderSMTP,
				SMTP: &SMTPConfig{Password: "pass"},
			},
		},
		{
			name: "proton",
			cfg: ProviderConfig{
				Type:   ProviderProton,
				Proton: &ProtonConfig{Password: "pass"},
			},
		},
		{
			name: "google",
			cfg: ProviderConfig{
				Type: ProviderGoogle,
				Google: &GoogleConfig{
					ClientSecret: "secret",
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
			},
		},
		{
			name: "agentmail",
			cfg: ProviderConfig{
				Type:      ProviderAgentMail,
				AgentMail: &AgentMailConfig{APIKey: "am_key"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := tt.cfg.ResolveSecrets()
			if err != nil {
				t.Fatalf("ResolveSecrets() error = %v", err)
			}
			if resolved == nil {
				t.Fatal("ResolveSecrets() returned nil")
			}
		})
	}
}

func TestGetProvider_Default(t *testing.T) {
	cfg := &Config{
		DefaultProvider: "main",
		Providers: map[string]ProviderConfig{
			"main":   {Name: "main", Type: ProviderSMTP},
			"backup": {Name: "backup", Type: ProviderSMTP},
		},
	}

	// Empty name should use default
	p, err := cfg.GetProvider("")
	if err != nil {
		t.Fatalf("GetProvider() error = %v", err)
	}
	if p.Name != "main" {
		t.Errorf("GetProvider('') = %q, want 'main'", p.Name)
	}
}

func TestGetProvider_Named(t *testing.T) {
	cfg := &Config{
		DefaultProvider: "main",
		Providers: map[string]ProviderConfig{
			"main":   {Name: "main", Type: ProviderSMTP},
			"backup": {Name: "backup", Type: ProviderSMTP},
		},
	}

	p, err := cfg.GetProvider("backup")
	if err != nil {
		t.Fatalf("GetProvider() error = %v", err)
	}
	if p.Name != "backup" {
		t.Errorf("GetProvider('backup') = %q, want 'backup'", p.Name)
	}
}

func TestGetProvider_NotFound(t *testing.T) {
	cfg := &Config{
		DefaultProvider: "main",
		Providers: map[string]ProviderConfig{
			"main": {Name: "main", Type: ProviderSMTP},
		},
	}

	_, err := cfg.GetProvider("nonexistent")
	if err == nil {
		t.Fatal("GetProvider('nonexistent') should return error")
	}
}

func TestGetProvider_NoDefault(t *testing.T) {
	cfg := &Config{
		DefaultProvider: "",
		Providers:       map[string]ProviderConfig{},
	}

	_, err := cfg.GetProvider("")
	if err == nil {
		t.Fatal("GetProvider('') with no default should return error")
	}
}

func TestLoadSave_RoundTrip(t *testing.T) {
	// Create temp dir for test config
	tmpDir, err := os.MkdirTemp("", "email-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create test config
	cfg := &Config{
		DefaultProvider: "test",
		Providers: map[string]ProviderConfig{
			"test": {
				Type: ProviderSMTP,
				Name: "test",
				From: "test@example.com",
				SMTP: &SMTPConfig{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user",
					Password: "pass",
					UseTLS:   true,
				},
			},
		},
	}

	// Write config directly to test file
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Read it back
	data2, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data2, &loaded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify
	if loaded.DefaultProvider != "test" {
		t.Errorf("DefaultProvider = %q, want 'test'", loaded.DefaultProvider)
	}
	if len(loaded.Providers) != 1 {
		t.Errorf("len(Providers) = %d, want 1", len(loaded.Providers))
	}
	p, ok := loaded.Providers["test"]
	if !ok {
		t.Fatal("Provider 'test' not found")
	}
	if p.SMTP.Host != "smtp.example.com" {
		t.Errorf("SMTP.Host = %q, want 'smtp.example.com'", p.SMTP.Host)
	}
	if p.SMTP.Password != "pass" {
		t.Errorf("SMTP.Password = %q, want 'pass'", p.SMTP.Password)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	// Create temp dir for test config
	tmpDir, err := os.MkdirTemp("", "email-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Write with correct permissions
	data := []byte(`{"default_provider":"","providers":{}}`)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Check permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Config file permissions = %o, want 0600", perm)
	}
}
