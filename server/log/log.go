package log

//This Section is just for creating an object to log messages in a readble,clear way.

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func newLogger() *slog.Logger {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: "15:04:05",
		AddSource:  false,
		NoColor:    false,
	})
	return slog.New(handler)
}

var Logger = newLogger()

/*
func main() {
	logger := newLogger()

	logger.Info("server starting",
		"service", "chat-api",
		"version", "1.0.0",
		"port", 8080,
	)

	err := run()
	if err != nil {
		logger.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	// app logic
	return nil
}
*/
