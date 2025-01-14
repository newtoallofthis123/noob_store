package types

type Blob struct {
	Id        string
	Name      string
	Bucket    string
	Start     uint64
	Content   []byte
	Size      uint64
	Checksum  []byte
	Deleted   bool
	CreatedAt string
}

type BlobRes struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	Size      uint64 `json:"size,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

func MapBlobToRes(blob Blob) BlobRes {
	return BlobRes{
		Id:        blob.Id,
		Name:      blob.Name,
		Bucket:    blob.Bucket,
		Size:      blob.Size,
		CreatedAt: blob.CreatedAt,
	}
}

type Metadata struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Path      string `json:"path,omitempty"`
	Parent    string `json:"parent,omitempty"`
	Mime      string `json:"mime,omitempty"`
	UserId    string `json:"user_id,omitempty"`
	Blob      string `json:"blob,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type User struct {
	Id        string `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  []byte `json:"password,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type Session struct {
	Id        string `json:"id,omitempty"`
	UserId    string `json:"user_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}
