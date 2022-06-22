package main

import (
	"errors"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/kmulvey/imageconvert/pkg/imageconvert"
	"github.com/kmulvey/imageupsizer"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	var inputPath string
	var outputPath string
	var logLevel string
	flag.StringVar(&inputPath, "input", "", "A image file or directory path")
	flag.StringVar(&outputPath, "output", "", "A directory to put the larger image in")
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
		log.Error("output path must be directory: ", outputPath)
		return
	}

	log.WithFields(log.Fields{
		"inputDir":  inputPath,
		"outputDir": outputPath,
		"log-level": logLevel,
	}).Info("Started")

	var warnings []logrus.Fields
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
					if errors.Is(err, imageupsizer.ErrNoLargerAvailable) || errors.Is(err, imageupsizer.ErrNoResults) {
						log.Infof("[%s] Larger image not available", path)
						return nil // we just keep going
					}
					log.Errorf("GetLargerImageFromFile, %s, %v", path, err)
					return nil
				}

				if largerImage.Area > originalImage.Width*originalImage.Height {
					var rename, _ = imageconvert.Convert(largerImage.LocalPath)

					// rename larger image to same name as original
					err = os.Rename(rename, filepath.Join(outputPath, filepath.Base(path)))
					if err != nil {
						return fmt.Errorf("replace old file, %s, %w", path, err)
					}
				}
				var areaIncrease = (float64(largerImage.Area) - float64(originalImage.Area)) / float64(originalImage.Area)
				var fileIncrease = (float64(largerImage.FileSize) - float64(originalImage.FileSize)) / float64(originalImage.FileSize)

				if fileIncrease > areaIncrease {
					warnings = append(warnings, log.Fields{
						"path":               originalImage.LocalPath,
						"original area":      originalImage.Area,
						"new area":           largerImage.Area,
						"area increace":      fmt.Sprintf("%.2f%%", areaIncrease),
						"file size increace": fmt.Sprintf("%.2f%%", fileIncrease),
					})
				}
				log.WithFields(log.Fields{
					"path":               originalImage.LocalPath,
					"original area":      originalImage.Area,
					"new area":           largerImage.Area,
					"area increace":      fmt.Sprintf("%.2f%%", areaIncrease),
					"file size increace": fmt.Sprintf("%.2f%%", fileIncrease),
				}).Info("upsized image")
			default:
				log.Infof("[%s] Skip !", path)
			}
		}

		return nil
	}); err != nil {
		log.Errorf("error in walk loop: %v", err)
	}

	for _, f := range warnings {
		log.WithFields(f).Warn("upsized image is a lot bigger in file size")
	}
}
