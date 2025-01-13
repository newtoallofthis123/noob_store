package types

type Blob struct {
	Id         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
	Start      uint64 `json:"offset,omitempty"`
	Content    []byte `json:"content,omitempty"`
	Size       uint64 `json:"size,omitempty"`
	Checksum   []byte `json:"checksum,omitempty"`
	Deleted    bool   `json:"deleted,omitempty"`
	Created_at string `json:"created_at,omitempty"`
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
