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
	// ImageTypeJpeg is type of Jpeg
	ImageTypeJpeg = "jpeg"
	// ImageTypeGif is type of Gif
	ImageTypeGif = "gif"
	// ImageTypePng is type of PNG
	ImageTypePng = "png"
	// ImageTypeBmp is type of BMP
	ImageTypeBmp = "bmp"
	// ImageTypeTiff is type of TIFF
	ImageTypeTiff = "tiff"
)

// ImageTypeList is list of image types
var ImageTypeList = []string{ImageTypeJpeg, ImageTypeGif, ImageTypePng, ImageTypeBmp, ImageTypeTiff}

// ImageValidator struct
type ImageValidator struct {
	imagetype   []string
	maxwidth    int
	maxheight   int
	maxfilesize int
	ctx         context.Context
}

// NewImageValidator creates a new ImageValidator
func NewImageValidator(ctx context.Context, imagetype []string, maxwidth, maxheight, maxfilesize int) *ImageValidator {
	I := &ImageValidator{
		imagetype:   imagetype,
		maxwidth:    maxwidth,
		maxheight:   maxheight,
		maxfilesize: maxfilesize,
		ctx:         ctx,
	}
	I.convertImageType()
	return I
}

// Validate process to validate  uploadfile
func (val *ImageValidator) Validate(uploadfile io.Reader) error {
	imageinfo, format, err := image.DecodeConfig(uploadfile)
	log.Debug(val.ctx, fmt.Sprintf("%#v", imageinfo))
	log.Debug(val.ctx, format)
	if err != nil {
		return errors.New(err.Error())
	}
	if utils.StringArrayContains(val.imagetype, format) == false {
		return fmt.Errorf("Invalid Image type: %s", strings.Join(val.imagetype, "/"))
	}
	sizeerrormsg := make([]string, 0)
	if val.maxwidth != 0 && imageinfo.Width > val.maxwidth {
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image width: %d px", val.maxwidth))
	}
	if val.maxheight != 0 && imageinfo.Height > val.maxheight {
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image height: %d px", val.maxheight))
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(uploadfile)
	log.Debug(val.ctx, val.maxfilesize)
	log.Debug(val.ctx, float64(val.maxfilesize)/1024/1024, 2)
	log.Debug(val.ctx, utils.RoundFloat(float64(val.maxfilesize)/1024/1024, 2))
	if val.maxfilesize != 0 && buf.Len() > val.maxfilesize {
		sizeerrormsg = append(sizeerrormsg, fmt.Sprintf("Over Image filesize: %v MB", utils.RoundFloat(float64(val.maxfilesize)/1024/1024, 2)))
	}
	if len(sizeerrormsg) > 0 {
		return errors.New(strings.Join(sizeerrormsg, "/"))
	}
	return nil
}

func (val *ImageValidator) convertImageType() {
	replacelist := make([]string, 0)
	for _, imagetype := range val.imagetype {
		if utils.StringArrayContains(ImageTypeList, imagetype) == true {
			replacelist = append(replacelist, imagetype)
		}
	}
	val.imagetype = replacelist
}
