package actor

import (
	"bytes"
	"encoding/gob"
)

type CachedContent struct {
	ContentType  string
	LastModified string
	Content      []byte
}

// シリアライズ用のメソッド
// レシーバ(e)の値をシリアライズしてbyte配列にする
func (e *CachedContent) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	if err := encoder.Encode(e.ContentType); err != nil {
		return nil, err
	}
	if err := encoder.Encode(e.LastModified); err != nil {
		return nil, err
	}
	if err := encoder.Encode(e.Content); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (e *CachedContent) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	if err := decoder.Decode(&e.ContentType); err != nil {
		return err
	}
	if err := decoder.Decode(&e.LastModified); err != nil {
		return err
	}
	if err := decoder.Decode(&e.Content); err != nil {
		return err
	}
	return nil
}
