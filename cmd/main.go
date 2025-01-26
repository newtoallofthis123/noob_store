package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/newtoallofthis123/noob_store/api"
	"github.com/newtoallofthis123/noob_store/utils"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 6969, "Port to serve")
	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	env := utils.ReadEnv()
	env.ListenAddr = fmt.Sprintf(":%d", port)

	server := api.NewServer(&env, logger)
	server.Start()
}
