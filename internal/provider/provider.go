package provider

import (
	"fmt"

	"github.com/tnm/email-cli/internal/config"
)

type Email struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	HTML        bool
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Path     string
	Content  []byte
}

type Provider interface {
	Send(email *Email) error
	Name() string
}

func New(cfg *config.ProviderConfig) (Provider, error) {
	switch cfg.Type {
	case config.ProviderGoogle:
		if cfg.Google == nil {
			return nil, fmt.Errorf("google config missing")
		}
		return NewGoogle(cfg.From, cfg.Google)
	case config.ProviderProton:
		if cfg.Proton == nil {
			return nil, fmt.Errorf("proton config missing")
		}
		return NewProton(cfg.From, cfg.Proton)
	case config.ProviderSMTP:
		if cfg.SMTP == nil {
			return nil, fmt.Errorf("smtp config missing")
		}
		return NewSMTP(cfg.From, cfg.SMTP)
	case config.ProviderAgentMail:
		if cfg.AgentMail == nil {
			return nil, fmt.Errorf("agentmail config missing")
		}
		return NewAgentMail(cfg.AgentMail)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}
}
