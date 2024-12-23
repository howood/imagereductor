package actor

import (
	"bytes"
	"encoding/gob"

	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/domain/repository"
)

// CachedContentOperator struct.
type CachedContentOperator struct {
	repository.CachedContentRepository
}

// NewCachedContentOperator creates a new CachedContentRepository.
func NewCachedContentOperator() *CachedContentOperator {
	return &CachedContentOperator{&cachedContentCreator{}}
}

// cachedContentCreator struct.
type cachedContentCreator struct {
	chachedData entity.CachedContent
}

// Set sets contentType,lastModified and  content to  cahced content.
func (e *cachedContentCreator) Set(contentType, lastModified string, content []byte) {
	e.chachedData.ContentType = contentType
	e.chachedData.LastModified = lastModified
	e.chachedData.Content = content
}

// GetContentType returns contenttype of cahced content.
func (e *cachedContentCreator) GetContentType() string {
	return e.chachedData.ContentType
}

// GetLastModified returns lastmodified of cahced content.
func (e *cachedContentCreator) GetLastModified() string {
	return e.chachedData.LastModified
}

// GetContent returns content of cahced content.
func (e *cachedContentCreator) GetContent() []byte {
	return e.chachedData.Content
}

// GobEncode serialized cached data to bytes.
func (e *cachedContentCreator) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	if err := encoder.Encode(e.chachedData.ContentType); err != nil {
		return nil, err
	}
	if err := encoder.Encode(e.chachedData.LastModified); err != nil {
		return nil, err
	}
	if err := encoder.Encode(e.chachedData.Content); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode decode bytes to cached data.
func (e *cachedContentCreator) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	if err := decoder.Decode(&e.chachedData.ContentType); err != nil {
		return err
	}
	if err := decoder.Decode(&e.chachedData.LastModified); err != nil {
		return err
	}
	if err := decoder.Decode(&e.chachedData.Content); err != nil {
		return err
	}
	return nil
}
