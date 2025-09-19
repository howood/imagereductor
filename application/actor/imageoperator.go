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
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

const (
	// ImageRotateRight is rotate image 90 degree right.
	ImageRotateRight = "right"
	// ImageRotateLeft is rotate image 90 degree left.
	ImageRotateLeft = "left"
	// ImageRotateUpsidedown is rotate image upside down.
	ImageRotateUpsidedown = "upsidedown"
	// ImageRotateAutoVertical is rotate image auto vertical.
	ImageRotateAutoVertical = "autovertical"
	// ImageRotateAutoHorizontal is rotate image auto horizontal.
	ImageRotateAutoHorizontal = "autohorizontal"
	// ImageRotateExifOrientation is rotate image by exif orientation.
	ImageRotateExifOrientation = "exiforientation"
)

// ImageOperator struct.
type ImageOperator struct {
	repository.ImageObjectRepository
}

// NewImageOperator creates a new ImageObjectRepository.
func NewImageOperator(contenttype string, option ImageOperatorOption) *ImageOperator {
	objectOption := entity.ImageObjectOption(option)
	return &ImageOperator{
		&imageCreator{
			object: &entity.ImageObject{
				ContentType: contenttype,
			},
			option: &objectOption,
		},
	}
}

// imageCreator struct.
type imageCreator struct {
	object          *entity.ImageObject
	option          *entity.ImageObjectOption
	exifOrientation int
}

// ImageOperatorOption is Option of ImageOperator struct.
type ImageOperatorOption entity.ImageObjectOption

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

// Decode images.
func (im *imageCreator) Decode(ctx context.Context, src io.ReadSeeker) error {
	var err error
	im.object.Source, im.object.ImageName, err = image.Decode(src)
	if err == nil {
		im.decodeExifOrientation(ctx, src)
	}
	if err == nil {
		rectang := im.object.Source.Bounds()
		im.object.OriginX = rectang.Bounds().Dx()
		im.object.OriginY = rectang.Bounds().Dy()
		log.Debug(ctx, fmt.Sprintf("OriginX: %d / OriginY: %d", im.object.OriginX, im.object.OriginY))
	}
	return err
}

// Process images process resize and more.
func (im *imageCreator) Process(ctx context.Context) error {
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
		err := im.rotate(ctx)
		if err != nil {
			return err
		}
		im.calcResizeXY(ctx)
		return im.resize()
	case !reflect.DeepEqual(im.option.Crop, [4]int{}):
		im.calcResizeXYWithCrop(ctx)
		return im.cropAndResize()
	default:
		im.calcResizeXY(ctx)
		return im.resize()
	}
}

