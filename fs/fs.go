package fs

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/newtoallofthis123/noob_store/db"
	"github.com/newtoallofthis123/noob_store/types"
	"github.com/newtoallofthis123/ranhash"
	"github.com/zRedShift/mimemagic"
)

const VERSION = 0

// Bucket represents a bucket file
type Bucket struct {
	file *os.File
	path string
	id   string
	pos  uint64
}

// Handler handles delegation of buckets, store and logger
type Handler struct {
	buckets map[string]*Bucket
	logger  *slog.Logger
}

// NewHandler initializes a new handler
func NewHandler(bucketPaths []string, store *db.Store, logger *slog.Logger) Handler {
	buckets := make(map[string]*Bucket, 0)
	for _, path := range bucketPaths {
		b, err := NewBucket(path)
		if err != nil {
			logger.Error("Unable to create bucket with path: " + path)
		}

		buckets[path] = b
	}

	return Handler{
		buckets: buckets,
		logger:  logger,
	}
}

// selectBucket selects and returns a bucket randomly
func (h *Handler) selectBucket() *Bucket {
	ran := rand.Intn(len(h.buckets))
	c := 0
	for _, b := range h.buckets {
		if c == ran {
			return b
		}
		c++
	}
	return nil
}

// Insert inserts a new blob into a random bucket
func (h *Handler) Insert(fullPath string, content []byte, userId string) (types.Blob, types.Metadata, error) {
	b := h.selectBucket()
	meta := NewMetaData(fullPath, userId)
	blob, err := b.NewBlob(meta.Path, content)
	if err != nil {
		h.logger.Error("Error appending blob: " + err.Error())
		return types.Blob{}, types.Metadata{}, err
	}
	h.logger.Debug("Appened Blob to bucket: " + meta.Path)
	meta.Blob = blob.Id

	return blob, meta, nil
}

// NewMetaData returns a new metadata struct
func NewMetaData(fullPath string, userId string) types.Metadata {
	fullPath = filepath.Clean(fullPath)
	name := filepath.Base(fullPath)
	parent := filepath.Dir(fullPath)
	mime := ""
	m, err := mimemagic.MatchFilePath(name)
	if err == nil {
		mime = m.MediaType()
	}

	return types.Metadata{
		Id:     ranhash.GenerateRandomString(8),
		Name:   name,
		UserId: userId,
		Parent: parent,
		Mime:   mime,
		Path:   fullPath,
	}
}

// fillBlob fills in the details of a blob
func (h *Handler) fillBlob(blob *types.Blob) error {
	b := h.buckets[blob.Bucket]

	file, err := os.Open(b.path)
	if err != nil {
		h.logger.Error("Unable to open file: " + err.Error())
		return err
	}

	buff := make([]byte, blob.Size)

	fmt.Println(len(buff))

	n, err := file.ReadAt(buff, int64(blob.Start))
	if err != nil {
		h.logger.Error("Error reading bucket: " + err.Error())
		return err
	}

	// _, content := parseContent(string(buff[:n]))

	blob.Content = []byte(buff[:n])

	return nil
}

// Get gets a blob from the buckets by using the given blob path
func (h *Handler) Get(blob *types.Blob) error {
	err := h.fillBlob(blob)
	if err != nil {
		h.logger.Error("Error filling blob: " + blob.Name)
		return err
	}

	return nil
}

// GetDir gets all the blobs in the dir
func (h *Handler) GetDir(blobs []*types.Blob) error {
	for _, b := range blobs {
		err := h.fillBlob(b)
		if err != nil {
			h.logger.Error("Error getting blob: " + b.Name)
			continue
		}
	}

	return nil
}
