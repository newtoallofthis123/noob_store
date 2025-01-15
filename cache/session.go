package cache

import (
	"encoding/json"
	"time"

	"github.com/newtoallofthis123/noob_store/types"
)

func (c *Cache) InsertSession(session types.Session) error {
	session_encoded, err := json.Marshal(session)
	if err != nil {
		return err
	}

	err = c.r.Set(c.ctx, session.Id, string(session_encoded), time.Hour*24*7).Err()
	return err
}

func (c *Cache) GetSession(sessionId string) (types.Session, error) {
	var session types.Session

	session_encoded, err := c.r.Get(c.ctx, sessionId).Result()
	if err != nil {
		return types.Session{}, err
	}

	err = json.Unmarshal([]byte(session_encoded), &session)
	if err != nil {
		return types.Session{}, err
	}

	return session, nil
}
