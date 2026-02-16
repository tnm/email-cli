package keychain

import "testing"

func TestIsKeychainRef(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"keychain:provider/field", true},
		{"keychain:", true},
		{"keychain:a", true},
		{"plaintext", false},
		{"", false},
		{"KEYCHAIN:provider/field", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := IsKeychainRef(tt.value); got != tt.want {
				t.Errorf("IsKeychainRef(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseKeychainRef(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"keychain:provider/field", "provider/field"},
		{"keychain:test/api-key", "test/api-key"},
		{"keychain:", ""},
		{"plaintext", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := ParseKeychainRef(tt.value); got != tt.want {
				t.Errorf("ParseKeychainRef(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestKeychainRef(t *testing.T) {
	got := KeychainRef("myprovider", "password")
	want := "keychain:myprovider/password"
	if got != want {
		t.Errorf("KeychainRef() = %q, want %q", got, want)
	}
}
