package api

import (
	"github.com/gin-gonic/gin"
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

	user, err := s.handler.NewUser(email, password)
	if err != nil {
		c.JSON(500, gin.H{"err": "Error creating user: " + err.Error()})
		return
	}

	session, err := s.handler.NewSession(user.Id)
	if err != nil {
		c.JSON(500, gin.H{"err": "Error creating session: " + err.Error()})
		return
	}

	err = s.cache.InsertSession(*session)
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

	session, err := s.handler.NewSession(user.Id)
	if err != nil {
		c.JSON(500, gin.H{"err": "Error creating session: " + err.Error()})
		return
	}

	err = s.cache.InsertSession(*session)
	if err != nil {
		s.logger.Error("Error in inserting session to cache" + err.Error())
	}

	c.JSON(200, session)
}
