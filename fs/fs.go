package fs

import (
	"fmt"
	"log/slog"
	"math/rand"
	"mime"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/newtoallofthis123/noob_store/db"
	"github.com/newtoallofthis123/noob_store/types"
	"github.com/newtoallofthis123/ranhash"
	"github.com/zRedShift/mimemagic"
)

const VERSION = 0

// THRESHOLD is set to about 1 GB
const THRESHOLD = 1024 * 1024 * 1024

// Bucket represents a bucket file
type Bucket struct {
	file *os.File
	path string
	size uint64
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

// selectRandomBucket selects and returns a bucket randomly
func (h *Handler) selectRandomBucket() *Bucket {
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

// selectFirstBucket selects the first bucket from the available buckets.
func (h *Handler) selectFirstBucket() *Bucket {
	for _, b := range h.buckets {
		return b
	}

	return nil
}

// selectBestBucket selects the best bucket from the available buckets.
func (h *Handler) selectBestBucket() *Bucket {
	lowest := h.selectFirstBucket()
	for _, b := range h.buckets {
		if b.size < lowest.size {
			lowest = b
		}
	}

	return lowest
}

// selectBucket selects bucket on a random chance of selecting the best bucket and selecting a random bucket
// There is a 66% chance of selecting a random bucket and 33% chance of selecting the best bucket
func (h *Handler) selectBucket(minSize uint64) *Bucket {
	ran := rand.Intn(3)
	var b *Bucket
	if ran == 1 {
		b = h.selectBestBucket()
	} else {
		b = h.selectRandomBucket()
	}
	if THRESHOLD-b.size < minSize {
		b = h.selectBucket(minSize)
	} else if b.size >= THRESHOLD {
		b = h.selectBucket(minSize)
	}
	return b
}

// Insert inserts a new blob into a random bucket
func (h *Handler) Insert(fullPath string, content []byte, userId string) (types.Blob, types.Metadata, error) {
	b := h.selectBucket(uint64(len(content)))
	meta := NewMetaData(fullPath, userId)
	if meta.Mime == "" {
		meta.Mime = mimemagic.MatchMagic(content).MediaType()
	}

	blob, err := b.NewBlob(meta.Path, content)
	if err != nil {
		h.logger.Error("Error appending blob: " + err.Error())
		return types.Blob{}, types.Metadata{}, err
	}
	meta.Blob = blob.Id

	return blob, meta, nil
}

// NewMetaData returns a new metadata struct
func NewMetaData(fullPath string, userId string) types.Metadata {
	fullPath = filepath.Clean(fullPath)
	name := filepath.Base(fullPath)
	parent := filepath.Dir(fullPath)
	mime := mime.TypeByExtension(filepath.Ext(name))
	if mime == "" {
		mime = mimemagic.MatchGlob(name).MediaType()
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

// LogBucketsInfo logs information about the bucket at the INFO level
func (h *Handler) LogBucketsInfo() {
	for i, b := range h.buckets {
		stat, _ := b.file.Stat()
		buckStr := fmt.Sprintf("name: %s | size: %s | filled: %t", i, humanize.Bytes(uint64(stat.Size())), stat.Size() > THRESHOLD)
		h.logger.Info(buckStr)
	}
}
