package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/newtoallofthis123/noob_store/types"
	"github.com/newtoallofthis123/noob_store/utils"
	"github.com/newtoallofthis123/ranhash"
)

// NewBucket initializes a new bucket from a bucket path
func NewBucket(bucketPath string) (*Bucket, error) {
	f, err := os.OpenFile(bucketPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	endPos, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &Bucket{id: bucketPath, file: f, size: uint64(stat.Size()), path: bucketPath, pos: uint64(endPos)}, nil
}

// writeData writes the data to the bucket
func (b *Bucket) writeData(name string, content []byte) (uint64, error) {
	contentLength := uint64(len(content))

	ogPos := b.pos
	n, err := b.file.Write(content)
	if err != nil {
		return 0, err
	}
	if n != int(contentLength) {
		return 0, fmt.Errorf("Data integrity spoilt for: " + name)
	}

	endPos, err := b.file.Seek(0, io.SeekEnd)
	b.pos = uint64(endPos)

	return uint64(endPos - int64(ogPos)), err
}

func (b *Bucket) deleteBlobs(blobs []types.Blob) ([]types.Blob, error) {
	bak := make([]byte, 0)
	start := uint64(0)
	newBlobs := make([]types.Blob, 0)

	sort.Slice(blobs, func(i, j int) bool {
		return blobs[i].Start < blobs[j].Start
	})

	file, err := os.OpenFile(b.file.Name(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	err = file.Truncate(0)
	if err != nil {
		return nil, err
	}

	for _, blob := range blobs {
		nBlob := blob
		if blob.Deleted {
			newBlobs = append(newBlobs, nBlob)
			continue
		}
		nBlob.Start = start
		bak = append(bak, blob.Content...)
		start += blob.Size
		newBlobs = append(newBlobs, nBlob)
	}

	_, err = file.Write(bak)
	if err != nil {
		return nil, err
	}

	return newBlobs, nil
}

// NewBlob returns a new blob for a filename and it's content
func (b *Bucket) NewBlob(name string, content []byte) (types.Blob, error) {
	id := ranhash.GenerateRandomString(8)
	checksum := utils.CalHash(content)

	bucketId := b.file.Name()
	start := b.pos

	size, err := b.writeData(name, content)
	b.pos += size

	blob := types.Blob{
		Id:       id,
		Name:     name,
		Bucket:   bucketId,
		Size:     size,
		Start:    start,
		Checksum: checksum,
	}
	if err != nil {
		return types.Blob{}, err
	}

	return blob, nil
}

// DiscoverBuckets discovers all viable buckets in a given path
func DiscoverBuckets(basePath string) ([]string, error) {
	dir, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0)

	for _, f := range dir {
		if !f.IsDir() && strings.Contains(f.Name(), ".bucket") {
			buckets = append(buckets, filepath.Join(basePath, f.Name()))
		}
	}

	return buckets, nil
}

// GenerateBuckets generates some bucket names to a given basePath
func GenerateBuckets(basePath string, n uint8) []string {
	buckets := make([]string, 0)

	i := 0
	for i < int(n) {
		buckets = append(buckets, filepath.Join(basePath, strings.ToLower(ranhash.GenerateRandomString(8))+".bucket"))
		i++
	}

	return buckets
}

func (b *Bucket) Name() string {
	return b.file.Name()
}
