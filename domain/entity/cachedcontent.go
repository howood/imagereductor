package entity

type CachedContent struct {
	ContentType  string
	LastModified string
	Content      []byte
}
