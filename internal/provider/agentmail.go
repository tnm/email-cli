package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tnm/email-cli/internal/config"
)

var agentMailAPIBase = "https://api.agentmail.to/v0"

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

type AgentMail struct {
	apiKey  string
	inboxID string
}

func NewAgentMail(cfg *config.AgentMailConfig) (*AgentMail, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("agentmail api_key is required")
	}
	if cfg.InboxID == "" {
		return nil, fmt.Errorf("agentmail inbox_id is required")
	}
	return &AgentMail{
		apiKey:  cfg.APIKey,
		inboxID: cfg.InboxID,
	}, nil
}

func (a *AgentMail) Name() string {
	return "agentmail"
}

type agentMailAttachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`      // base64 encoded
	ContentType string `json:"content_type"` // MIME type
}

type agentMailRequest struct {
	To          []string              `json:"to,omitempty"`
	Cc          []string              `json:"cc,omitempty"`
	Bcc         []string              `json:"bcc,omitempty"`
	Subject     string                `json:"subject,omitempty"`
	Text        string                `json:"text,omitempty"`
	HTML        string                `json:"html,omitempty"`
	Attachments []agentMailAttachment `json:"attachments,omitempty"`
}

type agentMailError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (a *AgentMail) Send(email *Email) error {
	req := agentMailRequest{
		To:      email.To,
		Cc:      email.Cc,
		Bcc:     email.Bcc,
		Subject: email.Subject,
	}

	// Set body - provide both text and html for best deliverability
	if email.HTML {
		req.HTML = email.Body
		// Strip HTML for text version (basic)
		req.Text = email.Body // AgentMail recommends both, but we'll let them handle it
	} else {
		req.Text = email.Body
	}

	// Handle attachments
	for _, att := range email.Attachments {
		var content []byte
		var err error

		if att.Path != "" {
			content, err = os.ReadFile(att.Path)
			if err != nil {
				return fmt.Errorf("failed to read attachment %s: %w", att.Path, err)
			}
		} else {
			content = att.Content
		}

		filename := att.Filename
		if filename == "" && att.Path != "" {
			filename = filepath.Base(att.Path)
		}
		if filename == "" {
			filename = "attachment"
		}

		mimeType := mime.TypeByExtension(filepath.Ext(filename))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		req.Attachments = append(req.Attachments, agentMailAttachment{
			Filename:    filename,
			Content:     base64.StdEncoding.EncodeToString(content),
			ContentType: mimeType,
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/inboxes/%s/messages/send", agentMailAPIBase, a.inboxID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		var apiErr agentMailError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("agentmail error: %s", apiErr.Message)
		}
		return fmt.Errorf("agentmail error: %s (status %d)", string(respBody), resp.StatusCode)
	}

	return nil
}
