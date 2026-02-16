package cmd

import (
	"testing"

	"github.com/tnm/email-cli/internal/config"
)

func TestRedactProviderConfig_RedactsSecrets(t *testing.T) {
	in := config.ProviderConfig{
		Type: config.ProviderGoogle,
		Name: "work",
		From: "me@example.com",
		Google: &config.GoogleConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			AccessToken:  "access",
			RefreshToken: "refresh",
		},
	}

	got := redactProviderConfig(in)
	if got.Google == nil {
		t.Fatalf("redacted google config is nil")
	}
	if got.Google.ClientSecret != "[REDACTED]" {
		t.Fatalf("client secret = %q, want [REDACTED]", got.Google.ClientSecret)
	}
	if got.Google.AccessToken != "[REDACTED]" {
		t.Fatalf("access token = %q, want [REDACTED]", got.Google.AccessToken)
	}
	if got.Google.RefreshToken != "[REDACTED]" {
		t.Fatalf("refresh token = %q, want [REDACTED]", got.Google.RefreshToken)
	}
}

func TestRedactProviderConfig_PreservesEmptyTokens(t *testing.T) {
	in := config.ProviderConfig{
		Type: config.ProviderGoogle,
		Google: &config.GoogleConfig{
			ClientID:     "id",
			ClientSecret: "",
			AccessToken:  "",
			RefreshToken: "",
		},
	}

	got := redactProviderConfig(in)
	if got.Google == nil {
		t.Fatalf("redacted google config is nil")
	}
	if got.Google.ClientSecret != "[REDACTED]" {
		t.Fatalf("client secret = %q, want [REDACTED]", got.Google.ClientSecret)
	}
	if got.Google.AccessToken != "" {
		t.Fatalf("access token = %q, want empty", got.Google.AccessToken)
	}
	if got.Google.RefreshToken != "" {
		t.Fatalf("refresh token = %q, want empty", got.Google.RefreshToken)
	}
}

func TestRedactConfig_RedactsNestedProviderSecrets(t *testing.T) {
	orig := &config.Config{
		DefaultProvider: "smtp1",
		Providers: map[string]config.ProviderConfig{
			"smtp1": {
				Type: config.ProviderSMTP,
				SMTP: &config.SMTPConfig{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "me",
					Password: "smtp-secret",
					UseTLS:   true,
				},
			},
			"proton1": {
				Type: config.ProviderProton,
				Proton: &config.ProtonConfig{
					Host:     "127.0.0.1",
					Port:     1025,
					Username: "me@proton.me",
					Password: "proton-secret",
				},
			},
		},
	}

	got := redactConfig(orig)
	if got.DefaultProvider != "smtp1" {
		t.Fatalf("default provider = %q, want smtp1", got.DefaultProvider)
	}
	if got.Providers["smtp1"].SMTP.Password != "[REDACTED]" {
		t.Fatalf("smtp password not redacted")
	}
	if got.Providers["proton1"].Proton.Password != "[REDACTED]" {
		t.Fatalf("proton password not redacted")
	}

	// Ensure original config is unchanged.
	if orig.Providers["smtp1"].SMTP.Password != "smtp-secret" {
		t.Fatalf("original smtp password mutated")
	}
	if orig.Providers["proton1"].Proton.Password != "proton-secret" {
		t.Fatalf("original proton password mutated")
	}
}

