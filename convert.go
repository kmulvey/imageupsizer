package imageupsizer

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
)

// Convert converts pngs and webps to jpeg
// this first string returned is the name of the new file
// the second string returned is the type of the input image (png, webp), as detected from its encoding, not file name
func Convert(from string) (string, string, error) {
	var origFile, err = os.Open(from)
	if err != nil {
		return "", "", fmt.Errorf("cannot open image: name: %s, error: %w", from, err)
	}
	defer origFile.Close()

	var ext = filepath.Ext(from)
	var newFile string
	if ext == "" {
		newFile = from + ".jpg"
	} else {
		newFile = strings.Replace(from, ext, ".jpg", 1)
	}

	imgData, imageType, err := image.Decode(origFile)
	if err != nil {
		return "", "", fmt.Errorf("img decode: name: %s, error: %w", from, err)
	}

	// dont bother converting jpegs
	if imageType == "jpeg" {
		return from, "jpeg", nil
	}

	err = os.Remove(from)
	if err != nil {
		return "", "", fmt.Errorf("remove input file: name: %s, error: %w", from, err)
	}

	out, err := os.Create(newFile)
	if err != nil {
		return "", "", fmt.Errorf("new jpg create: name: %s, error: %w", from, err)
	}

	err = jpeg.Encode(out, imgData, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", "", fmt.Errorf("jpg encode: name: %s, error: %w", from, err)
	}

	err = out.Close()
	if err != nil {
		return "", "", fmt.Errorf("new jpg close: name: %s, error: %w", from, err)
	}

	return newFile, imageType, nil
}
