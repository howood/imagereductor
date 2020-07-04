package actor

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/domain/repository"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"golang.org/x/image/draw"
)

// ImageOperator struct
type ImageOperator struct {
	object *entity.ImageObject
	option *entity.ImageObjectOption
}

// ImageOperatorOption is Option of ImageOperator struct
type ImageOperatorOption entity.ImageObjectOption

// NewImageOperator create ImageOperator
func NewImageOperator(contenttype string, option ImageOperatorOption) repository.ImageObjectRepository {
	objectOption := entity.ImageObjectOption(option)
	return &ImageOperator{
		object: &entity.ImageObject{
			ContentType: contenttype,
		},
		option: &objectOption,
	}
}

// Decode images
func (im *ImageOperator) Decode(src io.Reader) error {
	var err error
	im.object.Source, im.object.ImageName, err = image.Decode(src)
	rectang := im.object.Source.Bounds()
	im.object.OriginX = rectang.Bounds().Dx()
	im.object.OriginY = rectang.Bounds().Dy()
	log.Debug(fmt.Sprintf("OriginX: %d / OriginY: %d", im.object.OriginX, im.object.OriginY))
	return err
}

// Resize images
func (im *ImageOperator) Resize() {
	im.calcResizeXY()
	rect := image.Rect(0, 0, im.object.DstX, im.object.DstY)
	switch im.option.Quality {
	case 1:
		im.object.Dst = im.scale(im.object.Source, rect, draw.NearestNeighbor)
	case 2:
		im.object.Dst = im.scale(im.object.Source, rect, draw.ApproxBiLinear)
	case 3:
		im.object.Dst = im.scale(im.object.Source, rect, draw.BiLinear)
	case 4:
		im.object.Dst = im.scale(im.object.Source, rect, draw.CatmullRom)
	default:
		im.object.Dst = im.scale(im.object.Source, rect, draw.CatmullRom)
	}
}

// ImageByte get image bytes
func (im *ImageOperator) ImageByte() ([]byte, error) {
	buf := new(bytes.Buffer)
	switch im.object.ContentType {
	case "image/jpeg":
		if err := jpeg.Encode(buf, im.object.Dst, nil); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "image/png":
		if err := png.Encode(buf, im.object.Dst); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "image/gif":
		if err := gif.Encode(buf, im.object.Dst, nil); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		return nil, errors.New("invalid format")
	}
}

func (im *ImageOperator) scale(src image.Image, rect image.Rectangle, scaler draw.Scaler) image.Image {
	dst := image.NewRGBA(rect)
	scaler.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func (im *ImageOperator) calcResizeXY() {
	log.Debug(fmt.Sprintf("OptionX: %d / OptionY: %d", im.option.Width, im.option.Height))
	switch {
	case (im.option.Width == 0 && im.option.Height == 0):
		im.object.DstX = im.object.OriginX
		im.object.DstY = im.object.OriginY
	case (im.option.Width != 0 && im.option.Height == 0):
		im.calcResizeFitOptionWidth()
	case (im.option.Width == 0 && im.option.Height != 0):
		im.calcResizeFitOptionHeight()
	case (im.option.Width != 0 && im.option.Height != 0 && float64(im.object.OriginY)/float64(im.object.OriginX) <= float64(im.option.Height)/float64(im.option.Width)):
		im.calcResizeFitOptionWidth()
	case (im.option.Width != 0 && im.option.Height != 0 && float64(im.object.OriginY)/float64(im.object.OriginX) > float64(im.option.Height)/float64(im.option.Width)):
		im.calcResizeFitOptionHeight()
	}
	log.Debug(fmt.Sprintf("DstX: %d / DstY: %d", im.object.DstX, im.object.DstY))
}

func (im *ImageOperator) calcResizeFitOptionWidth() {
	im.object.DstX = im.option.Width
	if im.object.OriginX != 0 {
		im.object.DstY = int(float64(im.option.Width) * (float64(im.object.OriginY) / float64(im.object.OriginX)))
	} else {
		im.object.DstY = im.object.OriginY
	}
}

func (im *ImageOperator) calcResizeFitOptionHeight() {
	if im.object.OriginY != 0 {
		im.object.DstX = int(float64(im.option.Height) * (float64(im.object.OriginX) / float64(im.object.OriginY)))
	} else {
		im.object.DstX = im.object.OriginX
	}
	im.object.DstY = im.option.Height
}
