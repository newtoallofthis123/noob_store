package api

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/newtoallofthis123/noob_store/types"
	"github.com/newtoallofthis123/noob_store/utils"
)

func (s *Server) checkAuth(authKey string) (types.Session, bool) {
	if authKey == "" {
		return types.Session{}, false
	}

	authKey = strings.Replace(authKey, "Bearer ", "", 1)
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
	if authKey == "" {
		authKey = c.Query("session")
	}
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

	s.mu.RLock()
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
	s.mu.RUnlock()

	if meta.UserId != session.UserId {
		s.logger.Warn("Prevented Unauthorized access for file from userId" + session.UserId)
		c.JSON(500, gin.H{"err": "Unauthorized access to file from userId: " + session.UserId})
		return
	}

	s.mu.RLock()
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
	s.mu.RUnlock()

	err = s.handler.Get(&blob)
	if err != nil {
		s.logger.Error("Failed to fill data for blob: " + blob.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Failed to retrieve blob: " + err.Error()})
		return
	}

	checksum := utils.CalHash(blob.Content)
	if checksum != blob.Checksum {
		c.JSON(500, gin.H{"err": "File check validity failed! File recovery not possible: recommeneded deletion"})
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

	path = filepath.Clean(path)
	s.mu.RLock()
	existing, err := s.db.GetMetaDataByPath(path)
	if err == nil && existing.UserId == session.UserId {
		s.logger.Error("Attempt at adding duplicate path: " + path)
		c.JSON(500, gin.H{"err": "Path already exists for user in store"})
		return
	}
	s.mu.RUnlock()

	s.mu.Lock()
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
	s.mu.Unlock()

	s.logger.Debug("Successfully added blob: " + blob.Id)
	s.logger.Info("Added to bucket: " + blob.Bucket)
	s.handler.LogBucketsInfo()
	c.JSON(200, meta)
}

func (s *Server) handleDeleteFile(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		s.logger.Error("Unauthorized session: " + authKey)
		c.JSON(500, gin.H{"err": "Invalid Authorization or missing session"})
		return
	}

	fileId := c.Param("id")

	s.mu.RLock()
	meta, err := s.db.GetMetaDataById(fileId)
	if err != nil {
		s.logger.Error("Unable to find file with id: " + fileId + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to find file"})
		return
	}
	s.mu.RUnlock()

	if meta.UserId != session.UserId {
		s.logger.Warn("Prevented Unauthorized access of file: " + fileId + " by user " + session.UserId)
		c.JSON(500, gin.H{"err": "Unauthorized file access"})
		return
	}

	s.mu.Lock()
	err = s.db.DeleteMetadataById(meta.Id)
	if err != nil {
		s.logger.Error("Unable to delete file with id: " + fileId + " with err: " + err.Error())
		err = s.db.InsertMetaData(meta)
		if err != nil {
			s.logger.Error("Unable to retrieve file with id: " + fileId + " with err: " + err.Error())
			c.JSON(500, gin.H{"err": "Unable to atomically delete file + lost file access"})
			return
		}
		c.JSON(200, gin.H{"failure": "Failed to delete file: " + fileId + " but file is preserved."})
		return
	}

	_ = s.db.MarkBlobDelete(meta.Blob)
	s.mu.Unlock()

	c.JSON(200, gin.H{"success": "Deleted file with id: " + fileId})
}

func (s *Server) handleDeleteDir(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		s.logger.Error("Unauthorized session: " + authKey)
		c.JSON(500, gin.H{"err": "Invalid Authorization or missing session"})
		return
	}

	dir := c.Param("dir")

	s.mu.RLock()
	metas, err := s.db.GetMetaDataByDir(dir)
	if err != nil {
		s.logger.Error("Unable to find file with id: " + dir + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to find file"})
		return
	}
	s.mu.RUnlock()

	bak := make([]types.Metadata, 0)
	hasErr := false

	s.mu.Lock()
	for _, meta := range metas {
		bak = append(bak, meta)
		if meta.UserId != session.UserId {
			s.logger.Warn("Prevented Unauthorized access of file: " + dir + " by user " + session.UserId)
			hasErr = true
			break
		}

		err = s.db.DeleteMetadataById(meta.Id)
		if err != nil {
			s.logger.Error("Unable to delete file with id: " + dir + " with err: " + err.Error())
			hasErr = true
			break
		}
	}

	for _, m := range bak {
		if hasErr {
			err = s.db.InsertMetaData(m)
		} else {
			err = s.db.MarkBlobDelete(m.Blob)
		}
	}
	s.mu.Unlock()
	if err != nil {
		s.logger.Error("Unabled to recover from dir deletion operation failure: " + dir + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Delete failed unatomically"})
		return
	}

	if hasErr {
		c.JSON(200, gin.H{"failure": "Failed to delete dir: " + dir + " but files are preserved."})
	} else {
		c.JSON(200, gin.H{"success": "Deleted dir: " + dir})
	}
}

// DeleteFreeSpace interfaces with the fs handler, db and cache
// to freeup deleted space from buckets
// This is a very costly operation
func (s *Server) DeleteFreeSpace() error {
	// FIXME: Make me a admin route
	// TODO: Setup admin routing system

	buckets := s.handler.Buckets()
	s.logger.Info("Starting prune operation: Costly brace for impact")
	for id, bucket := range buckets {
		blobs, err := s.db.GetBlobsInBucket(id)
		if err != nil {
			return err
		}
		if len(blobs) == 0 {
			s.logger.Warn("No blobs found in bucket: " + bucket.Name() + "; Skipping")
			continue
		}
		err = s.cache.DeleteBlobs(blobs)
		if err != nil {
			s.logger.Warn("Error in Invalidating cache: " + err.Error())
		}
		nb, err := s.handler.FreeSpace(bucket, blobs)
		if err != nil {
			s.logger.Error("FS Handler unable to freespace in bucket: " + bucket.Name() + "; Non Atomic Err")
			return err
		}

		for _, b := range nb {
			var err error
			if b.Deleted {
				err = s.db.DeleteBlobById(b.Id)
				s.logger.Info("Deleted blob with id: " + b.Id)
			} else {
				err = s.db.ChangeBlobStart(b.Id, b.Start)
			}
			if err != nil {
				s.logger.Error("Store unable to update bucket info: " + bucket.Name() + "; Non Atomic Err")
				return err
			}
		}
	}
	return nil
}
