package imageupsizer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

// FindLargerImageFromFile takes a file and returns information about
// a larger image that was found. It does NOT download the image.
func FindLargerImageFromFile(filename string) (*ImageData, error) {

	originalImage, err := GetImageConfigFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error from GetImageConfigFromFile: %w", err)
	}

	redirectHTML, err := uploadImage(filename)
	if err != nil {
		return nil, fmt.Errorf("error from uploadImage: %w", err)
	}

	redirectURL, err := getURLFromUploadResponse(redirectHTML)
	if err != nil {
		return nil, fmt.Errorf("error from getURLFromUploadResponse: %w", err)
	}

	foundURL, err := scrape(redirectURL.String(), findImageSourceLinkInHtml)
	if err != nil {
		return nil, fmt.Errorf("error from scrape found url: %w", err)
	}

	allSizesURL, err := scrape(foundURL.String(), findAllSizesLinkInHtml)
	if err != nil {
		return nil, fmt.Errorf("error from scrape all sizes: %w", err)
	}

	largestImageURL, err := scrape(allSizesURL.String(), findLargestImageLinkInHtml)
	if err != nil {
		return nil, fmt.Errorf("error from scrape largest image: %w", err)
	}

	largerImage, err := getImage(largestImageURL.String())
	if err != nil {
		return nil, fmt.Errorf("error from getImage: %w", err)
	}

	if largerImage.Area > originalImage.Width*originalImage.Height {
		return largerImage, nil
	}
	return nil, ErrNoLargerAvailable
}

// GetLargerImageFromFile is just like FindLargerImageFromFile except it also downloads the file.
func GetLargerImageFromFile(filename, outputDir string) (*ImageData, error) {
	var largerImage, err = FindLargerImageFromFile(filename)
	if err != nil {
		return nil, err
	}

	imageInfo, err := getImage(largerImage.URL)
	if err != nil {
		return nil, err
	}
	// some file names are crazy long and cant be a named FS file
	var largerImageName = cleanURL(path.Base(imageInfo.URL), imageInfo.Extension)

	var newFile = filepath.Join(outputDir, largerImageName)
	if err := os.WriteFile(newFile, imageInfo.Bytes, os.ModePerm); err != nil {
		return nil, err
	}
	imageInfo.LocalPath = newFile

	// is the image a known error img?
	errImg, err := isErrorImage(imageInfo)
	if err != nil {
		return nil, err
	}

	if errImg {
		if err := os.Remove(newFile); err != nil {
			return nil, err
		}
		return nil, ErrNoLargerAvailable
	}
	return imageInfo, nil
}

// FindLargerImageFromBytes takes a bytes and returns information about
// a larger image that was found. It does NOT download the image.
func FindLargerImageFromBytes(image []byte, outputFile string) (*ImageData, error) {
	var tmpfile = "FindLargerImageFromBytesTmpfile.image"
	defer os.Remove(tmpfile)

	var err = os.WriteFile(tmpfile, image, 0600)
	if err != nil {
		return nil, err
	}

	largerImage, err := FindLargerImageFromFile(tmpfile)
	if err != nil {
		return nil, err
	}

	return largerImage, os.Remove(tmpfile)
}

// GetLargerImageFromBytes is just like FindLargerImageFromBytes except it also downloads the file.
func GetLargerImageFromBytes(image []byte, outputDir string) (*ImageData, error) {
	var tmpfile = "GetLargerImageFromBytesTmpfile.image"
	defer os.Remove(tmpfile)

	var err = os.WriteFile(tmpfile, image, 0600)
	if err != nil {
		return nil, err
	}
	return GetLargerImageFromFile(tmpfile, outputDir)
}

func cleanURL(link, ext string) string {

	var re = regexp.MustCompile(`[^\w]`)

	var largerImageName = re.ReplaceAllString(link, "")

	if len(largerImageName) > 100 {
		largerImageName = largerImageName[:100] + "." + ext
	} else {
		largerImageName = largerImageName + "." + ext
	}

	return largerImageName
}
