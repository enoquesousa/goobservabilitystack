package tracing

import (
	"net/url"
	"strings"
)

func validateHeaders(name string) error {
	value := setting(name)
	if value == "" {
		return nil
	}
	for _, pair := range strings.Split(value, ",") {
		key, encoded, found := strings.Cut(pair, "=")
		key = strings.TrimSpace(key)
		if !found || !validMetadataKey(key) {
			return invalidSetting(name, "must contain valid gRPC metadata key=value pairs")
		}
		decoded, err := url.PathUnescape(encoded)
		if err != nil {
			return invalidSetting(name, "must contain valid percent-encoded metadata values")
		}
		if !strings.HasSuffix(strings.ToLower(key), "-bin") && !printableMetadata(strings.TrimSpace(decoded)) {
			return invalidSetting(name, "must contain printable ASCII values for non-binary metadata")
		}
	}
	return nil
}

func validMetadataKey(key string) bool {
	if key == "" {
		return false
	}
	for _, character := range key {
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') && !strings.ContainsRune("-_.", character) {
			return false
		}
	}
	return true
}

func printableMetadata(value string) bool {
	for index := range len(value) {
		if value[index] < 0x20 || value[index] > 0x7e {
			return false
		}
	}
	return true
}
