package main

import (
	"log/slog"
	"os"

	"github.com/joakimcarlsson/go-router/example/internal/server"
)

func main() {
	slog := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	srv := server.NewServer(slog)
	httpServer := srv.HTTP()

	slog.Info("Server listening on port", "port", "6784")
	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error("Server failed to start", "err", err)
	}
}