// ImageByte get image bytes.
func (im *imageCreator) ImageByte(_ context.Context) ([]byte, error) {
	buf := new(bytes.Buffer)
	var err error
	switch im.object.ContentType {
	case "image/jpeg":
		err = jpeg.Encode(buf, im.object.Dst, im.jpegOption())
	case "image/png":
		err = png.Encode(buf, im.object.Dst)
	case "image/gif":
		err = gif.Encode(buf, im.object.Dst, nil)
	default:
		//nolint:err113
		err = errors.New("invalid format")
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// resize images.
func (im *imageCreator) resize() error {
	rect := image.Rect(0, 0, im.object.DstX, im.object.DstY)
	im.object.Dst = im.scale(im.object.Source, rect, im.getDrawer())
	return nil
}

// crop and resize images.
func (im *imageCreator) cropAndResize() error {
	croprect := image.Rect(im.option.Crop[0], im.option.Crop[1], im.option.Crop[2], im.option.Crop[3])
	cropimg, err := im.subimage(im.object.Source, croprect)
	if err != nil {
		return err
	}
	scalerect := image.Rect(0, 0, im.object.DstX, im.object.DstY)
	im.object.Dst = im.scale(cropimg, scalerect, im.getDrawer())
	return nil
}

// rotate images.
//
//nolint:mnd,cyclop
func (im *imageCreator) rotate(ctx context.Context) error {
	originX := im.object.OriginX
	originY := im.object.OriginY
	switch im.option.Rotate {
	case ImageRotateRight:
		rect := image.Rect(0, 0, im.object.OriginY, im.object.OriginX)
		im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 90.0, float64(im.object.OriginY), 0), im.getDrawer())
		im.object.OriginX = originY
		im.object.OriginY = originX
	case ImageRotateLeft:
		rect := image.Rect(0, 0, im.object.OriginY, im.object.OriginX)
		im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 270.0, 0, float64(im.object.OriginX)), im.getDrawer())
		im.object.OriginX = originY
		im.object.OriginY = originX
	case ImageRotateUpsidedown:
		rect := image.Rect(0, 0, im.object.OriginX, im.object.OriginY)
		im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 180.0, float64(im.object.OriginX), float64(im.object.OriginY)), im.getDrawer())
	case ImageRotateAutoVertical:
		if im.object.OriginX > im.object.OriginY {
			rect := image.Rect(0, 0, im.object.OriginY, im.object.OriginX)
			im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 90.0, float64(im.object.OriginY), 0), im.getDrawer())
			im.object.OriginX = originY
			im.object.OriginY = originX
		}
	case ImageRotateAutoHorizontal:
		if im.object.OriginY > im.object.OriginX {
			rect := image.Rect(0, 0, im.object.OriginY, im.object.OriginX)
			im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 270.0, 0, float64(im.object.OriginX)), im.getDrawer())
			im.object.OriginX = originY
			im.object.OriginY = originX
		}
	case ImageRotateExifOrientation:
		if im.exifOrientation == 3 {
			rect := image.Rect(0, 0, im.object.OriginY, im.object.OriginX)
			im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 180.0, 0, float64(im.object.OriginX)), im.getDrawer())
		}
		if im.exifOrientation == 6 {
			rect := image.Rect(0, 0, im.object.OriginY, im.object.OriginX)
			im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 90.0, 0, float64(im.object.OriginX)), im.getDrawer())
			im.object.OriginX = originY
			im.object.OriginY = originX
		}
		if im.exifOrientation == 8 {
			rect := image.Rect(0, 0, im.object.OriginY, im.object.OriginX)
			im.object.Source = im.transform(im.object.Source, rect, im.calcRotateAffine(ctx, 270.0, 0, float64(im.object.OriginX)), im.getDrawer())
			im.object.OriginX = originY
			im.object.OriginY = originX
		}
	default:
		//nolint:err113
		return errors.New("invalid Rotate Parameter")
	}
	return nil
}

