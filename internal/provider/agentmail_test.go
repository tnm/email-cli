package provider

import (
	"testing"
	"time"
)

func TestAgentMailClientHasTimeout(t *testing.T) {
	if httpClient.Timeout == 0 {
		t.Fatal("httpClient has no timeout set")
	}
	if httpClient.Timeout != 30*time.Second {
		t.Fatalf("httpClient.Timeout = %v, want 30s", httpClient.Timeout)
	}
}
