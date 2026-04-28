package models

// this is written to when we upload a file through a post method
// then LaunchUpdate() reads it to start and update (view example.go)
type UpdateChanel struct {
	ConnexionId string
	Filename    string
}

// journalctl logs format
type JournalctlLog struct {
	Type   string
	Status string
	Msg    string
}
