package api

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

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

func (s *Server) handleFileMetadata(c *gin.Context) {
	path := c.Query("path")
	name := c.Query("name")
	if path == "" {
		c.JSON(500, gin.H{"err": "Need a valid path"})
		return
	}
	if name == "" {
		name = filepath.Base(path)
	}

	path = filepath.Clean(path)
	name = filepath.Clean(name)

	fmt.Println(name, path)
	metadata, err := s.db.GetMetaData(name, path)
	if err != nil {
		c.JSON(500, gin.H{"err": "Failed to retrieve metadata: " + err.Error()})
		return
	}

	c.JSON(200, metadata)
}

func (s *Server) handleFileDownload(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(500, gin.H{"err": "Need a valid path"})
		return
	}

	blob, err := s.handler.Get(path)
	if err != nil {
		c.JSON(500, gin.H{"err": "Failed to retrieve blob: " + err.Error()})
		return
	}

	n, err := c.Writer.Write(blob.Content)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to write: " + err.Error()})
		return
	}
	if n != len(blob.Content) {
		c.JSON(500, gin.H{"err": "File integrity check failed"})
		return
	}
}

func (s *Server) handleFileMetadataById(c *gin.Context) {
	id, exists := c.Params.Get("id")
	if !exists {
		c.JSON(500, gin.H{"err": "id needed"})
		return
	}

	metadata, err := s.db.GetMetaDataById(id)
	if err != nil {
		c.JSON(500, gin.H{"err": "Failed to retrieve metadata: " + err.Error()})
		return
	}

	c.JSON(200, metadata)
}

func (s *Server) handleFileDownloadById(c *gin.Context) {
	id, exists := c.Params.Get("id")
	if !exists {
		c.JSON(500, gin.H{"err": "blob id needed"})
		return
	}

	meta, err := s.db.GetMetaDataById(id)
	if err != nil {
		c.JSON(500, gin.H{"err": "Failed to retrieve metadata: " + err.Error()})
		return
	}

	blob, err := s.handler.Get(meta.Path)
	if err != nil {
		c.JSON(500, gin.H{"err": "Failed to retrieve blob: " + err.Error()})
		return
	}

	n, err := c.Writer.Write(blob.Content)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to write: " + err.Error()})
		return
	}
	if n != len(blob.Content) {
		c.JSON(500, gin.H{"err": "File integrity check failed"})
		return
	}
}
func (s *Server) handleFileAdd(c *gin.Context) {
	path, exists := c.GetPostForm("path")
	if !exists {
		c.JSON(500, gin.H{"err": "Path is needed in the post form"})
		return
	}

	fileHeader, err := c.FormFile("content")
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to read file: " + err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to read file: " + err.Error()})
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to read file: " + err.Error()})
		return
	}

	_, meta, err := s.handler.Insert(path, content)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}

	c.JSON(200, meta)
}

func (s *Server) Start() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "pong")
	})
	r.GET("/ls", s.handleListFiles)
	r.GET("/get", s.handleFileMetadata)
	r.GET("/download", s.handleFileDownload)
	r.POST("/add", s.handleFileAdd)
	r.GET("/metadata/:id", s.handleFileMetadataById)
	r.GET("/blob/:id", s.handleFileDownloadById)

	err := r.Run(s.listenAddr)
	if err != nil {
		s.logger.Error("Closing Server with err: " + err.Error())
	}
}
