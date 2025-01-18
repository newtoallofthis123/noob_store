package api

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/newtoallofthis123/noob_store/types"
	"github.com/newtoallofthis123/ranhash"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleCreateUser(c *gin.Context) {
	email, exists := c.GetPostForm("email")
	if !exists {
		c.JSON(500, gin.H{"err": "Email is needed"})
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		c.JSON(500, gin.H{"err": "Password is needed"})
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		s.logger.Error("Failure in hashing password: " + err.Error())
		c.JSON(500, gin.H{"err": "Password hashing failed with err: " + err.Error()})
		return
	}

	user := types.User{
		Id:       ranhash.GenerateRandomString(8),
		Email:    email,
		Password: string(passHash),
	}

	err = s.db.CreateUser(user)
	if err != nil {
		s.logger.Error("Failed to create user" + user.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Error creating user: " + err.Error()})
		return
	}

	session := types.Session{
		Id:     ranhash.GenerateRandomString(16),
		UserId: user.Id,
	}

	err = s.db.CreateSession(session)
	if err != nil {
		s.logger.Error("Failed to create session" + session.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Error creating session: " + err.Error()})
		return
	}

	err = s.cache.InsertSession(session)
	if err != nil {
		s.logger.Error("Error in inserting session to cache" + err.Error())
	}

	c.JSON(200, session)
}

func (s *Server) handleLoginUser(c *gin.Context) {
	email, exists := c.GetPostForm("email")
	if !exists {
		c.JSON(500, gin.H{"err": "Email is needed"})
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		c.JSON(500, gin.H{"err": "Password is needed"})
		return
	}

	user, err := s.db.GetUserByEmail(email)
	if err != nil {
		c.JSON(500, gin.H{"err": "User not found"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		s.logger.Error("Matching passwords not found for")
		c.JSON(500, gin.H{"err": "Authorization failed"})
		return
	}

	session := types.Session{
		Id:     ranhash.GenerateRandomString(16),
		UserId: user.Id,
	}
	err = s.db.CreateSession(session)
	if err != nil {
		s.logger.Error("Failed to create session" + session.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Error creating session: " + err.Error()})
		return
	}

	err = s.cache.InsertSession(session)
	if err != nil {
		s.logger.Error("Error in inserting session to cache" + err.Error())
	}

	c.JSON(200, session)
}

// buildDir constructs a multi-directory system as a nested map.
func buildDir(files []types.Metadata) map[string]interface{} {
	dirStructure := make(map[string]interface{})

	for _, f := range files {
		components := strings.Split(f.Path, "/")
		current := dirStructure

		for i, comp := range components {
			if i == len(components)-1 {
				current[comp] = nil
			} else {
				if _, exists := current[comp]; !exists {
					current[comp] = make(map[string]interface{})
				}
				current = current[comp].(map[string]interface{})
			}
		}
	}

	return dirStructure
}

func (s *Server) handleUserLs(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		s.logger.Error("Unauthorized session: " + authKey)
		c.JSON(500, gin.H{"err": "Invalid Authorization or missing session"})
		return
	}

	user, err := s.db.GetUser(session.UserId)
	if err != nil {
		s.logger.Error("Unable to get user with id: " + user.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "No valid user found for sessionId " + session.Id})
		return
	}

	dir, exists := c.GetQuery("dir")
	var metas []types.Metadata

	if exists {
		dir = filepath.Clean(dir)
		metas, err = s.db.GetMetadataDirByUser(user.Id, dir)
	} else {
		metas, err = s.db.GetMetadatasByUser(user.Id)
	}

	if err != nil {
		s.logger.Error("Unable to fetch user files for userId: " + user.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to fetch user files"})
		return
	}

	s.logger.Debug("Successfully returned user ls")
	c.JSON(200, metas)
}

func (s *Server) handleUserPathLs(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	session, exists := s.checkAuth(authKey)
	if !exists {
		s.logger.Error("Unauthorized session: " + authKey)
		c.JSON(500, gin.H{"err": "Invalid Authorization or missing session"})
		return
	}

	user, err := s.db.GetUser(session.UserId)
	if err != nil {
		s.logger.Error("Unable to get user with id: " + user.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "No valid user found for sessionId " + session.Id})
		return
	}

	dir, exists := c.GetQuery("dir")
	var metas []types.Metadata

	if exists {
		dir = filepath.Clean(dir)
		metas, err = s.db.GetMetadataDirByUser(user.Id, dir)
	} else {
		metas, err = s.db.GetMetadatasByUser(user.Id)
	}

	if err != nil {
		s.logger.Error("Unable to fetch user files for userId: " + user.Id + " with err: " + err.Error())
		c.JSON(500, gin.H{"err": "Unable to fetch user files"})
		return
	}

	s.logger.Debug("Successfully returned user ls")
	c.JSON(200, buildDir(metas))
}
