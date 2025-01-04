package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sentriz/audiotags"
)

var usage = func() {
	_, _ = fmt.Fprintf(os.Stderr, "usage: %s [optional flags] filename\n", os.Args[0])
	flag.PrintDefaults()
}

func _main(filename string) error {
	f, err := audiotags.Open(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	if !f.HasMedia() {
		return fmt.Errorf("no supported media in file")
	}

	fmt.Printf("\nTags: \n")
	for tag, value := range f.ReadTags() {
		fmt.Printf("%s: %v\n", tag, value)
	}

	cover, err := f.ReadImage()
	if err != nil {
		return fmt.Errorf("error reading cover: %v", err)
	}

	if cover != nil {
		fmt.Printf("Cover: %dx%d\n", cover.Bounds().Size().X, cover.Bounds().Size().Y)
	}

	fmt.Printf("\nProps: \n")
	props := f.ReadAudioProperties()
	fmt.Printf("Bitrate: %v\n", props.Bitrate)
	fmt.Printf("Length: %v\n", props.Length)
	fmt.Printf("Samplerate: %d\n", props.Samplerate)
	fmt.Printf("Channels: %d\n", props.Channels)

	// coverFile, err := os.Create("cover.png")
	// if err != nil {
	// 	return fmt.Errorf("error opening output cover: %v", err)
	// }

	// if err = png.Encode(coverFile, cover); err != nil {
	// 	return fmt.Errorf("error writing output cover: %v", err)
	// }
	// coverFile.Close()

	// file, err := os.Open("YOUR TEST COVER")
	// if err != nil {
	// 	return fmt.Errorf("error opening input cover: %v", err)
	// }
	// defer file.Close()

	// img, _, err := image.Decode(file)
	// if err != nil {
	// 	return fmt.Errorf("error reading input cover: %v", err)
	// }

	// // JPEG = 0
	// if err = f.WriteImage(img, 0); err != nil {
	// 	return fmt.Errorf("writing image cover: %v", err)
	// }

	return nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		return
	}

	if err := _main(flag.Arg(0)); err != nil {
		log.Fatalln("Error: " + err.Error())
	}
}
