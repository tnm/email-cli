package provider

import (
	"testing"

	"github.com/tnm/email-cli/internal/config"
)

func TestNew_SMTP(t *testing.T) {
	cfg := &config.ProviderConfig{
		Type: config.ProviderSMTP,
		From: "test@example.com",
		SMTP: &config.SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "pass",
			UseTLS:   true,
		},
	}

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if p.Name() != "smtp" {
		t.Errorf("Name() = %q, want 'smtp'", p.Name())
	}
}

func TestNew_Proton(t *testing.T) {
	cfg := &config.ProviderConfig{
		Type: config.ProviderProton,
		From: "test@proton.me",
		Proton: &config.ProtonConfig{
			Host:     "127.0.0.1",
			Port:     1025,
			Username: "user",
			Password: "pass",
		},
	}

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if p.Name() != "proton" {
		t.Errorf("Name() = %q, want 'proton'", p.Name())
	}
}

func TestNew_AgentMail(t *testing.T) {
	cfg := &config.ProviderConfig{
		Type: config.ProviderAgentMail,
		AgentMail: &config.AgentMailConfig{
			APIKey:  "am_test",
			InboxID: "test@agentmail.to",
		},
	}

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if p.Name() != "agentmail" {
		t.Errorf("Name() = %q, want 'agentmail'", p.Name())
	}
}

func TestNew_MissingConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.ProviderConfig
	}{
		{
			name: "smtp missing config",
			cfg:  config.ProviderConfig{Type: config.ProviderSMTP, From: "test@example.com"},
		},
		{
			name: "proton missing config",
			cfg:  config.ProviderConfig{Type: config.ProviderProton, From: "test@proton.me"},
		},
		{
			name: "google missing config",
			cfg:  config.ProviderConfig{Type: config.ProviderGoogle, From: "test@gmail.com"},
		},
		{
			name: "agentmail missing config",
			cfg:  config.ProviderConfig{Type: config.ProviderAgentMail},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&tt.cfg)
			if err == nil {
				t.Error("New() should fail with missing config")
			}
		})
	}
}

func TestNew_UnknownType(t *testing.T) {
	cfg := &config.ProviderConfig{
		Type: "unknown",
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("New() should fail with unknown provider type")
	}
}
