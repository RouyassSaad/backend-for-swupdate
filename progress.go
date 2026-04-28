/*
 * (C) Copyright 2025
 *
 * SPDX-License-Identifier:     MIT
 */

package swupdate

/*
#cgo LDFLAGS: -lswupdate

#include "cgo.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"log"
)

// ProgressMsg is an wrapper to struct progress_msg from <progress_ipc.h>.
type ProgressMsg = C.wrapper_progress_msg

// PrintProgress is an helper function for printing progress, useful for callbacks.
func PrintProgress(progress ProgressMsg) {
	val := C.cast_union_wrapper_progress_msg_as_member_msg(&progress)
	image := GoStringFromSlice(val.cur_image[:])
	handler := GoStringFromSlice(val.hnd_name[:])
	info := GoStringFromSlice(val.info[:val.infolen])
	log.Printf("Update Status: %d, image: %v handler: %v, info: %v\n", val.status, image, handler, info)
}

// ProgressForEach calls th cb callback till the progression channel is open.
func ProgressForEach(progression <-chan ProgressMsg, cb func(ProgressMsg)) {
	for progress := range progression {
		cb(progress)
	}
}

// ProgressRead will receive messages from the progress ipc.
func (update *Swupdate) ProgressRead(progression chan<- ProgressMsg, stop <-chan RecoveryStatus) StatusResult {
	if progression != nil {
		defer close(progression)
	}

	res := StatusResult{
		status: C.FAILURE,
		err:    nil,
	}

loop:
	for res.err == nil {
		var msg *ProgressMsg

		msg, res.err = update.progressRead()
		if res.err != nil {
			break loop
		}


		status := update.sendProgression(progression, stop, *msg)
		if status != nil {
			res.status = *status
			res.err = fmt.Errorf("forced to stop the progress with status: %d", res.status)

			break loop
		}

		progress := C.cast_union_wrapper_progress_msg_as_member_msg(msg)

		// end of the update
		if progress.status == C.FAILURE || progress.status == C.SUCCESS {
			res.status = progress.status

			break loop
		}

		// the caller can explicitly ask for stopping the update process
		select {
		case res.status = <-stop:
			res.err = fmt.Errorf("forced to stop the progress with status: %d", res.status)

		default:
			//
		}
	}

	err := update.progress.Close()
	if err != nil {
		log.Printf("WARNING: could not close properly the progress ipc: %v\n", err)
	}
	update.progress = nil

	return res
}

func (update *Swupdate) progressRead() (*ProgressMsg, error) {
	if update.progress == nil {
		return nil, errors.New("progress ipc was not created, did you forget to Dial it ?")
	}

	var msg ProgressMsg

	_, err := update.progress.Read(msg[:])
	if err != nil {
		return nil, fmt.Errorf("could not read from progress ipc: %w", err)
	}

	return &msg, nil
}

func (update *Swupdate) sendProgression(
	progression chan<- ProgressMsg,
	stop <-chan RecoveryStatus,
	msg ProgressMsg,
) *RecoveryStatus {
	// send the message or stop
	select {
	case progression <- msg:
		return nil
	case status := <-stop:
		return &status
	}
}
