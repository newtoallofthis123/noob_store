package fs

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/newtoallofthis123/noob_store/types"
	"github.com/newtoallofthis123/ranhash"
)

func NewBucket(bucketPath string) (*Bucket, error) {
	f, err := os.OpenFile(bucketPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}

	endPos, err := f.Seek(0, io.SeekEnd)

	return &Bucket{id: bucketPath, file: f, path: bucketPath, pos: uint64(endPos)}, nil
}

func parseContent(content string) (string, string) {
	strSplit := strings.SplitN(content, "--", 4)
	return strSplit[1], strSplit[2]
}

func (b *Bucket) writeData(name string, content []byte) (uint64, error) {
	// TODO: Handle the case of the len(content) exceeding length of int
	contentLength := uint64(len(content))

	ogPos := b.pos
	err := b.writeHeader(name, contentLength)
	if err != nil {
		return 0, err
	}

	n, err := b.file.Write(content)
	if err != nil {
		return 0, err
	}
	if n != int(contentLength) {
		return 0, fmt.Errorf("Data integrity spoilt")
	}

	err = b.writeFooter()

	endPos, err := b.file.Seek(0, io.SeekEnd)
	b.pos = uint64(endPos)

	return uint64(endPos - int64(ogPos)), err
}

func (b *Bucket) writeHeader(name string, size uint64) error {
	// WARNING: We assume that the writing of the header is always valid

	_, err := b.file.Write([]byte(fmt.Sprintf("--%d%s&%d&%d--", VERSION, name, size, b.pos)))
	return err
}

func (b *Bucket) writeFooter() error {
	// WARNING: We assume that the writing of the footer is always valid
	_, err := b.file.Write([]byte("--&--"))
	return err
}

func (b *Bucket) NewBlob(name string, content []byte) (*types.Blob, error) {
	id := ranhash.GenerateRandomString(8)
	hash := md5.New()

	lim := 100
	// Only take the checksum of 1024 bytes cause we don't want the checksum to be too big
	toHash := make([]byte, len(content))
	copy(toHash, content)

	if len(toHash) > lim {
		toHash = toHash[:lim]
	}

	checksum := hash.Sum(toHash)
	bucketId := b.file.Name()
	start := b.pos

	size, err := b.writeData(name, content)

	blob := types.Blob{
		Id:       id,
		Name:     name,
		Bucket:   bucketId,
		Size:     size,
		Start:    start,
		Checksum: checksum,
	}
	if err != nil {
		return nil, err
	}

	return &blob, nil
}

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

func GenerateBuckets(basePath string, n uint8) []string {
	buckets := make([]string, 0)

	i := 0
	for i < int(n) {
		buckets = append(buckets, filepath.Join(basePath, strings.ToLower(ranhash.GenerateRandomString(8))+".bucket"))
		i++
	}

	return buckets
}
