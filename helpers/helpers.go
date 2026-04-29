package helpers

import (
	"time"

	"strings"
	"unicode"

	"github.com/google/uuid"

	"golang.org/x/text/unicode/norm"
)

func CurrentTime() string {
	t := time.Now()
	return t.Format("2006-01-02 15:04:05")
}

// This will be used to bind each connection
func GenerateUUID() string {
	return uuid.NewString()
}

func CurrentTimestampISO() string {
	return time.Now().Format(time.RFC3339)
}

// to ckeck if a connection obj should be deleted or not
func ShouldItLive(ttl string, createdAt string) bool {
	return true
}

func SanitizeFilename(filename string) string {
	// Normalize Unicode (NFC)
	filename = norm.NFC.String(filename)

	sanitized := make([]rune, 0, len(filename))

	for _, r := range filename {
		// Remove control chars, DEL, slash, non-printable
		if r <= 31 || r == 127 || r == '/' || !unicode.IsPrint(r) {
			continue
		}

		//shell-dangerous characters
		switch r {
		case '*', '?', '[', ']', '{', '}', '$', '!', '\'', '"', '\\', '`', '|', '<', '>', '&':
			continue
		}

		sanitized = append(sanitized, r)
	}

	if len(sanitized) == 0 {
		return "unnamed"
	}

	out := strings.TrimSpace(string(sanitized))

	if out == "." || out == ".." || out == "" {
		return "unnamed"
	}

	return out
}
