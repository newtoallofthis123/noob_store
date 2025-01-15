package cache

import (
	"encoding/json"
	"time"

	"github.com/newtoallofthis123/noob_store/types"
)

func (c *Cache) InsertUser(user types.User) error {
	user_encoded, err := json.Marshal(user)
	if err != nil {
		return err
	}

	cmd := c.r.Set(c.ctx, user.Id, string(user_encoded), time.Hour*24*7)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (c *Cache) GetUser(userId string) (types.User, error) {
	var user types.User
	user_encoded, err := c.r.Get(c.ctx, userId).Result()
	if err != nil {
		return types.User{}, err
	}
	err = json.Unmarshal([]byte(user_encoded), &user)
	if err != nil {
		return types.User{}, err
	}

	return user, nil
}
