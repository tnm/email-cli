package provider

import (
	"encoding/hex"
	"net/url"
	"strings"
	"testing"
)

func TestGenerateGoogleOAuthState_Format(t *testing.T) {
	state, err := GenerateGoogleOAuthState()
	if err != nil {
		t.Fatalf("GenerateGoogleOAuthState() error = %v", err)
	}
	if len(state) != 32 {
		t.Fatalf("state length = %d, want 32", len(state))
	}
	if _, err := hex.DecodeString(state); err != nil {
		t.Fatalf("state is not valid hex: %v", err)
	}
}

func TestGetGoogleAuthURL_ContainsStateAndRedirect(t *testing.T) {
	u := GetGoogleAuthURL("client-id", "client-secret", "state123")
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	if got := parsed.Query().Get("state"); got != "state123" {
		t.Fatalf("state = %q, want state123", got)
	}
	if got := parsed.Query().Get("redirect_uri"); got != googleRedirectURL {
		t.Fatalf("redirect_uri = %q, want %q", got, googleRedirectURL)
	}
	if !strings.Contains(parsed.Host, "google") {
		t.Fatalf("unexpected auth host: %q", parsed.Host)
	}
}

func TestRunGoogleAuthServer_RequiresState(t *testing.T) {
	_, err := RunGoogleAuthServer("")
	if err == nil {
		t.Fatalf("RunGoogleAuthServer() expected error for empty state")
	}
}

