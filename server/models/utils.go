package models

import (
	"strings"
)

func FormatJournalLog(input string) *JournalctlLog {
	parts := strings.Split(input, ":")

	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	switch len(parts) {
	case 0:
		return &JournalctlLog{Type: "", Status: "", Msg: ""}
	case 1:
		return &JournalctlLog{Type: parts[0], Status: "", Msg: ""}
	case 2:
		return &JournalctlLog{Type: parts[0], Status: parts[1], Msg: ""}
	case 3:
		return &JournalctlLog{Type: parts[0], Status: parts[1], Msg: parts[2]}
	default:
		return &JournalctlLog{Type: parts[0], Status: parts[1], Msg: strings.Join(parts[2:], " : ")}
	}
}
