package websocket

import (
	helper "swupdate/bindings/golang/helpers"
)

func NewSwupdateMessage(time string, status int, image string, handler string, info string, total_bytes int,
	total_steps int, downloaded_bytes int, current_step int, current_percent int) SwupdateMessage {
	return SwupdateMessage{
		Time:            time,
		Status:          status,
		Image:           image,
		Handler:         handler,
		Info:            info,
		TotalBytes:      total_bytes,
		TotalSteps:      total_steps,
		DownloadedBytes: downloaded_bytes,
		CurrentStep:     current_step,
		CurrentPercent:  current_percent,
	}
}

func NewOutgoingMessage(msgType string, msgError error, id string, data any) OutgoingMessage {
	return OutgoingMessage{
		Type:      msgType,
		TimeStamp: helper.CurrentTimestampISO(),
		Error:     msgError,
		Id:        id,
		Data:      data,
	}
}
