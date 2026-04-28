/*
 * (C) Copyright 2025
 *
 * SPDX-License-Identifier:     MIT
 */

package swupdate

/*
#include "cgo.h"

// Go cannot use values from #define, therefore we wrap it inside variables
__typeof__(((ipc_message*)0)->magic) IPC_MAGIC_NUMBER = IPC_MAGIC;
__typeof__(((union_msgdata_as_member_instmsg*)0)->req.apiversion) API_VERSION = SWUPDATE_API_VERSION;

*/
import "C"

import (
	"errors"
	"fmt"
	"log"
)

// IpcMessage is a wrapper to struct ipc_message from <network_ipc.h>.
type IpcMessage = C.wrapper_ipc_message

// SendImage will send the chunks to swupdate ipc.
func (update *Swupdate) SendImage(chunks <-chan []byte, stop <-chan RecoveryStatus) StatusResult {
	res := StatusResult{
		status: C.SUCCESS,
		err:    nil,
	}

	if update.ipc == nil {
		res.status = C.FAILURE
		res.err = errors.New("swupdate ipc was not created, you probably forgot to use the 'Start' method")

		return res
	}

	// send the chunks through ipc
	for buf := range chunks {
		if len(buf) < 1 {
			res.status = C.FAILURE
			res.err = fmt.Errorf("received bad image chunk (size: %d)", len(buf))
		}

		_, err := update.ipc.Write(buf)
		if err != nil {
			res.status = C.FAILURE
			res.err = fmt.Errorf("failed to send image chunk to swupdate: %w", err)

			break
		}

		// stop sending the buffer if we get a end status
		select {
		case res.status = <-stop:
			res.err = fmt.Errorf("forced to stop the update with status: %d", res.status)

			// do not break, the user should close the chunks channel to get out of here
		default:
			//
		}
	}

	errClose := update.ipc.Close()
	if errClose != nil {
		log.Printf("WARNING: could not close properly the update ipc: %v\n", errClose)
	}

	return res
}

// SendMsg sends a message to 'Swupdate' ipc.
func (update *Swupdate) SendMsg(msg IpcMessage) error {
/*	buf := make([]byte, unsafe.Sizeof(msg))

	_, err := binary.Encode(buf, binary.NativeEndian, &msg)
	if err != nil {
		return fmt.Errorf("could not encode the message as binary: %w\n", err)
	}
*/
	_, err := update.ipc.Write(msg[:])
	if err != nil {
		return fmt.Errorf("could not send the message to ipc: %w", err)
	}

	return nil
}

// PrepareInstMsg prepares a message of type 'instmsg' from update options.
func (update *Swupdate) PrepareInstMsg() IpcMessage {
	// union inner anonymous structs have to be casted from union type
	data := C.msgdata{}
	instmsg := C.cast_union_msgdata_as_member_instmsg(&data)
	//         ^ we have to use this generated function from "cgo.h"
	//           because cgo is not able to use union members

	// instmsg.req is type C.struct_swupdate_request
	instmsg.req.apiversion = C.API_VERSION
	instmsg.req.dry_run = C.RUN_DEFAULT
	instmsg.req.disable_store_swu = C.bool(update.opts.DisableStoreSwu)
	CopyToCChars(instmsg.req.software_set[:], update.opts.SoftwareSet)
	CopyToCChars(instmsg.req.running_mode[:], update.opts.RunningMode)

	msg := IpcMessage{}
	ipcmsg := C.cast_union_wrapper_ipc_message_as_member_msg(&msg)
	//         ^ we have to use this generated function from "cgo.h"
	//           because cgo is not able to use union members

	ipcmsg.magic = C.IPC_MAGIC_NUMBER
	// 'type' becomes '_type' with cgo
	ipcmsg._type = C.REQ_INSTALL
	ipcmsg.data =  data
	
	return msg
}
