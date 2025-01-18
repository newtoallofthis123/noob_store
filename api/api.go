package api

import (
	"log/slog"
	"math/rand"
	"sync"

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
	counter    int64
	mu         sync.RWMutex
	pruning    bool
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

	handler := fs.NewHandler(buckets, logger, &env)

	logger.Info("Initialized new fs handler")

	return &Server{
		listenAddr: env.ListenAddr,
		logger:     logger,
		db:         &store,
		cache:      &cache,
		handler:    &handler,
		counter:    0,
	}
}

// Pruner returns a gin.HandlerFunc that prunes old or unnecessary data from the server.
func (s *Server) Pruner() gin.HandlerFunc {
	return func(c *gin.Context) {
		// To make sure that this runs very very very rarely, we make up a math thing that it has to get through
		n := rand.Int63()
		s.mu.Lock()
		if s.counter%7 == 0 && n%14 == 0 && !s.pruning {
			err := s.DeleteFreeSpace()
			if err != nil {
				s.logger.Error("Failed to prune free space: With err: " + err.Error())
			}
			s.pruning = true
		}
		s.counter++
		s.mu.Unlock()
	}
}

func (s *Server) Start() {
	r := gin.Default()

	// Setup pruner as a middleware
	r.Use(s.Pruner())

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
