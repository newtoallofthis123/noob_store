package main

import (
	"log/slog"
	"os"

	"github.com/newtoallofthis123/noob_store/api"
)

func main() {
	// New Json Logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	server := api.NewServer(logger)
	server.Start()
}
