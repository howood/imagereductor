package repository

// CachedContentRepository interface.
type CachedContentRepository interface {
	Set(contentType, lastModified string, content []byte)
	GetContentType() string
	GetLastModified() string
	GetContent() []byte
	GobEncode() ([]byte, error)
	GobDecode(buf []byte) error
}
