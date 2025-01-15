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

	err = store.InitTables()
	if err != nil {
		panic(err)
	}

	cache, err := cache.NewCache(env.CacheConn)
	if err != nil {
		panic(err)
	}

	buckets, err := fs.DiscoverBuckets(env.BucketPath)
	if len(buckets) == 0 || err != nil {
		buckets = fs.GenerateBuckets(env.BucketPath, 8)
	}

	handler := fs.NewHandler(buckets, &store, logger)

	return &Server{
		listenAddr: env.ListenAddr,
		logger:     logger,
		db:         &store,
		cache:      &cache,
		handler:    &handler,
	}
}

func (s *Server) Start() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.POST("/add", s.handleFileAdd)
	r.GET("/info/:id", s.handleFileMetadataById)
	r.GET("/file/:id", s.handleFileDownloadById)
	r.POST("/signup", s.handleCreateUser)
	r.POST("/login", s.handleLoginUser)

	err := r.Run(s.listenAddr)
	if err != nil {
		s.logger.Error("Closing Server with err: " + err.Error())
	}
}
