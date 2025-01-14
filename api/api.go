package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/newtoallofthis123/noob_store/db"
	"github.com/newtoallofthis123/noob_store/fs"
	"github.com/newtoallofthis123/noob_store/types"
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
	dir := c.Query("dir")
	var meta []types.Metadata
	var err error
	if dir == "" {
		meta, err = s.db.GetAllFiles()
		if err != nil {
			c.JSON(500, gin.H{"err": "Unable to get all files: " + err.Error()})
		}
	} else {
		meta, err = s.db.GetMetaDataByDir(dir)
		if err != nil {
			c.JSON(500, gin.H{"err": "Unable to get all files: " + err.Error()})
		}
	}

	c.JSON(200, meta)
}

func (s *Server) Start() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.GET("/ls", s.handleListFiles)
	r.GET("/get", s.handleFileMetadata)
	r.GET("/download", s.handleFileDownload)
	r.POST("/add", s.handleFileAdd)
	r.GET("/metadata/:id", s.handleFileMetadataById)
	r.GET("/blob/:id", s.handleFileDownloadById)
	r.POST("/signup", s.handleCreateUser)
	r.POST("/login", s.handleLoginUser)

	err := r.Run(s.listenAddr)
	if err != nil {
		s.logger.Error("Closing Server with err: " + err.Error())
	}
}
