package provider

import (
	"strings"
	"testing"
)

func TestSMTPBuildMessage_SanitizesHeaderInjection(t *testing.T) {
	s := &SMTP{from: "sender@example.com"}

	email := &Email{
		To:      []string{"to@example.com\r\nBcc:bad@example.com"},
		Cc:      []string{"cc@example.com\n"},
		Subject: "Hello\r\nX-Injected: true",
		Body:    "message body",
	}

	msgBytes, err := s.buildMessage(email)
	if err != nil {
		t.Fatalf("buildMessage() error = %v", err)
	}
	msg := string(msgBytes)

	if strings.Contains(msg, "\r\nBcc:") {
		t.Fatalf("message contains injected Bcc header:\n%s", msg)
	}
	if strings.Contains(msg, "\r\nX-Injected:") {
		t.Fatalf("message contains injected custom header:\n%s", msg)
	}
	if !strings.Contains(msg, "From: sender@example.com\r\n") {
		t.Fatalf("message missing expected From header:\n%s", msg)
	}
}

func TestSMTPBuildMessage_SanitizesAttachmentFilename(t *testing.T) {
	s := &SMTP{from: "sender@example.com"}

	email := &Email{
		To:      []string{"to@example.com"},
		Subject: "Attachment test",
		Body:    "body",
		Attachments: []Attachment{
			{
				Filename: "bad\"\r\nX-Test:1.txt",
				Content:  []byte("abc"),
			},
		},
	}

	msgBytes, err := s.buildMessage(email)
	if err != nil {
		t.Fatalf("buildMessage() error = %v", err)
	}
	msg := string(msgBytes)

	if strings.Contains(msg, "\r\nX-Test:1.txt") {
		t.Fatalf("message contains injected attachment header:\n%s", msg)
	}
	if !strings.Contains(msg, "filename=\"badX-Test:1.txt\"") {
		t.Fatalf("message did not contain sanitized filename:\n%s", msg)
	}
}

