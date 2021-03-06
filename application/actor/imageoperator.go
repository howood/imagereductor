package actor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"reflect"

	"github.com/howood/imagereductor/domain/entity"
	"github.com/howood/imagereductor/domain/repository"
	log "github.com/howood/imagereductor/infrastructure/logger"
	"github.com/howood/imagereductor/library/utils"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

const (
	// ImageRotateRight is rotate image 90 degree right
	ImageRotateRight = "right"
	// ImageRotateLeft is rotate image 90 degree left
	ImageRotateLeft = "left"
	// ImageRotateUpsidedown is rotate image upside down
	ImageRotateUpsidedown = "upsidedown"
)

// ImageOperator struct
type ImageOperator struct {
	object *entity.ImageObject
	option *entity.ImageObjectOption
	ctx    context.Context
}

// ImageOperatorOption is Option of ImageOperator struct
type ImageOperatorOption entity.ImageObjectOption

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

// NewImageOperator creates a new ImageObjectRepository
func NewImageOperator(ctx context.Context, contenttype string, option ImageOperatorOption) repository.ImageObjectRepository {
	objectOption := entity.ImageObjectOption(option)
	return &ImageOperator{
		object: &entity.ImageObject{
			ContentType: contenttype,
		},
		option: &objectOption,
		ctx:    ctx,
	}
}

// Decode images
func (im *ImageOperator) Decode(src io.Reader) error {
	var err error
	im.object.Source, im.object.ImageName, err = image.Decode(src)
	if err == nil {
		rectang := im.object.Source.Bounds()
		im.object.OriginX = rectang.Bounds().Dx()
		im.object.OriginY = rectang.Bounds().Dy()
		log.Debug(im.ctx, fmt.Sprintf("OriginX: %d / OriginY: %d", im.object.OriginX, im.object.OriginY))
	}
	return err
}

// Process images process resize and more
func (im *ImageOperator) Process() error {
	if im.option.Gamma != 0 {
		im.object.Source = im.gamma(im.object.Source)
	}
	if im.option.Contrast != 0 {
		im.object.Source = im.contrast(im.object.Source)
	}
	if im.option.Brightness != 0 {
		im.object.Source = im.brightness(im.object.Source)
	}
	switch {
	case (im.option.Rotate != ""):
		im.calcResizeXY()
		return im.rotateAndResize()
	case (reflect.DeepEqual(im.option.Crop, [4]int{}) == false):
		im.calcResizeXYWithCrop()
		return im.cropAndResize()
	default:
		im.calcResizeXY()
		return im.resize()
	}
}

// resize images
func (im *ImageOperator) resize() error {
	rect := image.Rect(0, 0, im.object.DstX, im.object.DstY)
	im.object.Dst = im.scale(im.object.Source, rect, im.getDrawer())
	return nil
}

// crop and resize images
func (im *ImageOperator) cropAndResize() error {
	croprect := image.Rect(im.option.Crop[0], im.option.Crop[1], im.option.Crop[2], im.option.Crop[3])
	cropimg, err := im.subimage(im.object.Source, croprect)
	if err != nil {
		return err
	}
	scalerect := image.Rect(0, 0, im.object.DstX, im.object.DstY)
	im.object.Dst = im.scale(cropimg, scalerect, im.getDrawer())
	return nil
}

// rotateAndResize images
func (im *ImageOperator) rotateAndResize() error {
	switch im.option.Rotate {
	case ImageRotateRight:
		rect := image.Rect(0, 0, im.object.DstY, im.object.DstX)
		im.object.Dst = im.transform(im.object.Source, rect, im.calcRotateAffine(90.0, float64(im.object.DstY), 0), im.getDrawer())
	case ImageRotateLeft:
		rect := image.Rect(0, 0, im.object.DstY, im.object.DstX)
		im.object.Dst = im.transform(im.object.Source, rect, im.calcRotateAffine(270.0, 0, float64(im.object.DstX)), im.getDrawer())
	case ImageRotateUpsidedown:
		rect := image.Rect(0, 0, im.object.DstX, im.object.DstY)
		im.object.Dst = im.transform(im.object.Source, rect, im.calcRotateAffine(180.0, float64(im.object.DstX), float64(im.object.DstY)), im.getDrawer())
	default:
		return fmt.Errorf("Invalid Rotate Parameter")
	}
	return nil
}

