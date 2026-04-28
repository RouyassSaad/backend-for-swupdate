package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	config "swupdate/bindings/golang/server/config"
	handler "swupdate/bindings/golang/server/handlers"
	lg "swupdate/bindings/golang/server/log"
	wsmanager "swupdate/bindings/golang/server/websocket"

	swupdate "swupdate/bindings/golang"
	sd "swupdate/bindings/golang/sd-journal"
	t "swupdate/bindings/golang/server/models"
)

// takes a file name from a chanel and start update
func LaunchUpdate(update *swupdate.Swupdate, updateChan chan *t.UpdateChanel, journalctlChan chan *string) {

	for {
		syncChan := <-updateChan

		dir := "/tmp/" + syncChan.Filename
		connectionId := syncChan.ConnexionId
		journalctlChan <- &connectionId

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

		bufferSize := 16
		chunks := make(chan []byte, bufferSize)
		progression := make(chan swupdate.ProgressMsg, bufferSize)

		ret := update.AsyncRunOpts(
			ctx,
			chunks,
			progression,
		)

		go progressWithCallback(progression, swupdate.ProgressCallbackClosure(connectionId)) //ProgressCallbackClosure(connectionId)
		go writeImage(cancel, ret.Done(), chunks, dir)

		waitUpdateEnd(ret)
		cancel()
	}

}

func main() {

	lg.Logger.Info("Starting Server", "project", "backend-for-swupdate")

	configFilePath := flag.String("configPath", "server/config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configFilePath)

	updateChan := make(chan *t.UpdateChanel) // to sync the server with swupdate service
	journalctlChan := make(chan *string)

	if err != nil {
		lg.Logger.Error("Failed To Load config.yaml", "Error", err)
		return
	}

	go wsmanager.GlobalHub.Run() //go routine to Listen/Brodcast to connections

	update := swupdate.New(
		swupdate.With("IpcAddr", "/run/swupdate/sockinstctrl"),
		swupdate.With("ProgressAddr", "/run/swupdate/swupdateprog"),
		swupdate.With("DisableStoreSwu", true),
		swupdate.With("SoftwareSet", "stable"),
		swupdate.With("RunningMode", "working"),
		swupdate.With("NetTimeout", 3*time.Second),
	)
	//log.Logger.Info("SWUpdate object created", "obj", update)

	//wait for a request to start an update
	go LaunchUpdate(update, updateChan, journalctlChan)

	http.HandleFunc(cfg.Routes.Root, handler.RootHandler)
	http.HandleFunc(cfg.Routes.Ws, handler.WsHandler)
	http.HandleFunc(cfg.Routes.UploadFile, handler.FileUploadHandler(updateChan))
	http.HandleFunc(cfg.Routes.GetUploadedFiles, handler.GetUploadedFiles)
	http.HandleFunc(cfg.Routes.Health, handler.HealthHandler)

	go sd.GetSystemMdLogs(journalctlChan)

	if err := http.ListenAndServe(cfg.Server.Port, nil); err != nil {
		lg.Logger.Error("Failed to start server", "Error", err)
	}

}

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
