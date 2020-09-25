package utils

import (
	"bytes"
	"io"
	"net/http"
	"os"
)

//GetContentTypeByReadSeeker is get content type by Readseeker
func GetContentTypeByReadSeeker(reader io.ReadSeeker) string {
	reader.Seek(0, os.SEEK_SET)
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return http.DetectContentType(buf.Bytes())
}
