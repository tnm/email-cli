package cmd

import (
	"bytes"
	"errors"
	"strings"
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
			accounts := keychainAccountsToDelete(&tt.provider)
			if len(accounts) == 0 {
				t.Fatal("expected at least one account to delete")
			}
			if accounts[0] != tt.wantAccount {
				t.Errorf("first account = %q, want %q", accounts[0], tt.wantAccount)
			}
		})
	}
}

func TestCleanupKeychainSecretsWithDeleter_DeletesParsedAccounts(t *testing.T) {
	p := &config.ProviderConfig{
		Type: config.ProviderGoogle,
		Google: &config.GoogleConfig{
			ClientSecret: "keychain:a/client-secret",
			AccessToken:  "keychain:b/access-token",
			RefreshToken: "keychain:c/refresh-token",
		},
	}

	var deleted []string
	deleteFn := func(account string) error {
		deleted = append(deleted, account)
		return nil
	}

	var stderr bytes.Buffer
	cleanupKeychainSecretsWithDeleter(p, deleteFn, &stderr)

	want := []string{"a/client-secret", "b/access-token", "c/refresh-token"}
	if len(deleted) != len(want) {
		t.Fatalf("deleted %d accounts, want %d", len(deleted), len(want))
	}
	for i := range want {
		if deleted[i] != want[i] {
			t.Fatalf("deleted[%d] = %q, want %q", i, deleted[i], want[i])
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got: %q", stderr.String())
	}
}

func TestCleanupKeychainSecretsWithDeleter_WarnsOnDeleteError(t *testing.T) {
	p := &config.ProviderConfig{
		Type: config.ProviderSMTP,
		SMTP: &config.SMTPConfig{
			Password: "keychain:acct/password",
		},
	}

	deleteFn := func(account string) error {
		if account != "acct/password" {
			t.Fatalf("delete called with %q, want acct/password", account)
		}
		return errors.New("boom")
	}

	var stderr bytes.Buffer
	cleanupKeychainSecretsWithDeleter(p, deleteFn, &stderr)

	if !strings.Contains(stderr.String(), `failed to remove keychain entry "acct/password": boom`) {
		t.Fatalf("expected warning in stderr, got: %q", stderr.String())
	}
}
