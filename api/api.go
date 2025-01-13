package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/newtoallofthis123/noob_store/db"
	"github.com/newtoallofthis123/noob_store/fs"
	"github.com/newtoallofthis123/noob_store/utils"
)

type Server struct {
	listenAddr string
	logger     *slog.Logger
	db         *db.Store
	handler    *fs.Handler
}

func NewServer(logger *slog.Logger) *Server {
	env := utils.ReadEnv()

	store, err := db.NewStore(env.ConnString)
	if err != nil {
		panic(err)
	}

	err = store.InitTables()
	if err != nil {
		panic(err)
	}

	buckets, err := fs.DiscoverBuckets(env.BucketPath)

	handler := fs.NewHandler(buckets, store, logger)

	return &Server{
		listenAddr: env.ListenAddr,
		logger:     logger,
		db:         store,
		handler:    &handler,
	}
}

func (s *Server) handleListFiles(c *gin.Context) {

}

func (s *Server) Start() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "pong")
	})

	err := r.Run(s.listenAddr)
	if err != nil {
		s.logger.Error("Closing Server with err: " + err.Error())
	}
}
