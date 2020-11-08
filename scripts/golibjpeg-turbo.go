package main

import (
	"bufio"
	"github.com/pixiv/go-libjpeg/jpeg"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]

	// Decoding JPEG into image.Image
	io, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	img, err := jpeg.Decode(io, &jpeg.DecoderOptions{})
	if err != nil {
		log.Fatalf("Decode returns error: %v\n", err)
	}

	// Encode JPEG
	f, err := os.Create(args[1])
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(f)
	if err := jpeg.Encode(w, img, &jpeg.EncoderOptions{Quality: 100}); err != nil {
		log.Printf("Encode returns error: %v\n", err)
		return
	}
	w.Flush()
	f.Close()
}
