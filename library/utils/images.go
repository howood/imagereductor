package utils

import (
	"bytes"
	"io"
	"net/http"
)

// GetContentTypeByReadSeeker is get content type by Readseeker.
func GetContentTypeByReadSeeker(reader io.ReadSeeker) (string, error) {
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(buf.Bytes()), nil
}
