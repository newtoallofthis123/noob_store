package api

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/newtoallofthis123/noob_store/types"
)

func (s *Server) checkAuth(authKey string) (types.Session, bool) {
	if authKey == "" {
		return types.Session{}, false
	}

	session, err := s.cache.GetSession(authKey)
	if err != nil {
		session, err = s.db.GetSession(authKey)
		if err != nil {
			return types.Session{}, false
		}
	}

	return session, true
}

func (s *Server) handleFileMetadataById(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		c.JSON(500, gin.H{"err": "Invalid Authorization or missing session"})
		return
	}

	id, exists := c.Params.Get("id")
	if !exists {
		c.JSON(500, gin.H{"err": "id needed"})
		return
	}

	metadata, err := s.cache.GetMetadata(id)
	if err != nil {
		metadata, err = s.db.GetMetaDataById(id)
		if err != nil {
			c.JSON(500, gin.H{"err": "Failed to retrieve metadata: " + err.Error()})
			return
		}
	}

	if metadata.UserId != session.UserId {
		c.JSON(500, gin.H{"err": "Unauthorized access to file from userId: " + session.UserId})
		return
	}

	c.JSON(200, metadata)
}

func (s *Server) handleFileDownloadById(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		c.JSON(500, gin.H{"err": "Invalid Authorization or missing session"})
		return
	}

	id, exists := c.Params.Get("id")
	if !exists {
		c.JSON(500, gin.H{"err": "blob id needed"})
		return
	}

	meta, err := s.cache.GetMetadata(id)
	if err != nil {
		meta, err = s.db.GetMetaDataById(id)
		if err != nil {
			c.JSON(500, gin.H{"err": "Failed to retrieve metadata: " + err.Error()})
			return
		}
	}

	if meta.UserId != session.UserId {
		c.JSON(500, gin.H{"err": "Unauthorized access to file from userId: " + session.UserId})
		return
	}

	blob, err := s.cache.GetBlob(meta.Blob)
	if err != nil {
		blob, err = s.db.GetBlobById(meta.Blob)
		if err != nil {
			c.JSON(500, gin.H{"err": "Failed to retrieve blob: " + err.Error()})
			return
		}
	}

	err = s.handler.Get(&blob)
	if err != nil {
		c.JSON(500, gin.H{"err": "Failed to retrieve blob: " + err.Error()})
		return
	}

	fmt.Println(len(blob.Content))

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
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		c.JSON(500, gin.H{"err": "Invalid Authorization or missing session"})
		return
	}

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

	blob, meta, err := s.handler.Insert(path, content, session.UserId)
	err = s.db.InsertBlob(blob)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}
	err = s.db.InsertMetaData(meta)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}
	err = s.cache.InsertMetadata(meta)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}
	err = s.cache.InsertBlob(blob)
	if err != nil {
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}

	c.JSON(200, meta)
}
