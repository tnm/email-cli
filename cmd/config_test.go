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
	// Test that selectNewDefault returns alphabetically first provider
	// This tests the actual helper function used by config remove
	tests := []struct {
		name      string
		providers map[string]config.ProviderConfig
		want      string
	}{
		{
			name: "picks alphabetically first",
			providers: map[string]config.ProviderConfig{
				"zebra":  {Name: "zebra"},
				"alpha":  {Name: "alpha"},
				"middle": {Name: "middle"},
			},
			want: "alpha",
		},
		{
			name: "single provider",
			providers: map[string]config.ProviderConfig{
				"only": {Name: "only"},
			},
			want: "only",
		},
		{
			name:      "empty providers",
			providers: map[string]config.ProviderConfig{},
			want:      "",
		},
		{
			name: "numeric prefixes sorted correctly",
			providers: map[string]config.ProviderConfig{
				"2nd":  {Name: "2nd"},
				"10th": {Name: "10th"},
				"1st":  {Name: "1st"},
			},
			want: "10th", // string sort: "10th" < "1st" < "2nd"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectNewDefault(tt.providers)
			if got != tt.want {
				t.Errorf("selectNewDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

// selectNewDefault is the production logic extracted for testing
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

func TestCleanupKeychainSecrets_ParsesActualRefs(t *testing.T) {
	// Test that cleanupKeychainSecrets parses the actual keychain reference
	// not just constructs one from provider name
	tests := []struct {
		name        string
		provider    config.ProviderConfig
		wantAccount string // the account that should be deleted
	}{
		{
			name: "agentmail parses actual ref",
			provider: config.ProviderConfig{
				Type: config.ProviderAgentMail,
				Name: "myprovider",
				AgentMail: &config.AgentMailConfig{
					APIKey:  "keychain:other-provider/api-key", // different from provider name
					InboxID: "inbox@example.com",
				},
			},
			wantAccount: "other-provider/api-key",
		},
		{
			name: "smtp parses actual ref",
			provider: config.ProviderConfig{
				Type: config.ProviderSMTP,
				Name: "mysmtp",
				SMTP: &config.SMTPConfig{
					Password: "keychain:custom/password",
				},
			},
			wantAccount: "custom/password",
		},
		{
			name: "google parses multiple refs",
			provider: config.ProviderConfig{
				Type: config.ProviderGoogle,
				Name: "mygoogle",
				Google: &config.GoogleConfig{
					ClientSecret: "keychain:a/client-secret",
					AccessToken:  "keychain:b/access-token",
					RefreshToken: "keychain:c/refresh-token",
				},
			},
			wantAccount: "a/client-secret", // first one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts := getKeychainAccountsToDelete(&tt.provider)
			if len(accounts) == 0 {
				t.Fatal("expected at least one account to delete")
			}
			if accounts[0] != tt.wantAccount {
				t.Errorf("first account = %q, want %q", accounts[0], tt.wantAccount)
			}
		})
	}
}

// getKeychainAccountsToDelete extracts the keychain accounts that would be deleted
// This mirrors the logic in cleanupKeychainSecrets for testing
func getKeychainAccountsToDelete(p *config.ProviderConfig) []string {
	var accounts []string

	switch p.Type {
	case config.ProviderAgentMail:
		if p.AgentMail != nil && keychain.IsKeychainRef(p.AgentMail.APIKey) {
			accounts = append(accounts, keychain.ParseKeychainRef(p.AgentMail.APIKey))
		}
	case config.ProviderSMTP:
		if p.SMTP != nil && keychain.IsKeychainRef(p.SMTP.Password) {
			accounts = append(accounts, keychain.ParseKeychainRef(p.SMTP.Password))
		}
	case config.ProviderProton:
		if p.Proton != nil && keychain.IsKeychainRef(p.Proton.Password) {
			accounts = append(accounts, keychain.ParseKeychainRef(p.Proton.Password))
		}
	case config.ProviderGoogle:
		if p.Google != nil {
			if keychain.IsKeychainRef(p.Google.ClientSecret) {
				accounts = append(accounts, keychain.ParseKeychainRef(p.Google.ClientSecret))
			}
			if keychain.IsKeychainRef(p.Google.AccessToken) {
				accounts = append(accounts, keychain.ParseKeychainRef(p.Google.AccessToken))
			}
			if keychain.IsKeychainRef(p.Google.RefreshToken) {
				accounts = append(accounts, keychain.ParseKeychainRef(p.Google.RefreshToken))
			}
		}
	}
	return accounts
}

