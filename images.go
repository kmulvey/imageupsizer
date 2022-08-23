package imageupsizer

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "golang.org/x/image/webp"
)

// ImageData represents all the information about an image in the app
type ImageData struct {
	URL       string
	Bytes     []byte
	Extension string
	image.Config
	Area      int
	FileSize  int64
	LocalPath string
}

// uploadImage uploads the given image to google images
// and returns the response as bytes.
func uploadImage(filename string) ([]byte, error) {
	var file, err = os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file; file: %s, error: %w", filename, err)
	}
	defer file.Close()
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading image contents; file: %s, error: %w", filename, err)
	}

	var buf = new(bytes.Buffer)
	var writer = multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("encoded_image", filename)
	if err != nil {
		return nil, fmt.Errorf("error creating html form; file: %s, error: %w", filename, err)
	}
	_, err = part.Write(fileContents)
	if err != nil {
		return nil, fmt.Errorf("error adding file to form; file: %s, error: %w", filename, err)
	}

	if err := writer.WriteField("image_url", ""); err != nil {
		return nil, fmt.Errorf("error adding form field image_url; file: %s, error: %w", filename, err)
	}
	if err := writer.WriteField("filename", ""); err != nil {
		return nil, fmt.Errorf("error adding form field filename; file: %s, error: %w", filename, err)
	}
	if err := writer.WriteField("hl", "en"); err != nil {
		return nil, fmt.Errorf("error adding form field hl; file: %s, error: %w", filename, err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing html form writer; file: %s, error: %w", filename, err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://images.google.com/searchbyimage/upload", buf)
	if err != nil {
		return nil, fmt.Errorf("error creating http request; file: %s, error: %w", filename, err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("origin", "https://images.google.com/")
	req.Header.Add("referer", "https://images.google.com/")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36")

	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending http request; file: %s, error: %w", filename, err)
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading resp.Body; file: %s, error: %w", filename, err)
	}

	return contents, nil
}

// getImage downloads the given image and returns the ImageData
// which includes the []byte.
func getImage(url string) (*ImageData, error) {
	var data = &ImageData{}

	var httpCient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	var req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating http req, url: %s, error: %w", url, err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:101.0) Gecko/20100101 Firefox/101.0")

	resp, err := httpCient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making http req, url: %s, error: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading resp.Body, url: %s, error: %w", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("non 2xx resp code: %d, url: %s", resp.StatusCode, url)
	}

	if strings.HasPrefix(resp.Header.Get("content-type"), "text/html") {
		return nil, errors.New("resp was html: " + url)
	}

	var cp = bytes.NewReader(body)
	imageDecode, ext, err := image.DecodeConfig(cp)
	if err != nil {
		return nil, fmt.Errorf("error decoding image config, length: %d, url: %s, error: %w", len(body), url, err)
	}

	data.URL = url
	data.Bytes = body
	data.Extension = ext
	data.Config = imageDecode
	data.Area = data.Config.Height * data.Config.Width
	data.FileSize = int64(len(body))

	return data, nil
}

// getLargestImage parses the search results and returns the
// largest image but does not download it.
func getLargestImage(contents []byte) (*ImageData, error) {
	var largeImgURL string
	var r = regexp.MustCompile(`(/search\?.*?simg:.*?)">`)

	for _, i := range r.FindAllStringSubmatch(string(contents), -1) {
		if len(i) < 2 {
			continue
		}

		if strings.Contains(i[1], ",isz:l") {
			largeImgURL = "https://google.com" + html.UnescapeString(i[1])
			break
		}
	}

	if len(largeImgURL) == 0 && bytes.Contains(contents, []byte("captcha")) {
		return nil, ErrCaptcha
	} else if len(largeImgURL) == 0 {
		return nil, ErrNoLargerAvailable
	}

	req, err := http.NewRequest(http.MethodGet, largeImgURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating http req, error: %w", err)
	}
	req.Header.Add("origin", "https://images.google.com/")
	req.Header.Add("referer", "https://images.google.com/")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36")

	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making http req, error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading resp.Body, error: %w", err)
	}
	var imgInfo = regexp.MustCompile(`\["(https://.*?.)",(\d+),(\d+)\]`)

	var data []*ImageData
	for _, arr := range imgInfo.FindAllStringSubmatch(string(body), -1) {
		if len(arr) < 4 {
			continue
		}

		urlUnquote, err := strconv.Unquote("\"" + arr[1] + "\"")
		if err != nil {
			continue
		}

		imgURL, err := url.Parse(urlUnquote)
		if err != nil {
			continue
		}

		imgHeight, err := strconv.Atoi(arr[2])
		if err != nil {
			continue
		}

		imgWidth, err := strconv.Atoi(arr[3])
		if err != nil {
			continue
		}

		data = append(data, &ImageData{
			URL:  imgURL.String(),
			Area: imgHeight * imgWidth,
		})
	}

	if len(data) == 0 {
		return nil, ErrNoResults
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Area > data[j].Area
	})

	return data[0], nil
}

// GetImageConfigFromFile returns ImageData for the given image
func GetImageConfigFromFile(filename string) (*ImageData, error) {
	var data = new(ImageData)
	data.LocalPath = filename

	var file, err = os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s, error: %w", filename, err)
	}
	defer file.Close()

	config, ext, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %s, error: %w", filename, err)
	}

	imageBody, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading image contents: %s, error: %w", filename, err)
	}

	data.Config = config
	data.Bytes = imageBody
	data.Extension = ext
	data.Area = config.Width * config.Height
	if stat, err := file.Stat(); err != nil {
		return nil, fmt.Errorf("error stat'ing file: %s, error: %w", filename, err)
	} else {
		data.FileSize = stat.Size()
	}

	return data, nil
}
