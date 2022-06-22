package main

import (
	"flag"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	var path = flag.String("path", "", "A image file or directory path")
	var output = flag.String("output", "", "Result output directory path")
	var logLevel = flag.String("log-level", "error", "Set the level of log output: (info, warn, error)")
	flag.Parse()

	switch strings.ToLower(*logLevel) {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		flag.PrintDefaults()
	}

	if len(*path) == 0 || len(*output) == 0 {
		flag.PrintDefaults()
		return
	}

	if err := os.MkdirAll(*output, os.ModePerm); err != nil {
		log.Error("output path must be directory")
		return
	}

	if err := filepath.Walk(*path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			switch filepath.Ext(path) {
			case ".jpg", ".png", ".jpeg", "webp":
				log.Infof("[%s] Getting original image info...", path)

				for _, i := range data {
					if i.Area > originalImage.Width*originalImage.Height {

						break
					}
				}
			default:
				log.Infof("[%s] Skip !", path)
			}
		}

		return nil
	}); err != nil {
		log.Error("please change the file or directory path and try again")
	}
}

//log.Infof("[%s] Saved: %s (%dx%d -> %dx%d)", path, newFilename, originalImage.Width, originalImage.Height, imageInfo.Config.Width, imageInfo.Config.Height)
