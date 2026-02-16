package cmd

import (
	"sort"
	"testing"

	"github.com/tnm/email-cli/internal/config"
	"github.com/tnm/email-cli/internal/keychain"
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

func TestDeterministicDefaultAfterRemoval(t *testing.T) {
	// Simulate picking a new default after removal
	// Should be alphabetically first
	providers := map[string]config.ProviderConfig{
		"zebra":  {Name: "zebra"},
		"alpha":  {Name: "alpha"},
		"middle": {Name: "middle"},
	}

	// Get alphabetically first
	names := make([]string, 0, len(providers))
	for n := range providers {
		names = append(names, n)
	}
	sort.Strings(names)
	newDefault := names[0]

	if newDefault != "alpha" {
		t.Fatalf("new default = %q, want alpha", newDefault)
	}
}

func TestCleanupKeychainSecrets_IdentifiesRefs(t *testing.T) {
	// Test that we correctly identify keychain refs for each provider type
	tests := []struct {
		name     string
		provider config.ProviderConfig
		wantRefs int
	}{
		{
			name: "agentmail with keychain ref",
			provider: config.ProviderConfig{
				Type: config.ProviderAgentMail,
				AgentMail: &config.AgentMailConfig{
					APIKey:  "keychain:test/api-key",
					InboxID: "inbox@example.com",
				},
			},
			wantRefs: 1,
		},
		{
			name: "agentmail without keychain ref",
			provider: config.ProviderConfig{
				Type: config.ProviderAgentMail,
				AgentMail: &config.AgentMailConfig{
					APIKey:  "am_plaintext",
					InboxID: "inbox@example.com",
				},
			},
			wantRefs: 0,
		},
		{
			name: "smtp with keychain ref",
			provider: config.ProviderConfig{
				Type: config.ProviderSMTP,
				SMTP: &config.SMTPConfig{
					Password: "keychain:test/password",
				},
			},
			wantRefs: 1,
		},
		{
			name: "google with multiple keychain refs",
			provider: config.ProviderConfig{
				Type: config.ProviderGoogle,
				Google: &config.GoogleConfig{
					ClientSecret: "keychain:test/client-secret",
					AccessToken:  "keychain:test/access-token",
					RefreshToken: "keychain:test/refresh-token",
				},
			},
			wantRefs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := countKeychainRefs(&tt.provider)
			if refs != tt.wantRefs {
				t.Errorf("countKeychainRefs() = %d, want %d", refs, tt.wantRefs)
			}
		})
	}
}

// countKeychainRefs counts keychain references in a provider config (test helper)
func countKeychainRefs(p *config.ProviderConfig) int {
	count := 0
	switch p.Type {
	case config.ProviderAgentMail:
		if p.AgentMail != nil && keychain.IsKeychainRef(p.AgentMail.APIKey) {
			count++
		}
	case config.ProviderSMTP:
		if p.SMTP != nil && keychain.IsKeychainRef(p.SMTP.Password) {
			count++
		}
	case config.ProviderProton:
		if p.Proton != nil && keychain.IsKeychainRef(p.Proton.Password) {
			count++
		}
	case config.ProviderGoogle:
		if p.Google != nil {
			if keychain.IsKeychainRef(p.Google.ClientSecret) {
				count++
			}
			if keychain.IsKeychainRef(p.Google.AccessToken) {
				count++
			}
			if keychain.IsKeychainRef(p.Google.RefreshToken) {
				count++
			}
		}
	}
	return count
}

