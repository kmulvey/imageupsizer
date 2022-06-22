package main

import (
	"bytes"
	"html"
	"image"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ImageData struct {
	URL       string
	Bytes     []byte
	Extension string
	image.Config
	Area     int
	FileSize int64
}

func uploadImage(filename string) ([]byte, error) {
	var file, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var buf = new(bytes.Buffer)
	var writer = multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("encoded_image", filename)
	if err != nil {
		return nil, err
	}
	_, err = part.Write(fileContents)
	if err != nil {
		return nil, err
	}

	if err := writer.WriteField("image_url", ""); err != nil {
		return nil, err
	}
	if err := writer.WriteField("filename", ""); err != nil {
		return nil, err
	}
	if err := writer.WriteField("hl", "en"); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://images.google.com/searchbyimage/upload", buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("origin", "https://images.google.com/")
	req.Header.Add("referer", "https://images.google.com/")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36")

	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func getImage(url string) (*imageData, error) {
	var data = &imageData{}

	var resp, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	imageDecode, ext, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	data.URL = url
	data.Bytes = body
	data.Extension = ext
	data.Config = imageDecode
	data.Area = data.Config.Height * data.Config.Width
	data.FileSize = int64(len(body))

	return data, nil
}

func getImageList(contents []byte) (*imageData, error) {
	var largeImgURL string
	var r, err = regexp.Compile(`(/search\?.*?simg:.*?)">`)
	if err != nil {
		return nil, err
	}

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
		return nil, errCaptcha
	} else if len(largeImgURL) == 0 {
		return nil, errNoLargerAvailable
	}

	req, err := http.NewRequest(http.MethodGet, largeImgURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("origin", "https://images.google.com/")
	req.Header.Add("referer", "https://images.google.com/")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36")

	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	imgInfo, err := regexp.Compile(`\["(https://.*?.)",(\d+),(\d+)\]`)
	if err != nil {
		return nil, err
	}

	var data []*imageData
	for _, i := range imgInfo.FindAllStringSubmatch(string(body), -1) {
		if len(i) < 4 {
			continue
		}

		urlUnquote, err := strconv.Unquote("\"" + i[1] + "\"")
		if err != nil {
			continue
		}

		imgURL, err := url.Parse(urlUnquote)
		if err != nil {
			continue
		}

		imgHeight, err := strconv.Atoi(i[2])
		if err != nil {
			continue
		}

		imgWidth, err := strconv.Atoi(i[3])
		if err != nil {
			continue
		}

		data = append(data, &imageData{
			URL:  imgURL.String(),
			Area: imgHeight * imgWidth,
		})
	}

	if len(data) == 0 {
		return nil, errNoResults
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Area > data[j].Area
	})

	return data[0], nil
}

func getImageConfigFromFile(filename string) (*imageData, error) {
	var data = new(imageData)

	var file, err = os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config, ext, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	imageBody, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	data.Config = config
	data.Bytes = imageBody
	data.Extension = ext
	data.Area = config.Width * config.Height
	if stat, err := file.Stat(); err != nil {
		return nil, err
	} else {
		data.FileSize = stat.Size()
	}

	return data, nil
}
