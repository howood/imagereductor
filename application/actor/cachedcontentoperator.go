package actor

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/domain/repository"
)

// CachedContentOperator struct
type CachedContentOperator struct {
	chachedData entity.CachedContent
	ctx         context.Context
}

// NewCachedContentOperator creates a new CachedContentRepository
func NewCachedContentOperator() repository.CachedContentRepository {
	return &CachedContentOperator{}
}

//Set sets contentType,lastModified and  content to  cahced content
func (e *CachedContentOperator) Set(contentType, lastModified string, content []byte) {
	e.chachedData.ContentType = contentType
	e.chachedData.LastModified = lastModified
	e.chachedData.Content = content
}

// Get returns  contenttype of cahced content
func (e *CachedContentOperator) GetContentType() string {
	return e.chachedData.ContentType
}
func (e *CachedContentOperator) GetLastModified() string {
	return e.chachedData.LastModified
}

func (e *CachedContentOperator) GetContent() []byte {
	return e.chachedData.Content
}

// GobEncode serialized cached data to bytes
func (e *CachedContentOperator) GobEncode() ([]byte, error) {
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

// GobDecode decode bytes to cached data
func (e *CachedContentOperator) GobDecode(buf []byte) error {
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
