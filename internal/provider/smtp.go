package provider

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tnm/email-cli/internal/config"
)

func generateBoundary() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("boundary-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("boundary-%x", b)
}

type SMTP struct {
	from   string
	config *config.SMTPConfig
}

func NewSMTP(from string, cfg *config.SMTPConfig) (*SMTP, error) {
	return &SMTP{
		from:   from,
		config: cfg,
	}, nil
}

func (s *SMTP) Name() string {
	return "smtp"
}

func (s *SMTP) Send(email *Email) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	mailFrom := sanitizeHeaderValue(s.from)

	// Build message
	msg, err := s.buildMessage(email)
	if err != nil {
		return err
	}

	// Collect all recipients
	recipients := make([]string, 0, len(email.To)+len(email.Cc)+len(email.Bcc))
	recipients = append(recipients, sanitizeAddressList(email.To)...)
	recipients = append(recipients, sanitizeAddressList(email.Cc)...)
	recipients = append(recipients, sanitizeAddressList(email.Bcc)...)

	if len(recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	var auth smtp.Auth
	if s.config.Username != "" || s.config.Password != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	if s.config.UseTLS {
		return s.sendTLS(addr, mailFrom, auth, recipients, msg)
	}

	return smtp.SendMail(addr, auth, mailFrom, recipients, msg)
}

func (s *SMTP) sendTLS(addr, mailFrom string, auth smtp.Auth, recipients []string, msg []byte) error {
	// Connect
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.config.Host,
	})
	if err != nil {
		// Try STARTTLS instead
		return s.sendSTARTTLS(addr, mailFrom, auth, recipients, msg)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	if err := client.Mail(mailFrom); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}

	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt to failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	return client.Quit()
}

func (s *SMTP) sendSTARTTLS(addr, mailFrom string, auth smtp.Auth, recipients []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer client.Close()

	if err := client.StartTLS(&tls.Config{ServerName: s.config.Host}); err != nil {
		return fmt.Errorf("starttls failed: %w", err)
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	if err := client.Mail(mailFrom); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}

	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt to failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	return client.Quit()
}

func (s *SMTP) buildMessage(email *Email) ([]byte, error) {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", sanitizeHeaderValue(s.from)))
	to := sanitizeAddressList(email.To)
	if len(to) > 0 {
		msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	}
	cc := sanitizeAddressList(email.Cc)
	if len(cc) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", sanitizeHeaderValue(email.Subject)))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(email.Attachments) > 0 {
		boundary := generateBoundary()
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
		msg.WriteString("\r\n")

		// Body part
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		contentType := "text/plain"
		if email.HTML {
			contentType = "text/html"
		}
		msg.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"\r\n", contentType))
		msg.WriteString("\r\n")
		msg.WriteString(email.Body)
		msg.WriteString("\r\n")

		// Attachments
		for _, att := range email.Attachments {
			content := att.Content
			if content == nil && att.Path != "" {
				data, err := os.ReadFile(att.Path)
				if err != nil {
					return nil, fmt.Errorf("failed to read attachment %s: %w", att.Path, err)
				}
				content = data
			}

				filename := att.Filename
				if filename == "" && att.Path != "" {
					filename = filepath.Base(att.Path)
				}
				filename = sanitizeFilename(filename)

				mimeType := mime.TypeByExtension(filepath.Ext(filename))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}

			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", mimeType))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
			msg.WriteString("\r\n")

			encoded := base64.StdEncoding.EncodeToString(content)
			// Wrap at 76 chars
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				msg.WriteString(encoded[i:end])
				msg.WriteString("\r\n")
			}
		}

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		contentType := "text/plain"
		if email.HTML {
			contentType = "text/html"
		}
		msg.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"\r\n", contentType))
		msg.WriteString("\r\n")
		msg.WriteString(email.Body)
	}

	return []byte(msg.String()), nil
}
