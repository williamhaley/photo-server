package main

import (
	"github.com/koofr/goepeg"
	"github.com/koofr/gothumb"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	input := args[0]
	output := args[1]

	size, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal(err)
	}

	quality := 100

	in, err := os.Open(input)

	if err != nil {
		panic(err)
	}

	out, err := gothumb.Thumbnail(in, size, quality, goepeg.ScaleTypeFitMax)

	if err != nil {
		panic(err)
	}

	defer out.Close()

	outputFile, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0664)

	if err != nil {
		panic(err)
	}

	_, err = io.Copy(outputFile, out)

	if err != nil {
		panic(err)
	}
}
