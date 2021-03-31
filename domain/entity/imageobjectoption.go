package entity

// ImageObjectOption entity
type ImageObjectOption struct {
	Width      int
	Height     int
	Quality    int
	Rotate     string
	Crop       [4]int
	Brightness int
	Contrast   int
	Gamma      float64
}
