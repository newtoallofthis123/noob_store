package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/newtoallofthis123/noob_store/cache"
	"github.com/newtoallofthis123/noob_store/db"
	"github.com/newtoallofthis123/noob_store/fs"
	"github.com/newtoallofthis123/noob_store/utils"
)

type Server struct {
	listenAddr string
	logger     *slog.Logger
	db         *db.Store
	cache      *cache.Cache
	handler    *fs.Handler
}

func NewServer(logger *slog.Logger) *Server {
	env := utils.ReadEnv()

	store, err := db.NewStore(env.ConnString)
	if err != nil {
		panic(err)
	}

	logger.Info("Connected db storage")

	err = store.InitTables()
	if err != nil {
		panic(err)
	}

	logger.Info("Initialized Tables")

	cache, err := cache.NewCache(env.CacheConn)
	if err != nil {
		panic(err)
	}

	logger.Info("Connecyed to Cache")

	buckets, err := fs.DiscoverBuckets(env.BucketPath)
	if len(buckets) == 0 || err != nil {
		buckets = fs.GenerateBuckets(env.BucketPath, 8)
	}

	logger.Info("Discovered and found buckets")

	handler := fs.NewHandler(buckets, &store, logger)

	logger.Info("Initialized new fs handler")

	return &Server{
		listenAddr: env.ListenAddr,
		logger:     logger,
		db:         &store,
		cache:      &cache,
		handler:    &handler,
	}
}

func (s *Server) Start() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"noob_store": gin.H{"version": "0.1", "author": "NoobScience", "status": "up"}})
	})
	r.POST("/add", s.handleFileAdd)
	r.GET("/info/:id", s.handleFileMetadataById)
	r.GET("/file/:id", s.handleFileDownloadById)
	r.DELETE("/delete/:id", s.handleDeleteFile)

	user := r.Group("/user")

	user.POST("/create", s.handleCreateUser)
	user.POST("/login", s.handleLoginUser)
	user.GET("/ls", s.handleUserLs)
	user.GET("/path_ls", s.handleUserPathLs)
	user.DELETE("/delete_dir/:dir", s.handleDeleteDir)

	s.logger.Info("Initialized routes")
	s.handler.LogBucketsInfo()

	err := r.Run(s.listenAddr)
	if err != nil {
		s.logger.Error("Closing Server with err: " + err.Error())
	}
}
