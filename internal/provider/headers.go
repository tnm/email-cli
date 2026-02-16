package provider

import "strings"

func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return strings.TrimSpace(value)
}

func sanitizeAddressList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		safe := sanitizeHeaderValue(value)
		if safe != "" {
			out = append(out, safe)
		}
	}
	return out
}

func sanitizeFilename(value string) string {
	value = sanitizeHeaderValue(value)
	value = strings.ReplaceAll(value, "\"", "")
	if value == "" {
		return "attachment"
	}
	return value
}

