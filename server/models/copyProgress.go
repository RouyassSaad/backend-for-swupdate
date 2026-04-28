package models

import (
	"strconv"

	log "swupdate/bindings/golang/server/log"
	ws "swupdate/bindings/golang/server/websocket"
)

type ProgressWriter struct {
	Id          string
	Total       int64
	Written     int64
	LastPercent int
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {

	n := len(p)
	pw.Written += int64(n)
	if pw.Total > 0 {
		percent := int(float64(pw.Written) * 100 / float64(pw.Total))
		if percent != pw.LastPercent {

			pw.LastPercent = percent
			outgoingMsg := ws.NewOutgoingMessage(
				"copy_progress",
				nil,
				pw.Id,
				strconv.Itoa(pw.LastPercent),
			)
			ws.GlobalHub.Contact(outgoingMsg)

		}
	} else {
		log.Logger.Info("written:", "bytes", pw.Written)
	}
	return n, nil
}