// scale image.
func (im *imageCreator) scale(src image.Image, rect image.Rectangle, scaler draw.Scaler) image.Image {
	dst := image.NewNRGBA(rect)
	scaler.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

// crop image.
func (im *imageCreator) subimage(src image.Image, rect image.Rectangle) (image.Image, error) {
	simg, ok := src.(subImager)
	if !ok {
		//nolint:err113
		return nil, errors.New("image not support Crop")
	}
	return simg.SubImage(rect), nil
}

// rotate image.
func (im *imageCreator) transform(src image.Image, rect image.Rectangle, t f64.Aff3, scaler draw.Transformer) image.Image {
	dst := image.NewNRGBA(rect)
	scaler.Transform(dst, t, src, src.Bounds(), draw.Over, nil)
	return dst
}

// change brightness.
//
//nolint:mnd
func (im *imageCreator) brightness(src image.Image) *image.NRGBA {
	lookup := make([]uint8, 256)
	percentage := math.Min(math.Max(float64(im.option.Brightness), -100.0), 100.0)
	for i := range 256 {
		lookup[i] = uint8(utils.InRanged(float64(i)*(percentage/100.0), 0, 255))
	}
	return im.convertLuminance(src, lookup)
}

// change contrast.
//
//nolint:mnd
func (im *imageCreator) contrast(src image.Image) *image.NRGBA {
	lookup := make([]uint8, 256)
	percentage := math.Min(math.Max(float64(im.option.Contrast), -100.0), 100.0)
	v := (100.0 + percentage) / 100.0
	for i := range 256 {
		lookup[i] = uint8(utils.InRanged(((((float64(i)/255)-0.5)*v)+0.5)*255, 0, 255))
	}
	return im.convertLuminance(src, lookup)
}

// change gamma.
//
//nolint:mnd
func (im *imageCreator) gamma(src image.Image) *image.NRGBA {
	lookup := make([]uint8, 256)
	e := 1.0 / math.Max(im.option.Gamma, 0.0001)
	for i := range 256 {
		lookup[i] = uint8(utils.InRanged(math.Pow(float64(i)/255.0, e)*255.0, 0, 255))
	}
	return im.convertLuminance(src, lookup)
}

//nolint:mnd
func (im *imageCreator) convertLuminance(src image.Image, lookup []uint8) *image.NRGBA {
	fnc := func(c color.RGBA) color.RGBA {
		return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
	}
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	utils.ApplyParallel(0, dst.Bounds().Dy(), func(start, end int) {
		for y := start; y < end; y++ {
			for x := range dst.Bounds().Dx() {
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

//nolint:mnd,ireturn
func (im *imageCreator) getDrawer() draw.Interpolator {
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

//nolint:cyclop
func (im *imageCreator) calcResizeXY(ctx context.Context) {
	log.Debug(ctx, fmt.Sprintf("OptionX: %d / OptionY: %d", im.option.Width, im.option.Height))
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
	log.Debug(ctx, fmt.Sprintf("DstX: %d / DstY: %d", im.object.DstX, im.object.DstY))
}

//nolint:cyclop
func (im *imageCreator) calcResizeXYWithCrop(ctx context.Context) {
	log.Debug(ctx, fmt.Sprintf("OptionX: %d / OptionY: %d", im.option.Width, im.option.Height))
	log.Debug(ctx, fmt.Sprintf("Crop: %v", im.option.Crop))
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
	log.Debug(ctx, fmt.Sprintf("DstX: %d / DstY: %d", im.object.DstX, im.object.DstY))
}

func (im *imageCreator) calcResizeFitOptionWidth(originx, originy int) {
	im.object.DstX = im.option.Width
	im.object.DstY = originy
	if originx != 0 {
		im.object.DstY = int(float64(im.option.Width) * (float64(originy) / float64(originx)))
	}
}

func (im *imageCreator) calcResizeFitOptionHeight(originx, originy int) {
	im.object.DstX = originx
	if originy != 0 {
		im.object.DstX = int(float64(im.option.Height) * (float64(originx) / float64(originy)))
	}
	im.object.DstY = im.option.Height
}

//nolint:mnd
func (im *imageCreator) calcRotateAffine(ctx context.Context, deg, moveleft, movedown float64) f64.Aff3 {
	log.Debug(ctx, fmt.Sprintf("deg: %v, moveleft: %v, movedown: %v", deg, moveleft, movedown))
	rad := deg * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	return f64.Aff3{
		+cos, -sin, moveleft,
		+sin, +cos, movedown,
	}
}

//nolint:mnd
func (im *imageCreator) jpegOption() *jpeg.Options {
	switch im.option.Quality {
	case 1:
		return &jpeg.Options{Quality: 75}
	case 2:
		return &jpeg.Options{Quality: 85}
	case 3:
		return &jpeg.Options{Quality: 90}
	case 4:
		return &jpeg.Options{Quality: 100}
	default:
		return &jpeg.Options{Quality: 85}
	}
}

func (im *imageCreator) decodeExifOrientation(ctx context.Context, src io.ReadSeeker) {
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		log.Debug(ctx, fmt.Sprintf("reader seek 0 error %v", err.Error()))
		return
	}
	decodedExif, err := exif.Decode(src)
	if err != nil {
		log.Debug(ctx, fmt.Sprintf("exif decode error %v", err.Error()))
		return
	}
	orientation, err := decodedExif.Get(exif.Orientation)
	if err != nil {
		log.Debug(ctx, fmt.Sprintf("exif orientation error %v", err.Error()))
		return
	}
	orientationvVal, err := orientation.Int(0)
	if err != nil {
		log.Debug(ctx, fmt.Sprintf("exif orientation int error %v", err.Error()))
		return
	}
	im.exifOrientation = orientationvVal
	log.Debug(ctx, fmt.Sprintf("exif orientation %v", im.exifOrientation))
}
