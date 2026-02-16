package provider

import "testing"

func TestSanitizeHeaderValue_StripsCRLFAndTrim(t *testing.T) {
	got := sanitizeHeaderValue("  hello\r\nworld\n  ")
	want := "helloworld"
	if got != want {
		t.Fatalf("sanitizeHeaderValue() = %q, want %q", got, want)
	}
}

func TestSanitizeAddressList_DropsEmptyAfterSanitize(t *testing.T) {
	got := sanitizeAddressList([]string{"a@example.com", "\n", " b@example.com\r "})
	if len(got) != 2 {
		t.Fatalf("sanitizeAddressList() len = %d, want 2", len(got))
	}
	if got[0] != "a@example.com" {
		t.Fatalf("sanitizeAddressList()[0] = %q, want a@example.com", got[0])
	}
	if got[1] != "b@example.com" {
		t.Fatalf("sanitizeAddressList()[1] = %q, want b@example.com", got[1])
	}
}

func TestSanitizeFilename_RemovesQuotesAndFallback(t *testing.T) {
	got := sanitizeFilename("\"report\r\n.csv\"")
	want := "report.csv"
	if got != want {
		t.Fatalf("sanitizeFilename() = %q, want %q", got, want)
	}

	empty := sanitizeFilename(" \r\n ")
	if empty != "attachment" {
		t.Fatalf("sanitizeFilename(empty) = %q, want attachment", empty)
	}
}

