package api

import (
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
		s.logger.Debug("Cache miss for session with id: " + authKey)
		if err != nil {
			return types.Session{}, false
		}
		_ = s.cache.InsertSession(session)
		s.logger.Debug("Cache refreshed for id: " + authKey)
	}

	s.logger.Debug("Authenticated request from: " + session.UserId)
	return session, true
}

func (s *Server) handleFileMetadataById(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		s.logger.Error("Unauthorized session: " + authKey)
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
		s.logger.Debug("Cache miss for metadata: " + id)
		metadata, err = s.db.GetMetaDataById(id)
		if err != nil {
			s.logger.Error("No metadata with id: " + id + " with err: " + err.Error())
			c.JSON(500, gin.H{"err": "Failed to retrieve metadata: " + err.Error()})
			return
		}

		_ = s.cache.InsertMetadata(metadata)
		s.logger.Debug("Cache refreshed for metadata with id: " + metadata.Id)
	}

	if metadata.UserId != session.UserId {
		s.logger.Warn("Prevented Unauthorized access for file from userId" + session.UserId)
		c.JSON(500, gin.H{"err": "Unauthorized access to file from userId: " + session.UserId})
		return
	}

	s.logger.Debug("Successfully outputed metadata for id: " + metadata.Id)
	c.JSON(200, metadata)
}

func (s *Server) handleFileDownloadById(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		s.logger.Error("Unauthorized session: " + authKey)
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
		s.logger.Debug("Cache miss for metadata: " + id + " with err: " + err.Error())
		meta, err = s.db.GetMetaDataById(id)
		if err != nil {
			s.logger.Error("No metadata with id: " + id + " with err: " + err.Error())
			c.JSON(500, gin.H{"err": "Failed to retrieve metadata: " + err.Error()})
			return
		}

		_ = s.cache.InsertMetadata(meta)
		s.logger.Debug("Cache refreshed for metadata with id: " + meta.Id)
	}

	if meta.UserId != session.UserId {
		s.logger.Warn("Prevented Unauthorized access for file from userId" + session.UserId)
		c.JSON(500, gin.H{"err": "Unauthorized access to file from userId: " + session.UserId})
		return
	}

	blob, err := s.cache.GetBlob(meta.Blob)
	if err != nil {
		s.logger.Debug("Cache miss for blob with id: " + meta.Blob + " with err: " + err.Error())
		blob, err = s.db.GetBlobById(meta.Blob)
		if err != nil {
			s.logger.Error("No Cache with Id: " + meta.Blob + " with err: " + err.Error())
			c.JSON(500, gin.H{"err": "Failed to retrieve blob: " + err.Error()})
			return
		}
		s.logger.Debug("Cache refreshed for blob with id: " + meta.Blob)
	}

	err = s.handler.Get(&blob)
	if err != nil {
		s.logger.Error("Failed to fill data for blob: " + blob.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Failed to retrieve blob: " + err.Error()})
		return
	}

	n, err := c.Writer.Write(blob.Content)
	if err != nil {
		s.logger.Error("Failed to write data to response for blob: " + blob.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to write: " + err.Error()})
		return
	}
	if n != len(blob.Content) {
		s.logger.Error("Failed integrity data for blob: " + blob.Id)
		c.JSON(500, gin.H{"err": "File integrity check failed"})
		return
	}
	//TODO: Add hash check
}
func (s *Server) handleFileAdd(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		s.logger.Error("Unauthorized session: " + authKey)
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
		s.logger.Error("Unable to read content file for: " + authKey + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to read file: " + err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		s.logger.Error("Unable to read content file for " + authKey + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to read file: " + err.Error()})
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		s.logger.Error("Unable to read content file for " + authKey + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to read file: " + err.Error()})
		return
	}

	blob, meta, err := s.handler.Insert(path, content, session.UserId)
	err = s.db.InsertBlob(blob)
	if err != nil {
		s.logger.Debug("Unable to insert blob " + blob.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}
	err = s.db.InsertMetaData(meta)
	if err != nil {
		s.logger.Debug("Unable to insert metadata " + meta.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}
	err = s.cache.InsertBlob(blob)
	if err != nil {
		s.logger.Error("Unable to insert blob to cache " + blob.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}
	err = s.cache.InsertMetadata(meta)
	if err != nil {
		s.logger.Error("Unable to insert metadata to cache " + meta.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to insert file: " + err.Error()})
		return
	}

	s.logger.Debug("Successfully added blob: " + blob.Id)
	c.JSON(200, meta)
}
