package cache

import (
	"encoding/json"
	"time"

	"github.com/newtoallofthis123/noob_store/types"
)

func (c *Cache) InsertMetadata(meta types.Metadata) error {
	meta_encoded, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	cmd := c.r.Set(c.ctx, meta.Id, string(meta_encoded), time.Hour*24*7)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}

func (c *Cache) GetMetadata(metaId string) (types.Metadata, error) {
	var meta types.Metadata

	meta_encoded, err := c.r.Get(c.ctx, metaId).Result()
	if err != nil {
		return types.Metadata{}, err
	}

	err = json.Unmarshal([]byte(meta_encoded), &meta)
	if err != nil {
		return types.Metadata{}, err
	}

	return meta, nil
}
