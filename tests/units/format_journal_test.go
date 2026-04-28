package tests

import (
	utils "swupdate/bindings/golang/server/models"
	"testing"
)

func TestFormatJournalLog(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantT string
		wantS string
		wantM string
	}{
		{
			name:  "empty string",
			input: "",
			wantT: "",
			wantS: "",
			wantM: "",
		},
		{
			name:  "one part",
			input: "INFO",
			wantT: "INFO",
			wantS: "",
			wantM: "",
		},
		{
			name:  "two parts",
			input: "INFO: OK",
			wantT: "INFO",
			wantS: "OK",
			wantM: "",
		},
		{
			name:  "three parts",
			input: "ERROR: FAILED: Disk full",
			wantT: "ERROR",
			wantS: "FAILED",
			wantM: "Disk full",
		},
		{
			name:  "four parts (default case)",
			input: "WARN: TIMEOUT: retrying: attempt 3",
			wantT: "WARN",
			wantS: "TIMEOUT",
			wantM: "retrying : attempt 3",
		},
		{
			name:  "trims whitespace",
			input: "  INFO  :   OK   :   done   ",
			wantT: "INFO",
			wantS: "OK",
			wantM: "done",
		},
		{
			name:  "multiple colons in message",
			input: "DEBUG: STATE: part1:part2:part3",
			wantT: "DEBUG",
			wantS: "STATE",
			wantM: "part1 : part2 : part3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.FormatJournalLog(tc.input)

			if got.Type != tc.wantT {
				t.Errorf("Type: expected %q, got %q", tc.wantT, got.Type)
			}
			if got.Status != tc.wantS {
				t.Errorf("Status: expected %q, got %q", tc.wantS, got.Status)
			}
			if got.Msg != tc.wantM {
				t.Errorf("Msg: expected %q, got %q", tc.wantM, got.Msg)
			}
		})
	}
}
