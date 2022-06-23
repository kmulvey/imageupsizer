package imageupsizer

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func FindLargerImageFromFile(filename string) (*ImageData, error) {
	originalImage, err := GetImageConfigFromFile(filename)
	if err != nil {
		return nil, err
	}

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
	return nil, ErrNoLargerAvailable
}

func GetLargerImageFromFile(filename, outputDir string) (*ImageData, error) {
	var largerImage, err = FindLargerImageFromFile(filename)
	if err != nil {
		return nil, err
	}

	imageInfo, err := getImage(largerImage.URL)
	if err != nil {
		return nil, err
	}

	var newFile = filepath.Join(outputDir, path.Base(imageInfo.URL))
	if err := ioutil.WriteFile(newFile, imageInfo.Bytes, os.ModePerm); err != nil {
		log.Error(err)
	}
	imageInfo.LocalPath = newFile

	return imageInfo, ioutil.WriteFile(path.Base(imageInfo.URL), imageInfo.Bytes, 0755)
}

func FindLargerImageFromBytes(image []byte, outputFile string) (*ImageData, error) {
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

func GetLargerImageFromBytes(image []byte, outputDir string) (*ImageData, error) {
	var tmpfile = "GetLargerImageFromBytesTmpfile.image"
	defer os.Remove(tmpfile)

	var err = ioutil.WriteFile(tmpfile, image, 0755)
	if err != nil {
		return nil, err
	}
	return GetLargerImageFromFile(tmpfile, outputDir)
}
