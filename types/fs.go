package types

type Blob struct {
	Id         string
	Name       string
	Bucket     string
	Start      uint64
	Content    []byte
	Size       uint64
	Checksum   []byte
	Deleted    bool
	Created_at string
}

type BlobRes struct {
	Id         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
	Size       uint64 `json:"size,omitempty"`
	Created_at string `json:"created_at,omitempty"`
}

func MapBlobToRes(blob Blob) BlobRes {
	return BlobRes{
		Id:         blob.Id,
		Name:       blob.Name,
		Bucket:     blob.Bucket,
		Size:       blob.Size,
		Created_at: blob.Created_at,
	}
}

type Metadata struct {
	Id         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Path       string `json:"path,omitempty"`
	Parent     string `json:"parent,omitempty"`
	Mime       string `json:"mime,omitempty"`
	Blob       string `json:"blob,omitempty"`
	Created_at string `json:"created_at,omitempty"`
}
