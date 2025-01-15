package types

// Blob represents an object in the store
type Blob struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	Start     uint64 `json:"start,omitempty"`
	Content   []byte `json:"content,omitempty"`
	Size      uint64 `json:"size,omitempty"`
	Checksum  []byte `json:"checksum,omitempty"`
	Deleted   bool   `json:"deleted,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// BlobRes represents a user presentable blob
type BlobRes struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	Size      uint64 `json:"size,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// MapBlobToRes maps a private representation of a blob to the public one
func MapBlobToRes(blob Blob) BlobRes {
	return BlobRes{
		Id:        blob.Id,
		Name:      blob.Name,
		Bucket:    blob.Bucket,
		Size:      blob.Size,
		CreatedAt: blob.CreatedAt,
	}
}

// Metadata represents the file metadata for a blob
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

// User represents a user
type User struct {
	Id        string `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  []byte `json:"password,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// Session represents an authenticated session
type Session struct {
	Id        string `json:"id,omitempty"`
	UserId    string `json:"user_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}
