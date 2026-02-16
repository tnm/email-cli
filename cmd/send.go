package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/tnm/email-cli/internal/config"
	"github.com/tnm/email-cli/internal/provider"
	"github.com/urfave/cli/v2"
)

func sendCommand() *cli.Command {
	return &cli.Command{
		Name:  "send",
		Usage: "Send an email",
		Description: "Send an email using the configured provider.\n\n" +
			"Examples:\n" +
			"  # Send a simple email\n" +
			"  email-cli send --to user@example.com --subject \"Hello\" --body \"Hi there!\"\n\n" +
			"  # Send with attachments\n" +
			"  email-cli send --to user@example.com --subject \"Report\" --body \"See attached\" --attach report.pdf\n\n" +
			"  # Send HTML email\n" +
			"  email-cli send --to user@example.com --subject \"Newsletter\" --body \"<h1>Hello</h1>\" --html\n\n" +
			"  # Read body from stdin\n" +
			"  echo \"Hello world\" | email-cli send --to user@example.com --subject \"Test\"\n\n" +
			"  # Use specific provider\n" +
			"  email-cli send --provider google --to user@example.com --subject \"Via Gmail\" --body \"Sent via Google\"",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{Name: "to", Aliases: []string{"t"}, Usage: "Recipient email addresses (repeatable)"},
			&cli.StringSliceFlag{Name: "cc", Aliases: []string{"c"}, Usage: "CC recipients"},
			&cli.StringSliceFlag{Name: "bcc", Aliases: []string{"b"}, Usage: "BCC recipients"},
			&cli.StringFlag{Name: "subject", Aliases: []string{"s"}, Usage: "Email subject"},
			&cli.StringFlag{Name: "body", Aliases: []string{"m"}, Usage: "Email body (reads from stdin if not provided)"},
			&cli.BoolFlag{Name: "html", Usage: "Treat body as HTML"},
			&cli.StringSliceFlag{Name: "attach", Aliases: []string{"a"}, Usage: "File attachments (repeatable)"},
			&cli.StringFlag{Name: "provider", Aliases: []string{"p"}, Usage: "Provider to use (default: configured default)"},
		},
		Action: runSend,
	}
}

func runSend(c *cli.Context) error {
	sendTo := c.StringSlice("to")
	sendCc := c.StringSlice("cc")
	sendBcc := c.StringSlice("bcc")
	sendSubject := c.String("subject")
	sendBody := c.String("body")
	sendHTML := c.Bool("html")
	sendAttachments := c.StringSlice("attach")
	sendProvider := c.String("provider")

	if len(sendTo) == 0 {
		return fmt.Errorf("--to is required")
	}
	if sendSubject == "" {
		return fmt.Errorf("--subject is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	providerCfg, err := cfg.GetProvider(sendProvider)
	if err != nil {
		return err
	}

	p, err := provider.New(providerCfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Read body from stdin if not provided.
	body := sendBody
	if body == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			body = string(data)
		}
	}
	if body == "" {
		return fmt.Errorf("email body is required (use --body or pipe via stdin)")
	}

	attachments := make([]provider.Attachment, 0, len(sendAttachments))
	for _, path := range sendAttachments {
		attachments = append(attachments, provider.Attachment{Path: path})
	}

	email := &provider.Email{
		To:          sendTo,
		Cc:          sendCc,
		Bcc:         sendBcc,
		Subject:     sendSubject,
		Body:        body,
		HTML:        sendHTML,
		Attachments: attachments,
	}

	if err := p.Send(email); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "Email sent successfully via %s\n", p.Name())
	return nil
}

