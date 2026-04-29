package systemmd

import (
	log "swupdate/bindings/golang/server/log"
	util "swupdate/bindings/golang/server/models"
	ws "swupdate/bindings/golang/server/websocket"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal" //this requires "libsystemd-dev" to run(sudo apt install libsystemd-dev)
)

func GetSystemMdLogs(journalctlChan chan *string) {

	journal, err := sdjournal.NewJournal()
	if err != nil {
		//log.Fatalf("Failed to open journal: %v", err)
		log.Logger.Error("Failed to open journal", "error", err)
		return
	}
	defer journal.Close()

	err = journal.AddMatch("_SYSTEMD_UNIT=swupdate.service")
	if err != nil {
		//log.Fatalf("Failed to add match: %v", err)
		log.Logger.Error("Failed to add match:", "error", err)
		return
	}

	// Seek to the end (like journalctl -f)
	err = journal.SeekTail()
	if err != nil {
		//log.Fatalf("Failed to seek to tail: %v", err)
		log.Logger.Error("Failed to seek to tail", "error", err)
		return
	}

	//one step back so we don't miss the last entry
	_, err = journal.Previous()
	if err != nil {
		//log.Fatalf("Failed to move back: %v", err)
		log.Logger.Error("Failed to move back", "error", err)
		return
	}

	//fmt.Println("Streaming logs from swupdate.service...")
	log.Logger.Info("Streaming logs from swupdate.service...")
	var currentID *string
	for {

		// Wait for new log entries (timeout 1s)
		n := journal.Wait(1 * time.Second)

		select {
		case currentID = <-journalctlChan:
			log.Logger.Info("New client subscribed", "id", *currentID)
		default:
			// no new client
		}

		if currentID == nil {
			continue // no client connected yet
		}
		if n == sdjournal.SD_JOURNAL_APPEND {
			// Read all new entries
			for {
				r, err := journal.Next()
				if err != nil {
					//log.Printf("Next error: %v", err)
					log.Logger.Error("Next error", "error", err)
					break
				}
				if r == 0 {
					break // no more entries
				}

				entry, err := journal.GetEntry()
				if err != nil {
					//log.Printf("GetEntry error: %v", err)
					log.Logger.Error("GetEntry error", "error", err)
					continue
				}

				// Print the MESSAGE field
				msg := entry.Fields["MESSAGE"]

				formattedMsg := util.FormatJournalLog(msg)

				log := ws.JournalctlLogMessage{Type: formattedMsg.Type, Status: formattedMsg.Status,
					Msg: formattedMsg.Msg}

				outgoingMsg := ws.NewOutgoingMessage(
					"journalctl",
					nil,
					*currentID,
					log,
				)
				ws.GlobalHub.Contact(outgoingMsg)

				//log.Logger.Info("journal log", "log", formattedMsg)

			}
		}
	}
}
