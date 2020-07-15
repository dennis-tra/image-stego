package chunk

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path"
)

// OpenImageFile opens the file at the given path and returns the decoded *image.RGBA
func OpenImageFile(filename string) (*image.RGBA, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return ImageToRGBA(img), nil
}

// SaveImageFile saves the given image data to the given filepath as a PNG image.
func SaveImageFile(filepath string, img image.Image) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	return nil
}

// SetExtension sets the file extension to nexExt and removes the old one
func SetExtension(filename string, newExt string) string {
	ext := path.Ext(filename)
	return filename[0:len(filename)-len(ext)] + newExt
}

// ImageToRGBA converts an image.Image to an *image.RGBA
func ImageToRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(rgba, rgba.Bounds(), src, bounds.Min, draw.Src)
	return rgba
}
