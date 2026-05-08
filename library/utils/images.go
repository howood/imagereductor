package utils

import (
	"io"
	"net/http"
)

// GetContentTypeByReadSeeker is get content type by Readseeker.
func GetContentTypeByReadSeeker(reader io.ReadSeeker) (string, error) {
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	buf := make([]byte, 512) //nolint:mnd
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}
