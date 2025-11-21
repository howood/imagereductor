package validator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"strings"

	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/library/utils"
)

const (
	// ImageTypeJpeg is type of Jpeg.
	ImageTypeJpeg = "jpeg"
	// ImageTypeGif is type of Gif.
	ImageTypeGif = "gif"
	// ImageTypePng is type of PNG.
	ImageTypePng = "png"
	// ImageTypeBmp is type of BMP.
	ImageTypeBmp = "bmp"
	// ImageTypeTiff is type of TIFF.
	ImageTypeTiff = "tiff"
)

// Sentinel errors for image validation.
var (
	ErrInvalidImageType   = errors.New("invalid image type")
	ErrImageSizeExceeded  = errors.New("image size exceeded")
	ErrImageDecodeConfig  = errors.New("failed to decode image config")
	ErrImageReadRemaining = errors.New("failed to read remaining image data")
)

// imageTypeList is list of image types.
//
//nolint:gochecknoglobals
var imageTypeList = []string{ImageTypeJpeg, ImageTypeGif, ImageTypePng, ImageTypeBmp, ImageTypeTiff}

// ImageValidator struct.
type ImageValidator struct {
	imagetype   []string
	maxwidth    int
	maxheight   int
	maxfilesize int
}

// NewImageValidator creates a new ImageValidator.
func NewImageValidator(imagetype []string, maxwidth, maxheight, maxfilesize int) *ImageValidator {
	I := &ImageValidator{
		imagetype:   imagetype,
		maxwidth:    maxwidth,
		maxheight:   maxheight,
		maxfilesize: maxfilesize,
	}
	I.convertImageType()
	return I
}

// Validate process to validate  uploadfile.
//
//nolint:cyclop
func (val *ImageValidator) Validate(ctx context.Context, uploadfile io.Reader) error {
	// Use TeeReader to read file data once for both DecodeConfig and size validation
	buf := new(bytes.Buffer)
	teeReader := io.TeeReader(uploadfile, buf)

	// Decode config from tee reader (reads and writes to buffer simultaneously)
	imageinfo, format, err := image.DecodeConfig(teeReader)
	log.Debug(ctx, fmt.Sprintf("%#v", imageinfo))
	log.Debug(ctx, format)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrImageDecodeConfig, err)
	}

	if !utils.StringArrayContains(val.imagetype, format) {
		return fmt.Errorf("%w: got %s, allowed: %s", ErrInvalidImageType, format, strings.Join(val.imagetype, ","))
	}

	sizeerrormsg := make([]string, 0)
	if val.maxwidth != 0 && imageinfo.Width > val.maxwidth {
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image width: %d px", val.maxwidth))
	}
	if val.maxheight != 0 && imageinfo.Height > val.maxheight {
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image height: %d px", val.maxheight))
	}

	// Read remaining data if DecodeConfig didn't consume all
	if _, err := io.Copy(buf, teeReader); err != nil {
		return fmt.Errorf("%w: %w", ErrImageReadRemaining, err)
	}

	filesize := buf.Len()
	log.Debug(ctx, val.maxfilesize)
	log.Debug(ctx, float64(val.maxfilesize)/1024/1024, 2)                   //nolint:mnd
	log.Debug(ctx, utils.RoundFloat(float64(val.maxfilesize)/1024/1024, 2)) //nolint:mnd
	if val.maxfilesize != 0 && filesize > val.maxfilesize {
		//nolint:mnd
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image filesize: %v MB", utils.RoundFloat(float64(val.maxfilesize)/1024/1024, 2)))
	}
	if len(sizeerrormsg) > 0 {
		return fmt.Errorf("%w: %s", ErrImageSizeExceeded, strings.Join(sizeerrormsg, "; "))
	}
	return nil
}

func (val *ImageValidator) convertImageType() {
	replacelist := make([]string, 0)
	for _, imagetype := range val.imagetype {
		if utils.StringArrayContains(imageTypeList, imagetype) {
			replacelist = append(replacelist, imagetype)
		}
	}
	val.imagetype = replacelist
}
