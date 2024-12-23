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
	imageinfo, format, err := image.DecodeConfig(uploadfile)
	log.Debug(ctx, fmt.Sprintf("%#v", imageinfo))
	log.Debug(ctx, format)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	//nolint:err113
	if !utils.StringArrayContains(val.imagetype, format) {
		return fmt.Errorf("invalid Image type: %s", strings.Join(val.imagetype, "/"))
	}
	sizeerrormsg := make([]string, 0)
	if val.maxwidth != 0 && imageinfo.Width > val.maxwidth {
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image width: %d px", val.maxwidth))
	}
	if val.maxheight != 0 && imageinfo.Height > val.maxheight {
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image height: %d px", val.maxheight))
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(uploadfile); err != nil {
		return err
	}
	log.Debug(ctx, val.maxfilesize)
	log.Debug(ctx, float64(val.maxfilesize)/1024/1024, 2)                   //nolint:mnd
	log.Debug(ctx, utils.RoundFloat(float64(val.maxfilesize)/1024/1024, 2)) //nolint:mnd
	if val.maxfilesize != 0 && buf.Len() > val.maxfilesize {
		//nolint:mnd
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image filesize: %v MB", utils.RoundFloat(float64(val.maxfilesize)/1024/1024, 2)))
	}
	if len(sizeerrormsg) > 0 {
		//nolint:err113
		return errors.New(strings.Join(sizeerrormsg, "/"))
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
