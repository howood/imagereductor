package entity

import "image"

type ImageObject struct {
	Source      image.Image
	Dst         image.Image
	OriginX     int
	OriginY     int
	DstX        int
	DstY        int
	ImageName   string
	ContentType string
}
