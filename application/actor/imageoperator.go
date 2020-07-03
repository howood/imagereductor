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

	log "github.com/howood/imagereductor/infrastructure/logger"
	"golang.org/x/image/draw"
)

type ImageOperator struct {
	source      image.Image
	dst         image.Image
	originX     int
	originY     int
	dstX        int
	dstY        int
	imagename   string
	ContentType string
	Option      ImageOperatorOption
}

type ImageOperatorOption struct {
	Width   int
	Height  int
	Quality int
}

func NewImageOperator(contenttype string, option ImageOperatorOption) *ImageOperator {
	return &ImageOperator{
		ContentType: contenttype,
		Option:      option,
	}
}

func (im *ImageOperator) Decode(src io.Reader) error {
	var err error
	im.source, im.imagename, err = image.Decode(src)
	rectang := im.source.Bounds()
	im.originX = rectang.Bounds().Dx()
	im.originY = rectang.Bounds().Dy()
	log.Debug(fmt.Sprintf("OriginX: %d / OriginY: %d", im.originX, im.originY))
	return err
}

func (im *ImageOperator) Resize() {
	im.calcResizeXY()
	rect := image.Rect(0, 0, im.dstX, im.dstY)
	switch im.Option.Quality {
	case 1:
		im.dst = im.scale(im.source, rect, draw.NearestNeighbor)
	case 2:
		im.dst = im.scale(im.source, rect, draw.ApproxBiLinear)
	case 3:
		im.dst = im.scale(im.source, rect, draw.BiLinear)
	case 4:
		im.dst = im.scale(im.source, rect, draw.CatmullRom)
	default:
		im.dst = im.scale(im.source, rect, draw.CatmullRom)
	}
}

func (im *ImageOperator) ImageByte() ([]byte, error) {
	buf := new(bytes.Buffer)
	switch im.ContentType {
	case "image/jpeg":
		if err := jpeg.Encode(buf, im.dst, nil); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "image/png":
		if err := png.Encode(buf, im.dst); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "image/gif":
		if err := gif.Encode(buf, im.dst, nil); err != nil {
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
	log.Debug(fmt.Sprintf("OptionX: %d / OptionY: %d", im.Option.Width, im.Option.Height))
	switch {
	case (im.Option.Width == 0 && im.Option.Height == 0):
		im.dstX = im.originX
		im.dstY = im.originY
	case (im.Option.Width != 0 && im.Option.Height == 0):
		im.calcResizeFitOptionWidth()
	case (im.Option.Width == 0 && im.Option.Height != 0):
		im.calcResizeFitOptionHeight()
	case (im.Option.Width != 0 && im.Option.Height != 0 && float64(im.originY)/float64(im.originX) <= float64(im.Option.Height)/float64(im.Option.Width)):
		im.calcResizeFitOptionWidth()
	case (im.Option.Width != 0 && im.Option.Height != 0 && float64(im.originY)/float64(im.originX) > float64(im.Option.Height)/float64(im.Option.Width)):
		im.calcResizeFitOptionHeight()
	}
	log.Debug(fmt.Sprintf("DstX: %d / DstY: %d", im.dstX, im.dstY))
}

func (im *ImageOperator) calcResizeFitOptionWidth() {
	log.Debug("######1########")
	im.dstX = im.Option.Width
	if im.originX != 0 {
		im.dstY = int(float64(im.Option.Width) * (float64(im.originY) / float64(im.originX)))
	} else {
		im.dstY = im.originY
	}
}

func (im *ImageOperator) calcResizeFitOptionHeight() {
	log.Debug("######2########")
	if im.originY != 0 {
		im.dstX = int(float64(im.Option.Height) * (float64(im.originX) / float64(im.originY)))
	} else {
		im.dstX = im.originX
	}
	im.dstY = im.Option.Height
}
