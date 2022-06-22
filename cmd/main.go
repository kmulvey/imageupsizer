package main

import (
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/kmulvey/imageconvert/pkg/imageconvert"
	"github.com/kmulvey/imageupsizer"
	log "github.com/sirupsen/logrus"
)

func main() {
	var inputPath string
	var outputPath string
	var logLevel string
	flag.StringVar(&inputPath, "path", "", "A image file or directory path")
	flag.StringVar(&outputPath, "output", "", "Result output directory path")
	flag.StringVar(&logLevel, "log-level", "error", "Set the level of log output: (info, warn, error)")
	flag.Parse()

	switch strings.ToLower(logLevel) {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		flag.PrintDefaults()
	}

	if len(inputPath) == 0 || len(outputPath) == 0 {
		flag.PrintDefaults()
		return
	}

	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		log.Error("output path must be directory")
		return
	}

	if err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			switch filepath.Ext(path) {
			case ".jpg", ".png", ".jpeg", "webp":
				log.Infof("[%s] Getting original image info...", path)
				var originalImage, err = imageupsizer.GetImageConfigFromFile(path)
				if err != nil {
					return fmt.Errorf("GetImageConfigFromFile, %s, %w", path, err)
				}

				largerImage, err := imageupsizer.GetLargerImageFromFile(path, outputPath)
				if err != nil {
					return fmt.Errorf("GetLargerImageFromFile, %s, %w", path, err)
				}

				if largerImage.Area > originalImage.Width*originalImage.Height {
					var rename, _ = imageconvert.Convert(largerImage.LocalPath)

					// rename larger image to same name as original
					err = os.Rename(rename, filepath.Join(outputPath, filepath.Base(path)))
					if err != nil {
						return fmt.Errorf("replace old file, %s, %w", path, err)
					}
				}
				var areaIncrease = (largerImage.Area - originalImage.Area) / originalImage.Area
				var fileIncrease = (largerImage.FileSize - originalImage.FileSize) / originalImage.FileSize

				if fileIncrease > int64(areaIncrease) {
					log.WithFields(log.Fields{
						"path":               originalImage.LocalPath,
						"original area":      originalImage.Area,
						"new area":           largerImage.Area,
						"area increace":      fmt.Sprintf("%d%%", (largerImage.Area-originalImage.Area)/originalImage.Area),
						"file size increace": fmt.Sprintf("%d%%", (largerImage.FileSize-originalImage.FileSize)/originalImage.FileSize),
					}).Warn("upsized image is a lot bigger in file size")
				} else {
					log.WithFields(log.Fields{
						"path":               originalImage.LocalPath,
						"original area":      originalImage.Area,
						"new area":           largerImage.Area,
						"area increace":      fmt.Sprintf("%d%%", (largerImage.Area-originalImage.Area)/originalImage.Area),
						"file size increace": fmt.Sprintf("%d%%", (largerImage.FileSize-originalImage.FileSize)/originalImage.FileSize),
					}).Info("upsized image")
				}
			default:
				log.Infof("[%s] Skip !", path)
			}
		}

		return nil
	}); err != nil {
		log.Errorf("error getting larger image: %v", err)
	}
}
