package cache

import (
	"encoding/json"
	"time"

	"github.com/newtoallofthis123/noob_store/types"
)

func (c *Cache) InsertBlob(blob types.Blob) error {
	blob_encoded, err := json.Marshal(blob)
	if err != nil {
		return err
	}

	cmd := c.r.Set(c.ctx, blob.Id, string(blob_encoded), time.Hour*24*7)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}

func (c *Cache) GetBlob(blobId string) (types.Blob, error) {
	var blob types.Blob

	blob_encoded, err := c.r.Get(c.ctx, blobId).Result()
	if err != nil {
		return types.Blob{}, err
	}

	err = json.Unmarshal([]byte(blob_encoded), &blob)
	if err != nil {
		return types.Blob{}, err
	}

	return blob, nil
}

func (c *Cache) DeleteBlobs(blobs []types.Blob) error {
	var err error
	for _, blobId := range blobs {
		_, err = c.r.Del(c.ctx, blobId.Id).Result()
	}
	return err
}
