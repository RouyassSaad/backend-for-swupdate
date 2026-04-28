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
