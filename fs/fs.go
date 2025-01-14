package fs

import (
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

type Bucket struct {
	file *os.File
	path string
	id   string
	pos  uint64
}

type Handler struct {
	buckets map[string]*Bucket
	store   *db.Store
	logger  *slog.Logger
}

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
		store:   store,
		logger:  logger,
	}
}

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

func (h *Handler) Insert(fullPath string, content []byte) (*types.Blob, *types.Metadata, error) {
	b := h.selectBucket()
	meta := NewMetaData(fullPath)
	blob, err := b.NewBlob(meta.Path, content)
	if err != nil {
		h.logger.Error("Error appending blob: " + err.Error())
		return nil, nil, err
	}
	h.logger.Debug("Appened Blob to bucket: " + meta.Path)

	err = h.store.InsertBlob(blob)
	if err != nil {
		h.logger.Error("Error inserting blob to store: " + err.Error())
		return nil, nil, err
	}

	meta.Blob = blob.Id

	err = h.store.InsertMetaData(meta)
	if err != nil {
		h.logger.Error("Error inserting blob to store: " + err.Error())
		return nil, nil, err
	}

	return blob, &meta, nil
}

func NewMetaData(fullPath string) types.Metadata {
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
		Parent: parent,
		Mime:   mime,
		Path:   fullPath,
	}
}

func (h *Handler) fillBlob(blobId string) (types.Blob, error) {
	blob, err := h.store.GetBlobById(blobId)
	if err != nil {
		h.logger.Error("Error getting blob: " + blobId)
		return types.Blob{}, err
	}

	b := h.buckets[blob.Bucket]

	file, err := os.Open(b.path)
	if err != nil {
		h.logger.Error("Unable to open file: " + err.Error())
		return types.Blob{}, err
	}

	buff := make([]byte, blob.Size)

	n, err := file.ReadAt(buff, int64(blob.Start))
	if err != nil {
		h.logger.Error("Error reading bucket: " + err.Error())
		return types.Blob{}, err
	}

	_, content := parseContent(string(buff[:n]))

	blob.Content = []byte(content)

	return blob, nil
}

func (h *Handler) Get(path string) (types.Blob, error) {
	name := filepath.Base(path)
	fullPath := filepath.Clean(path)
	meta, err := h.store.GetMetaData(name, fullPath)
	if err != nil {
		h.logger.Error("Error getting metadata for: " + fullPath)
		return types.Blob{}, err
	}

	blob, err := h.fillBlob(meta.Blob)
	if err != nil {
		h.logger.Error("Error getting blob: " + blob.Name)
		return types.Blob{}, err
	}

	return blob, nil
}

func (h *Handler) GetDir(dirPath string) ([]types.Blob, error) {
	fullPath := filepath.Clean(dirPath)
	metas, err := h.store.GetMetaDataByDir(fullPath)
	if err != nil {
		return nil, err
	}

	blobs := make([]types.Blob, 0)

	for _, m := range metas {
		blob, err := h.fillBlob(m.Blob)
		if err != nil {
			h.logger.Error("Error getting blob: " + m.Blob)
			continue
		}

		blobs = append(blobs, blob)
	}

	return blobs, nil
}

func (h *Handler) GetAll() ([]types.Blob, error) {
	metas, err := h.store.GetAllFiles()
	if err != nil {
		return nil, err
	}

	blobs := make([]types.Blob, 0)

	for _, m := range metas {
		blob, err := h.fillBlob(m.Blob)
		if err != nil {
			h.logger.Error("Error getting blob: " + m.Blob)
			continue
		}

		blobs = append(blobs, blob)
	}

	return blobs, nil
}
