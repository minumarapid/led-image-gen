package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"os"

	"led-image-gen/processor"
)

func main() {
	inPath := flag.String("in", "", "Path to the input PNG image")
	outPath := flag.String("out", "", "Path to save the output PNG image")
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		log.Fatal("Both -in and -out flags are required")
	}

	// 入力画像を開く
	inFile, err := os.Open(*inPath)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer inFile.Close()

	img, _, err := image.Decode(inFile)
	if err != nil {
		log.Fatalf("Failed to decode input file: %v", err)
	}

	config := &processor.Config{
		Border:        10,
		LEDSize:       4,
		LEDGap:        2,
		LEDGamma:      1.0,
		LEDExposure:   1.0,
		LEDShape:      false,
		MaxWorkers:    4,
		EnableGlow:    true,
		GlowStrength:  1.0,
		GlowGamma:     1.0,
		GlowExposure:  1.0,
		OffLightColor: color.RGBA{40, 40, 40, 255}, // 透明黒
	}

	resultImg, err := processor.GenerateLEDImage(img, config)
	if err != nil {
		log.Fatalf("Failed to generate LED image: %v", err)
	}

	outFile, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, resultImg)
	if err != nil {
		log.Fatalf("Failed to encode output file: %v", err)
	}

	log.Printf("Successfully processed image and saved to %s", *outPath)
}
