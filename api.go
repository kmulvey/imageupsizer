package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/apex/log"
)

func FindLargerImageFromFile(filename string) (*imageData, error) {
	originalImage, err := getImageConfigFromFile(filename)
	if err != nil {
		return nil, err
	}

	log.Infof("[%s] Uploading to google image server...", filename)
	contents, err := uploadImage(filename)
	if err != nil {
		return nil, err
	}

	largerImage, err := getImageList(contents)
	if err != nil {
		return nil, err
	}
	if largerImage.Area > originalImage.Width*originalImage.Height {
		return largerImage, nil
	}
	return nil, errNoLargerAvailable
}

func GetLargerImageFromFile(filename string) (*imageData, error) {
	var largerImage, err = FindLargerImageFromFile(filename)
	if err != nil {
		return nil, err
	}

	imageInfo, err := getImage(largerImage.URL)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(path.Base(imageInfo.URL), imageInfo.Bytes, os.ModePerm); err != nil {
		log.Error(err)
	}

	return imageInfo, ioutil.WriteFile(path.Base(imageInfo.URL), imageInfo.Bytes, 0755)
}

func FindLargerImageFromBytes(image []byte, outputFile string) (*imageData, error) {
	var tmpfile = "FindLargerImageFromBytesTmpfile.image"
	defer os.Remove(tmpfile)

	var err = ioutil.WriteFile(tmpfile, image, 0755)
	if err != nil {
		return nil, err
	}

	largerImage, err := FindLargerImageFromFile(tmpfile)
	if err != nil {
		return nil, err
	}

	return largerImage, os.Remove(tmpfile)
}

func GetLargerImageFromBytes(image []byte) (*imageData, error) {
	var tmpfile = "GetLargerImageFromBytesTmpfile.image"
	defer os.Remove(tmpfile)

	var err = ioutil.WriteFile(tmpfile, image, 0755)
	if err != nil {
		return nil, err
	}
	return GetLargerImageFromFile(tmpfile)
}
