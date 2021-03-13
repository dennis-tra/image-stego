package main

import (
	"flag"
	"log"
	"os"
	"path"

	"dennis-tra/image-stego/internal/chunk"
)

func main() {

	decodePtr := flag.Bool("d", false, "Whether to decode the given image file(s)")
	encodePtr := flag.Bool("e", false, "Whether to encode the given image file(s)")
	outputPtr := flag.String("o", "", "Output directory of an encoded image")

	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(path.Join(cwd, *outputPtr)); *encodePtr && os.IsNotExist(err) {
		log.Println("Output directory does not exist")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *decodePtr == *encodePtr {
		log.Println("Incompatible combination of decode and encode flags")
		log.Println("Please specify weather you want to encode -e or decode -d the image file(s)")
		flag.PrintDefaults()
		os.Exit(1)
	}

	for _, filename := range flag.Args() {

		if *decodePtr {
			err = chunk.Decode(filename)
		} else if *encodePtr {
			err = chunk.Encode(filename, *outputPtr)
		}
		if err != nil {
			log.Println(err)
		}
	}
}