// ImageByte get image bytes
func (im *ImageOperator) ImageByte() ([]byte, error) {
	buf := new(bytes.Buffer)
	var err error
	switch im.object.ContentType {
	case "image/jpeg":
		err = jpeg.Encode(buf, im.object.Dst, nil)
	case "image/png":
		err = png.Encode(buf, im.object.Dst)
	case "image/gif":
		err = gif.Encode(buf, im.object.Dst, nil)
	default:
		err = errors.New("invalid format")
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// scale image
func (im *ImageOperator) scale(src image.Image, rect image.Rectangle, scaler draw.Scaler) image.Image {
	dst := image.NewNRGBA(rect)
	scaler.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

// crop image
func (im *ImageOperator) subimage(src image.Image, rect image.Rectangle) (image.Image, error) {
	simg, ok := src.(subImager)
	if !ok {
		return nil, fmt.Errorf("Image not support Crop")
	}
	return simg.SubImage(rect), nil
}

// rotate image
func (im *ImageOperator) transform(src image.Image, rect image.Rectangle, t f64.Aff3, scaler draw.Transformer) image.Image {
	dst := image.NewNRGBA(rect)
	scaler.Transform(dst, t, src, src.Bounds(), draw.Over, nil)
	return dst
}

// change brightness
func (im *ImageOperator) brightness(src image.Image) *image.NRGBA {
	lookup := make([]uint8, 256)
	percentage := math.Min(math.Max(float64(im.option.Brightness), -100.0), 100.0)
	for i := 0; i < 256; i++ {
		lookup[i] = uint8(utils.InRanged(float64(i)*(percentage/100.0), 0, 255))
	}
	return im.convertLuminance(src, lookup)
}

// change contrast
func (im *ImageOperator) contrast(src image.Image) *image.NRGBA {
	lookup := make([]uint8, 256)
	percentage := math.Min(math.Max(float64(im.option.Contrast), -100.0), 100.0)
	v := (100.0 + percentage) / 100.0
	for i := 0; i < 256; i++ {
		lookup[i] = uint8(utils.InRanged(((((float64(i)/255)-0.5)*v)+0.5)*255, 0, 255))
	}
	return im.convertLuminance(src, lookup)
}

// change gamma
func (im *ImageOperator) gamma(src image.Image) *image.NRGBA {
	lookup := make([]uint8, 256)
	e := 1.0 / math.Max(im.option.Gamma, 0.0001)
	for i := 0; i < 256; i++ {
		lookup[i] = uint8(utils.InRanged(math.Pow(float64(i)/255.0, e)*255.0, 0, 255))
	}
	return im.convertLuminance(src, lookup)
}

func (im *ImageOperator) convertLuminance(src image.Image, lookup []uint8) *image.NRGBA {
	fnc := func(c color.RGBA) color.RGBA {
		return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
	}
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	utils.ApplyParallel(0, dst.Bounds().Dy(), func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < dst.Bounds().Dx(); x++ {
				dstPos := y*dst.Stride + x*4
				dr := &dst.Pix[dstPos+0]
				dg := &dst.Pix[dstPos+1]
				db := &dst.Pix[dstPos+2]
				da := &dst.Pix[dstPos+3]
				c := color.RGBA{
					R: *dr,
					G: *dg,
					B: *db,
					A: *da,
				}
				c = fnc(c)
				*dr = c.R
				*dg = c.G
				*db = c.B
				*da = c.A
			}
		}
	})
	return dst
}

func (im *ImageOperator) getDrawer() draw.Interpolator {
	switch im.option.Quality {
	case 1:
		return draw.NearestNeighbor
	case 2:
		return draw.ApproxBiLinear
	case 3:
		return draw.BiLinear
	case 4:
		return draw.CatmullRom
	default:
		return draw.CatmullRom
	}
}

func (im *ImageOperator) calcResizeXY() {
	log.Debug(im.ctx, fmt.Sprintf("OptionX: %d / OptionY: %d", im.option.Width, im.option.Height))
	switch {
	case (im.option.Width == 0 && im.option.Height == 0):
		im.object.DstX = im.object.OriginX
		im.object.DstY = im.object.OriginY
	case (im.option.Width != 0 && im.option.Height == 0),
		(im.option.Width != 0 && im.option.Height != 0 && float64(im.object.OriginY)/float64(im.object.OriginX) <= float64(im.option.Height)/float64(im.option.Width)):
		im.calcResizeFitOptionWidth(im.object.OriginX, im.object.OriginY)
	case (im.option.Width == 0 && im.option.Height != 0),
		(im.option.Width != 0 && im.option.Height != 0 && float64(im.object.OriginY)/float64(im.object.OriginX) > float64(im.option.Height)/float64(im.option.Width)):
		im.calcResizeFitOptionHeight(im.object.OriginX, im.object.OriginY)
	}
	log.Debug(im.ctx, fmt.Sprintf("DstX: %d / DstY: %d", im.object.DstX, im.object.DstY))
}

func (im *ImageOperator) calcResizeXYWithCrop() {
	log.Debug(im.ctx, fmt.Sprintf("OptionX: %d / OptionY: %d", im.option.Width, im.option.Height))
	log.Debug(im.ctx, fmt.Sprintf("Crop: %v", im.option.Crop))
	cropedX := int(math.Abs(float64(im.option.Crop[2] - im.option.Crop[0])))
	cropedY := int(math.Abs(float64(im.option.Crop[3] - im.option.Crop[1])))
	switch {
	case (im.option.Width == 0 && im.option.Height == 0):
		im.object.DstX = cropedX
		im.object.DstY = cropedY
	case (im.option.Width != 0 && im.option.Height == 0),
		(im.option.Width != 0 && im.option.Height != 0 && float64(cropedY)/float64(cropedX) <= float64(im.option.Height)/float64(im.option.Width)):
		im.calcResizeFitOptionWidth(cropedX, cropedY)
	case (im.option.Width == 0 && im.option.Height != 0),
		(im.option.Width != 0 && im.option.Height != 0 && float64(cropedY)/float64(cropedX) > float64(im.option.Height)/float64(im.option.Width)):
		im.calcResizeFitOptionHeight(cropedX, cropedY)
	}
	log.Debug(im.ctx, fmt.Sprintf("DstX: %d / DstY: %d", im.object.DstX, im.object.DstY))
}

func (im *ImageOperator) calcResizeFitOptionWidth(originx, originy int) {
	im.object.DstX = im.option.Width
	im.object.DstY = originy
	if originx != 0 {
		im.object.DstY = int(float64(im.option.Width) * (float64(originy) / float64(originx)))
	}
}

func (im *ImageOperator) calcResizeFitOptionHeight(originx, originy int) {
	im.object.DstX = originx
	if originy != 0 {
		im.object.DstX = int(float64(im.option.Height) * (float64(originx) / float64(originy)))
	}
	im.object.DstY = im.option.Height
}

func (im *ImageOperator) calcRotateAffine(deg, moveleft, movedown float64) f64.Aff3 {
	log.Debug(im.ctx, fmt.Sprintf("deg: %v, moveleft: %v, movedown: %v", deg, moveleft, movedown))
	rad := deg * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	return f64.Aff3{
		+cos, -sin, moveleft,
		+sin, +cos, movedown,
	}
}
