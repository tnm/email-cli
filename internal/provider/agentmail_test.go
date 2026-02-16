package provider

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/tnm/email-cli/internal/config"
)

func TestAgentMailClientHasTimeout(t *testing.T) {
	if httpClient.Timeout == 0 {
		t.Fatal("httpClient has no timeout set")
	}
	if httpClient.Timeout != 30*time.Second {
		t.Fatalf("httpClient.Timeout = %v, want 30s", httpClient.Timeout)
	}
}

func TestAgentMailSend_Timeout(t *testing.T) {
	originalBase := agentMailAPIBase
	originalTimeout := httpClient.Timeout
	agentMailAPIBase = ""
	httpClient.Timeout = 100 * time.Millisecond
	defer func() {
		agentMailAPIBase = originalBase
		httpClient.Timeout = originalTimeout
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	agentMailAPIBase = server.URL
	a, err := NewAgentMail(&config.AgentMailConfig{
		APIKey:  "am_test",
		InboxID: "test@agentmail.to",
	})
	if err != nil {
		t.Fatalf("NewAgentMail() error = %v", err)
	}
	err = a.Send(&Email{
		To:      []string{"user@example.com"},
		Subject: "test",
		Body:    "body",
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	var netErr *url.Error
	if !errors.As(err, &netErr) {
		t.Fatalf("expected url.Error, got: %T (%v)", err, err)
	}
	if !netErr.Timeout() {
		t.Fatalf("expected timeout url.Error, got: %v", err)
	}
}

func TestAgentMailSend_HTTPError(t *testing.T) {
	originalBase := agentMailAPIBase
	defer func() { agentMailAPIBase = originalBase }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/inboxes/test@agentmail.to/messages/send" {
			t.Fatalf("path = %q, want /inboxes/test@agentmail.to/messages/send", got)
		}
		if got := r.Method; got != http.MethodPost {
			t.Fatalf("method = %q, want POST", got)
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized","message":"Invalid API key"}`))
	}))
	defer server.Close()

	agentMailAPIBase = server.URL
	a, err := NewAgentMail(&config.AgentMailConfig{
		APIKey:  "am_test",
		InboxID: "test@agentmail.to",
	})
	if err != nil {
		t.Fatalf("NewAgentMail() error = %v", err)
	}

	err = a.Send(&Email{
		To:      []string{"user@example.com"},
		Subject: "test",
		Body:    "body",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	want := "agentmail error: Invalid API key"
	if got := err.Error(); got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestAgentMailSend_Success(t *testing.T) {
	originalBase := agentMailAPIBase
	defer func() { agentMailAPIBase = originalBase }()

	var authHeader string
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	agentMailAPIBase = server.URL
	a, err := NewAgentMail(&config.AgentMailConfig{
		APIKey:  "am_test",
		InboxID: "test@agentmail.to",
	})
	if err != nil {
		t.Fatalf("NewAgentMail() error = %v", err)
	}

	err = a.Send(&Email{
		To:      []string{"user@example.com"},
		Subject: "test",
		Body:    "body",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if authHeader != "Bearer am_test" {
		t.Fatalf("Authorization header = %q, want Bearer am_test", authHeader)
	}
	wantPath := "/inboxes/test@agentmail.to/messages/send"
	if gotPath != wantPath {
		t.Fatalf("path = %q, want %q", gotPath, wantPath)
	}
}

func TestNewAgentMail_RequiresAPIKey(t *testing.T) {
	_, err := NewAgentMail(&config.AgentMailConfig{
		APIKey:  "",
		InboxID: "test@agentmail.to",
	})
	if err == nil {
		t.Fatal("NewAgentMail() should fail without API key")
	}
}

func TestNewAgentMail_RequiresInboxID(t *testing.T) {
	_, err := NewAgentMail(&config.AgentMailConfig{
		APIKey:  "am_test",
		InboxID: "",
	})
	if err == nil {
		t.Fatal("NewAgentMail() should fail without inbox ID")
	}
}

func TestNewAgentMail_Success(t *testing.T) {
	am, err := NewAgentMail(&config.AgentMailConfig{
		APIKey:  "am_test",
		InboxID: "test@agentmail.to",
	})
	if err != nil {
		t.Fatalf("NewAgentMail() error = %v", err)
	}
	if am.Name() != "agentmail" {
		t.Errorf("Name() = %q, want 'agentmail'", am.Name())
	}
}

func TestNewSMTP_Success(t *testing.T) {
	smtp, err := NewSMTP("test@example.com", &config.SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		UseTLS:   true,
	})
	if err != nil {
		t.Fatalf("NewSMTP() error = %v", err)
	}
	if smtp.Name() != "smtp" {
		t.Errorf("Name() = %q, want 'smtp'", smtp.Name())
	}
}

func TestNewProton_Success(t *testing.T) {
	proton, err := NewProton("test@proton.me", &config.ProtonConfig{
		Host:     "127.0.0.1",
		Port:     1025,
		Username: "user",
		Password: "pass",
	})
	if err != nil {
		t.Fatalf("NewProton() error = %v", err)
	}
	if proton.Name() != "proton" {
		t.Errorf("Name() = %q, want 'proton'", proton.Name())
	}
}
