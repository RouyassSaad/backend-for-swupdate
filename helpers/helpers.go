package helpers

import (
	"time"

	"github.com/google/uuid"
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
