package main

import (
	"github.com/disintegration/imageorient"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"log"
	"math"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	inputFile, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	// decodedImage, err := jpeg.Decode(inputFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	decodedImage, _, err := imageorient.Decode(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	dimension, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal(err)
	}

	bounds := decodedImage.Bounds()

	width := dimension
	height := dimension

	maxDimension := math.Max(float64(bounds.Max.X), float64(bounds.Max.Y))
	minDimension := math.Min(float64(bounds.Max.X), float64(bounds.Max.Y))
	ratio := minDimension / maxDimension

	if bounds.Max.Y > bounds.Max.X {
		width = int(float64(width) * ratio)
	} else {
		height = int(float64(height) * ratio)
	}

	newSize := image.Rect(0, 0, width, height)
	scaledImage := scaleTo(decodedImage, newSize, draw.NearestNeighbor)

	outputFile, err := os.Create(args[1])
	if err != nil {
		log.Fatal(err)
	}

	jpeg.Encode(outputFile, scaledImage, nil)
}

func scaleTo(src image.Image,
	rect image.Rectangle, scale draw.Scaler) image.Image {
	dst := image.NewRGBA(rect)
	scale.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}
