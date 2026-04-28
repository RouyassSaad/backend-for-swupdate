/*
 * (C) Copyright 2025
 *
 * SPDX-License-Identifier:     MIT
 */

// Package swupdate implements an IPC client communicating with a swupdate daemon
package swupdate

/*
#cgo LDFLAGS: -lswupdate

#include "cgo.h"
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

// RecoveryStatus is an alias to the type of the status in progress_msg, which is compatible with RECOVERY_STATUS.
type RecoveryStatus = C.C_RECOVERY_STATUS

// Opts is helpful to construct a 'Swupdate' object with New(With("<Opt-Member>", "value")).
type Opts struct {
	// Addresses to unix ipc.
	IpcAddr      string
	ProgressAddr string

	// NetTimeout will stop net.Dial calls if too long. Set to -1 to never timeout.
	NetTimeout time.Duration

	// Update options
	SoftwareSet     string
	RunningMode     string
	DisableStoreSwu bool
}

// OptsFunc the function type of 'Opt' to be used in 'New'.
type OptsFunc func(*Opts) error

// Swupdate is an type containing Optss, its method are helpful to run an update through IPC.
type Swupdate struct {
	ipc      net.Conn
	progress net.Conn

	opts Opts
}

// StatusResult is a '(status, err)' tuple.
type StatusResult struct {
	status RecoveryStatus
	err    error
}

// AsyncError is returned by AsynRun(Opts).
type AsyncError struct {
	end chan interface{}
	err error
}

// newDefaults create the default 'Swupdate' object, behaviour of 'New' with no arguments.
func newDefault() *Swupdate {
	return &Swupdate{
		opts: Opts{
			IpcAddr:         C.GoString(C.get_ctrl_socket()),
			ProgressAddr:    C.GoString(C.get_prog_socket()),
			NetTimeout:      -1,
			SoftwareSet:     "",
			RunningMode:     "",
			DisableStoreSwu: false,
		},
		ipc:      nil,
		progress: nil,
	}
}

// New creates a 'SWupdate' object with given options with 'Opts'
//
// usage: `New(With("SomeField", "value"), With("OtherField", true))`
//
// Note that valid fields are members of the `Opts` struct.
func New(optFuncs ...OptsFunc) *Swupdate {
	val := newDefault()

	// apply options
	for _, optFunc := range optFuncs {
		err := optFunc(&val.opts)
		if err != nil {
			// Print a warning in case of bad 'Opts'
			log.Println("WARNING: ", err)
		}
	}

	return val
}

// With is useful to construct a 'Swupdate' with options from 'Opts'.
//
// Calling `With("MyField", value)(&opts)` is equivalent to `opts.MyField = value`
// but returns an error at runtime if it is not allowed instead of not compiling.
//
// This is used in the `New` function to accept the form `New(With("MyField", value))`.
func With(field string, value interface{}) OptsFunc {
	refValue := reflect.ValueOf(value)

	// this function will apply the `value` to the `field` in `opts`
	return func(opts *Opts) error {
		if !refValue.IsValid() {
			return fmt.Errorf("the value '%+v' given to field %s is not valid", value, field)
		}

		refOpts := reflect.ValueOf(opts).Elem().FieldByName(field)

		if !refOpts.IsValid() {
			return fmt.Errorf("the field %s is a not valid Opts member", field)
		}
		if !refOpts.CanSet() {
			return fmt.Errorf("the field %s is cannot be set in Opts", field)
		}

		optType := refOpts.Type().Name()
		valType := refValue.Type().Name()
		if optType != valType {
			return fmt.Errorf(
				"the field %s is of type %s, but a value of type %s was given in Opts",
				field, optType, valType,
			)
		}
		refOpts.Set(refValue)

		return nil
	}
}

// Start inits the update by sending an 'instmsg'.
func (update *Swupdate) Start() error {
	return update.StartCtx(context.Background())
}

// StartCtx inits the update by sending an 'instmsg' with a context.
func (update *Swupdate) StartCtx(ctx context.Context) error {
	var err error

	msg := update.PrepareInstMsg()

	var dialer net.Dialer
	var dialCtx context.Context
	var cancel context.CancelFunc

	if update.opts.NetTimeout >= 0 {
		dialCtx, cancel = context.WithTimeout(ctx, update.opts.NetTimeout)
	} else {
		dialCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	update.progress, err = dialer.DialContext(dialCtx, "unix", update.opts.ProgressAddr)
	if err != nil {
		return fmt.Errorf("could not dial progress ipc: %w", err)
	}

	update.ipc, err = dialer.DialContext(dialCtx, "unix", update.opts.IpcAddr)
	if err != nil {
		return fmt.Errorf("could not dial update ipc: %w", err)
	}

	err = update.SendMsg(msg)
	if err != nil {
		return fmt.Errorf("could not send update start message to swupdate: %w", err)
	}

	return nil
}

// Run will send chunks of the file to swupdate daemon, blocking until update end.
func (update *Swupdate) Run(path string) error {
	return update.RunOpts(context.Background(), path, nil)
}

// RunOpts will read progression et send chunks to swupdate daemon, blocking until update end.
func (update *Swupdate) RunOpts(
	ctx context.Context,
	path string,
	progressCallback func(ProgressMsg),
) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", path, err)
	}

	const bufferChannelSIze = 16

	var progression chan ProgressMsg
	chunks := make(chan []byte, bufferChannelSIze)

	if progressCallback != nil {
		progression = make(chan ProgressMsg, bufferChannelSIze)
	}

	asyncCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	ret := update.AsyncRunOpts(asyncCtx, chunks, progression)

	var group sync.WaitGroup

	// send image chunk by chunk
	group.Add(1)
	go func() {
		defer group.Done()
		SendFileAsChunks(asyncCtx, cancel, ret.Done(), file, chunks)
	}()

	// run the callback on received progress messages
	group.Add(1)
	go func() {
		defer group.Done()
		ProgressForEach(progression, progressCallback)
	}()

	// wait for update end before return
	<-ret.Done()
	group.Wait()

	return ret.Err()
}

// SendFileAsChunks will send the file to the chunks channel.
func SendFileAsChunks(
	parentCtx context.Context,
	cancelFunc context.CancelFunc,
	updateEnd <-chan interface{},
	file *os.File,
	chunks chan<- []byte,
) {
	var err error
	const chunkSize = 4096
	chunk := make([]byte, chunkSize)

	// read the file chunk by chunk
	lastSize := chunkSize
	for lastSize == chunkSize {
		lastSize, err = file.Read(chunk)
		// only stop if the error is not EOF
		if err != nil && !errors.Is(err, io.EOF) {
			cancelFunc()

			return
		}

		if lastSize == 0 {
			return
		}

		// send the chunk only if the update is still running
		select {
		case chunks <- chunk: //
		case <-updateEnd:
			return
		case <-parentCtx.Done():
			return
		}
	}
}

// AsyncRun runs the update with only an image as input.
func (update *Swupdate) AsyncRun(chunks <-chan []byte) *AsyncError {
	return update.AsyncRunOpts(context.Background(), chunks, nil)
}

// AsyncRunOpts will read progression et send chunks to swupdate daemon in a non-blocking manner.
// The returned channel is used as a signal that the update is ended.
func (update *Swupdate) AsyncRunOpts(
	ctx context.Context,
	chunks <-chan []byte,
	progression chan<- ProgressMsg,
) *AsyncError {
	if chunks == nil {
		return &AsyncError{
			end: nil,
			err: errors.New("the chunks channel cannot be nil"),
		}
	}

	err := update.StartCtx(ctx)
	if err != nil {
		return &AsyncError{
			end: nil,
			err: err,
		}
	}

	ret := AsyncError{
		end: make(chan interface{}, 1),
		err: nil,
	}

	// synchronization channels
	stopProgress, stopRun := make(chan RecoveryStatus, 1), make(chan RecoveryStatus, 1)
	progressEnd, runEnd := make(chan StatusResult), make(chan StatusResult)

	// synchronize routines terminaisons
	go func() {
		// close the channel to release 
		defer close(ret.end)
		// set the error before sending to user
		ret.err = synchronizeRoutines(ctx, stopRun, stopProgress, runEnd, progressEnd)
		ret.end <- nil
	}()

	// run progression read
	go func() { progressEnd <- update.ProgressRead(progression, stopProgress) }()

	// send the update image to swupdate
	go func() { runEnd <- update.SendImage(chunks, stopRun) }()

	// the return value is a channel, it will receive an error or nil when the update is ended
	return &ret
}

func synchronizeRoutines(
	ctx context.Context,
	stopRun chan<- RecoveryStatus,
	stopProgress chan<- RecoveryStatus,
	runEnd <-chan StatusResult,
	progressEnd <-chan StatusResult,
) error {	
	var res StatusResult

	runEnded := false
	progressEnded := false

	for !progressEnded || !runEnded {
		select {
		case <-ctx.Done():
			res.err = ctx.Err()
			// stop both progress and run
			runEnded = stopRoutine(res.status, runEnded, stopRun, runEnd)
			progressEnded = stopRoutine(res.status, progressEnded, stopProgress, progressEnd)

		case res = <-progressEnd:
			progressEnded = true
			if res.err == nil && res.status != C.SUCCESS {
				res.err = fmt.Errorf("update ended with status: %v", res)
			}
			// progress should not live more than run, stop run
			runEnded = stopRoutine(res.status, runEnded, stopRun, runEnd)

		case res = <-runEnd:
			runEnded = true
			// if an error occurred in run, stop progress
			if res.err != nil {
				progressEnded = stopRoutine(
					res.status, progressEnded, stopProgress, progressEnd)
			}
		}
	}

	return res.err
}

func stopRoutine(
	status RecoveryStatus,
	ended bool,
	stopChan chan<- RecoveryStatus,
	endChan <-chan StatusResult,
) bool {
	if !ended {
		select {
		case stopChan <- status:
			<-endChan
		case <-endChan:
			//
		}
	}

	return true
}

// Done returns the channel telling the update ended.
func (asyncError *AsyncError) Done() <-chan interface{} {
	return asyncError.end
}

// Err returns the error of the update if it exists, nil overwise.
func (asyncError *AsyncError) Err() error {
	return asyncError.err
}
