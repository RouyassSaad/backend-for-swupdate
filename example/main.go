// Package main
package main

import (
	"context"
	"log"
	"os"
	"time"
)

import "swupdate/bindings/golang"

func writeImage(cancel context.CancelFunc, stop <-chan interface{}, chunks chan<- []byte, path string) {
	defer close(chunks)

	const chunkSize = 4096

	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		cancel()
		log.Printf("could not open file %s: %v\n ", path, err)

		return
	}

	lastSize := chunkSize
	for lastSize == chunkSize {
		chunk := make([]byte, chunkSize)

		lastSize, err = file.Read(chunk)
		if err != nil {
			cancel()

			return
		}

		select {
		case chunks <- chunk: //
		case <-stop:
			return
		}
	}
}

func progressWithCallback(progression chan swupdate.ProgressMsg, cb func(swupdate.ProgressMsg)) {
	for progress := range progression {
		cb(progress)
	}
}

func waitUpdateEnd(end *swupdate.AsyncError) {
	<-end.Done()
	
	err := end.Err()
	if err != nil {
		log.Printf("Error during update: %+v\n\n", err)
	} else {
		log.Printf("Update ended successfully !\n\n")
	}
}
/*
func example1() {
	log.Printf("Example 1: Update with default values (synchronous)\n")

	update := swupdate.New()

	log.Printf("Swupdate with default values: %+v\n", update)

	err := update.Run("example/out.swu")
	if err != nil {
		log.Printf("Error during update: %+v\n\n", err)
	} else {
		log.Printf("Update ended successfully !\n\n")
	}
}
*/
func example2() {
	log.Printf("Example 2: Update with custom values and more usability (asynchronous)\n")

	// Construct with custom values
	update := swupdate.New(
		// you can give to Opt any member names from type SwupdateOpt
		swupdate.With("IpcAddr", "/run/swupdate/sockinstctrl"),
		swupdate.With("ProgressAddr", "/run/swupdate/swupdateprog"),
		swupdate.With("DisableStoreSwu", true),
		swupdate.With("SoftwareSet", "stable"),
		swupdate.With("RunningMode", "error"),
		swupdate.With("NetTimeout", 3 * time.Second),
		swupdate.With("ThisOptDoesNotExist", "Will never be inserted into the update obj"),
	)

	log.Printf("Swupdate with custom values: %+v\n", update)

	// create a context with one minute timeout for the update
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// chunks and progression can be buffered (16 here)
	bufferSize := 16
	chunks := make(chan []byte, bufferSize)
	progression := make(chan swupdate.ProgressMsg, bufferSize)
	

	ret := update.AsyncRunOpts(
		ctx,
		chunks,      // channel sending chunks to the update
		progression, // channel sending to the user the progress messages
	)

	// in the example, we start one routine per channel
	go progressWithCallback(progression, swupdate.PrintProgress)
	go writeImage(cancel, ret.Done(), chunks, "example/out.swu")

	// you could also put it inside a routine if needed
	waitUpdateEnd(ret)
}

func main() {
	//example1()

	example2()
}
