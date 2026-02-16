package provider

import (
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
