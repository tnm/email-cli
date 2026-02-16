package keychain

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	serviceName = "email-cli"
)

// IsSupported returns true if keychain is supported on this platform.
func IsSupported() bool {
	return runtime.GOOS == "darwin"
}

// Set stores a secret in the keychain.
func Set(account, secret string) error {
	if !IsSupported() {
		return fmt.Errorf("keychain is only supported on macOS")
	}

	// Delete existing entry first (ignore errors if it doesn't exist)
	_ = Delete(account)

	cmd := exec.Command("security", "add-generic-password",
		"-a", account,
		"-s", serviceName,
		"-w", secret,
		"-U", // Update if exists
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to store in keychain: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// Get retrieves a secret from the keychain.
func Get(account string) (string, error) {
	if !IsSupported() {
		return "", fmt.Errorf("keychain is only supported on macOS")
	}

	cmd := exec.Command("security", "find-generic-password",
		"-a", account,
		"-s", serviceName,
		"-w", // Output password only
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("secret not found in keychain for %q", account)
	}

	return strings.TrimSpace(string(output)), nil
}

// Delete removes a secret from the keychain.
func Delete(account string) error {
	if !IsSupported() {
		return fmt.Errorf("keychain is only supported on macOS")
	}

	cmd := exec.Command("security", "delete-generic-password",
		"-a", account,
		"-s", serviceName,
	)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete from keychain: %v", err)
	}
	return nil
}

// KeychainRef returns a keychain reference string for storing in config.
func KeychainRef(provider, field string) string {
	return fmt.Sprintf("keychain:%s/%s", provider, field)
}

// ParseKeychainRef parses a keychain reference and returns the account name.
// Returns empty string if not a keychain reference.
func ParseKeychainRef(value string) string {
	if strings.HasPrefix(value, "keychain:") {
		return strings.TrimPrefix(value, "keychain:")
	}
	return ""
}

// IsKeychainRef returns true if the value is a keychain reference.
func IsKeychainRef(value string) bool {
	return strings.HasPrefix(value, "keychain:")
}

// Resolve resolves a value, fetching from keychain if it's a reference.
func Resolve(value string) (string, error) {
	if !IsKeychainRef(value) {
		return value, nil
	}

	account := ParseKeychainRef(value)
	return Get(account)
}
