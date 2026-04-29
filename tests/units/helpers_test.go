package tests

import (
	"regexp"
	"testing"
	"time"

	helpers "swupdate/bindings/golang/helpers"
)

func TestCurrentTime(t *testing.T) {
	ts := helpers.CurrentTime()

	// Expected format: 2006-01-02 15:04:05
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)
	if !re.MatchString(ts) {
		t.Fatalf("CurrentTime() returned invalid format: %s", ts)
	}
}

func TestGenerateUUID(t *testing.T) {
	id := helpers.GenerateUUID()

	if id == "" {
		t.Fatalf("GenerateUUID() returned empty string")
	}

	// UUID v4 regex
	re := regexp.MustCompile(`^[a-f0-9-]{36}$`)
	if !re.MatchString(id) {
		t.Fatalf("GenerateUUID() returned invalid UUID: %s", id)
	}
}

func TestCurrentTimestampISO(t *testing.T) {
	ts := helpers.CurrentTimestampISO()

	_, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatalf("CurrentTimestampISO() returned invalid RFC3339 timestamp: %s", ts)
	}
}

func TestSanitizeLinuxAndUnix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove slash",
			input:    "my/file.txt",
			expected: "myfile.txt",
		},
		{
			name:     "Remove control characters",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "Trim whitespace",
			input:    "   hello.txt   ",
			expected: "hello.txt",
		},
		{
			name:     "Reserved dot",
			input:    ".",
			expected: "unnamed",
		},
		{
			name:     "Reserved double dot",
			input:    "..",
			expected: "unnamed",
		},
		{
			name:     "Unicode allowed",
			input:    "مرحبا.txt",
			expected: "مرحبا.txt",
		},
		{
			name:     "Shell-dangerous characters removed",
			input:    "file$name?.txt",
			expected: "filename.txt",
		},
		{
			name:     "Empty after sanitization",
			input:    "$$$",
			expected: "unnamed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helpers.SanitizeFilename(tc.input)
			if got != tc.expected {
				t.Errorf("sanitizeLinuxAndUnix(%q) = %q; expected %q",
					tc.input, got, tc.expected)
			}
		})
	}

}
