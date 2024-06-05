package logging

import (
	"log/slog"
	"os"
)

func NewJsonLogger() *slog.Logger {
	opts := slog.HandlerOptions{AddSource: true}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))

	return logger
}
