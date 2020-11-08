package main

import (
	"github.com/koofr/goepeg"
	"log"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	dimension, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal(err)
	}

	quality := 100
	if err := goepeg.Thumbnail(args[0], args[1], dimension, quality, goepeg.ScaleTypeFitMax); err != nil {
		log.Fatal(err)
	}
}
