package main

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const DOC = "" +
	"Skyboxer is a simple tool to cut up free skybox jpgs or pngs into usable textures.\n\n" +
	"Example Usage: \n" +
	"skyboxer -f ~/Downloads/skybox.jpg -o ./game/assets/skybox\n" +
	"Arguments:\n" +
	"-f | --file      <path to file to slice up>\n" +
	"-o | --out-dir   <path to desired out dir>\n" +
	"-h | --help      Displays this page"

func Has(haystack []string, needles []string) bool {
	needleLen := len(needles)
	for i := 0; i < len(haystack); i++ {
		for j := 0; j < needleLen; j++ {
			if haystack[i] == needles[j] {
				return true
			}
		}
	}
	return false
}

func GetArgValue(args []string, flags []string) (string, error) {
	flagLen := len(flags)
	for i := 0; i < len(args)-1; i++ {
		for j := 0; j < flagLen; j++ {
			if args[i] == flags[j] {
				return args[i+1], nil
			}
		}
	}
	return "", errors.New(fmt.Sprintf("Arg %s missing value", strings.Join(flags, " | ")))
}

type Bound struct {
	X    int
	Y    int
	Name string
}

func main() {
	args := os.Args[1:]

	helpFlags := []string{"--help", "-h"}
	fileFlags := []string{"--file", "-f"}
	outDirFlags := []string{"--out-dir", "-o"}

	if len(args) < 1 || Has(args, helpFlags) {
		fmt.Println(DOC)
		os.Exit(0)
		return
	}

	filePath, err := GetArgValue(args, fileFlags)
	if err != nil || filePath == "" {
		log.Fatal(err)
		os.Exit(1)
		return
	}

	outDir, err := GetArgValue(args, outDirFlags)
	if err != nil || outDir == "" {
		log.Fatal(err)
		os.Exit(1)
		return
	}

	fileAbsPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return
	}

	outDirAbsPath, err := filepath.Abs(outDir)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return
	}

	file, err := os.Open(fileAbsPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return
	}
	defer file.Close()

	img, imageType, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return
	}

	// Set Reader back to start
	file.Seek(0, 0)
	imgConfig, _, err := image.DecodeConfig(file)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return
	}

	pixelSizeX := imgConfig.Width / 4
	pixelSizeY := imgConfig.Height / 3

	bounds := []Bound{
		{X: pixelSizeX, Y: 0, Name: "posY"},
		{X: 0, Y: pixelSizeY, Name: "negX"},
		{X: pixelSizeX, Y: pixelSizeY, Name: "negZ"},
		{X: pixelSizeX, Y: pixelSizeY * 2, Name: "negY"},
		{X: pixelSizeX * 2, Y: pixelSizeY, Name: "posX"},
		{X: pixelSizeX * 3, Y: pixelSizeY, Name: "posZ"},
	}

	for _, bound := range bounds {
		outFile, err := os.Create(filepath.Join(outDirAbsPath, bound.Name) + "." + imageType)
    defer outFile.Close()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}

		newImg := image.NewRGBA(image.Rect(0, 0, pixelSizeX, pixelSizeY))
		for xi := bound.X; xi < pixelSizeX+bound.X; xi++ {
			for yi := bound.Y; yi < pixelSizeY+bound.Y; yi++ {
				newImg.Set(xi-bound.X, yi-bound.Y, img.At(xi, yi))
			}
		}

		if imageType == "jpg" {
			fmt.Println("encoding jpg", imageType)
			jpeg.Encode(outFile, newImg, nil)
		} else if imageType == "png" {
			fmt.Println("encoding png", imageType)
			png.Encode(outFile, newImg)
		}
	}
}
