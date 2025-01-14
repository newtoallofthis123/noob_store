package fs

import (
	"github.com/newtoallofthis123/noob_store/types"
	"github.com/newtoallofthis123/ranhash"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) NewUser(email, password string) (*types.User, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return nil, err
	}

	user := types.User{
		Id:       ranhash.GenerateRandomString(8),
		Email:    email,
		Password: passHash,
	}

	err = h.store.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *Handler) NewSession(userId string) (*types.Session, error) {
	session := types.Session{
		Id:     ranhash.GenerateRandomString(8),
		UserId: userId,
	}

	err := h.store.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
