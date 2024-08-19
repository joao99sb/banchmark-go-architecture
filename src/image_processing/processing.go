package imageprocessing

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

func ReadImage(path string) image.Image {
	inputFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	_, format, err := image.DecodeConfig(inputFile)
	if err != nil {
		log.Panic(err)
	}

	// Back to the beginning of the file
	_, err = inputFile.Seek(0, 0)
	if err != nil {
		log.Panic(err)
	}

	var img image.Image
	switch strings.ToLower(format) {
	case "jpeg":
		img, err = jpeg.Decode(inputFile)
	case "png":
		img, err = png.Decode(inputFile)
	default:
		log.Panic(fmt.Errorf("unsupported image format: %s", format))
	}

	if err != nil {
		log.Panic(err)
	}

	return img
}

func Resize(img image.Image) image.Image {
	newWidth := uint(500)
	newHeight := uint(500)
	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	return resizedImg
}

func Grayscale(img image.Image) image.Image {

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalPixel := img.At(x, y)
			grayPixel := color.GrayModel.Convert(originalPixel)
			grayImg.Set(x, y, grayPixel)
		}
	}
	return grayImg
}

func WriteImage(path string, img image.Image) error {

	pathSplited := strings.Split(path, "/")
	outDir := strings.Join(pathSplited[:len(pathSplited)-1], "/")
	err := os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Encode the image to the new file
	err = jpeg.Encode(outputFile, img, nil)
	if err != nil {
		return err
	}
	return nil

}
