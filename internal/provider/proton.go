package provider

import (
	"github.com/tnm/email-cli/internal/config"
)

// Proton Mail Bridge exposes a local SMTP server
// Default: 127.0.0.1:1025 (STARTTLS)

type Proton struct {
	smtp *SMTP
}

func NewProton(from string, cfg *config.ProtonConfig) (*Proton, error) {
	// Proton Bridge defaults
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 1025
	}

	smtpCfg := &config.SMTPConfig{
		Host:     host,
		Port:     port,
		Username: cfg.Username,
		Password: cfg.Password,
		UseTLS:   true, // Bridge uses STARTTLS
	}

	smtp, err := NewSMTP(from, smtpCfg)
	if err != nil {
		return nil, err
	}

	return &Proton{smtp: smtp}, nil
}

func (p *Proton) Name() string {
	return "proton"
}

func (p *Proton) Send(email *Email) error {
	return p.smtp.Send(email)
}
