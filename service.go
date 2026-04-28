package swupdate

/*
#cgo LDFLAGS: -lswupdate

#include "cgo.h"
*/
import "C"

import (
	helpers "swupdate/bindings/golang/helpers"
	ws "swupdate/bindings/golang/server/websocket"
)

func ProgressCallbackClosure(connectionId string) func(ProgressMsg) {

	return func(progress ProgressMsg) {
		val := C.cast_union_wrapper_progress_msg_as_member_msg(&progress)
		//C.GoString(&my_char_arr[0])
		image := C.GoString(&val.cur_image[0])
		handler := C.GoString(&val.hnd_name[0])
		info := GoStringFromSlice(val.info[:val.infolen])

		downloadedData := val.dwl_percent
		totalBytesToDownload := val.dwl_bytes
		totalSteps := val.nsteps
		currentStep := val.cur_step
		currPercent := val.cur_percent

		//log.Printf("Update Status: %d, image: %v handler: %v, info: %v\n", val.status, image, handler, info)
		//log.Printf("Total Bytes To Download: %d", totalBytesToDownload)
		//log.Printf("Downloaded Bytes: %d", downloadedData)
		//log.Printf("Total Steps: %d", totalSteps)
		//log.Printf("Current Step: %d", currentStep)
		//log.Printf("Current Percent: %d", currPercent)

		msgForWs := ws.NewSwupdateMessage(helpers.CurrentTime(), int(val.status), image, handler, info,
			int(totalBytesToDownload), int(totalSteps), int(downloadedData), int(currentStep), int(currPercent))

		//ws.GlobalHub.Broadcast(msgForWs)
		outgoingMsg := ws.NewOutgoingMessage(
			"swupdate_message",
			nil,
			connectionId,
			msgForWs,
		)
		ws.GlobalHub.Contact(outgoingMsg)
	}

}
