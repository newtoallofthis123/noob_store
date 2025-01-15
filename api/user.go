package api

import (
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
		c.JSON(500, gin.H{"err": "Password hashing failed with err: " + err.Error()})
		return
	}

	user := types.User{
		Id:       ranhash.GenerateRandomString(8),
		Email:    email,
		Password: passHash,
	}

	err = s.db.CreateUser(user)
	if err != nil {
		c.JSON(500, gin.H{"err": "Error creating user: " + err.Error()})
		return
	}

	session := types.Session{
		Id:     ranhash.GenerateRandomString(8),
		UserId: user.Id,
	}

	err = s.db.CreateSession(session)
	if err != nil {
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

	if bcrypt.CompareHashAndPassword(user.Password, []byte(password)) != nil {
		c.JSON(500, gin.H{"err": "Authorization failed"})
		return
	}

	session := types.Session{
		Id:     ranhash.GenerateRandomString(8),
		UserId: user.Id,
	}
	err = s.db.CreateSession(session)
	if err != nil {
		c.JSON(500, gin.H{"err": "Error creating session: " + err.Error()})
		return
	}

	err = s.cache.InsertSession(session)
	if err != nil {
		s.logger.Error("Error in inserting session to cache" + err.Error())
	}

	c.JSON(200, session)
}
