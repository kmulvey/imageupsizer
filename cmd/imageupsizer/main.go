package main

import (
	"errors"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/kmulvey/humantime"
	"github.com/kmulvey/imageupsizer"
	"github.com/kmulvey/path"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	var inputPath path.Path
	var outputPath string
	var logLevel string
	var tr humantime.TimeRange
	flag.Var(&inputPath, "input", "path to files, globbing must be quoted")
	flag.StringVar(&outputPath, "output", "./upsize", "A directory to put the larger image in")
	flag.Var(&tr, "modified-since", "process files chnaged since this time")
	flag.StringVar(&logLevel, "log-level", "error", "Set the level of log output: (info, warn, error)")
	flag.Parse()

	switch strings.ToLower(logLevel) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		flag.PrintDefaults()
	}

	if len(inputPath.Files) == 0 {
		log.Error("path not provided")
		flag.PrintDefaults()
		return
	}

	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		log.Error("output path must be directory: ", outputPath)
		return
	}

	log.Info("building file list")
	files, err := getFileList(inputPath, tr)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("upsizing %d files", len(files))

	var signals = make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	log.WithFields(log.Fields{
		"inputDir":       inputPath.GivenInput,
		"outputDir":      outputPath,
		"modified-since": tr.From,
		"log-level":      logLevel,
	}).Info("Started")

	var warnings []logrus.Fields
	for _, path := range files {
		if len(signals) > 0 {
			log.Info("shutting down")
			break
		}

		largerImage, err := imageupsizer.GetLargerImageFromFile(path, outputPath)
		if err != nil {
			if errors.Is(err, imageupsizer.ErrNoLargerAvailable) || errors.Is(err, imageupsizer.ErrNoResults) || errors.Is(err, imageupsizer.OtherSizesNotAvailableError) || errors.Is(err, imageupsizer.NoMatchesError) {
				log.Tracef("[%s] Larger image not available", path)
				continue // we just keep going
			}
			log.Errorf("GetLargerImageFromFile, %s, %v", path, err)
			continue
		}

		originalImage, err := imageupsizer.GetImageConfigFromFile(path)
		if err != nil {
			log.Errorf("GetImageConfigFromFile, %s, %s", path, err.Error())
			continue
		}

		rename, _, err := imageupsizer.Convert(largerImage.LocalPath)
		if err != nil {
			log.Errorf("error converting image: %s, err: %v", path, err)
			continue
		}

		// rename larger image to same name as original
		err = os.Rename(rename, filepath.Join(outputPath, filepath.Base(path)))
		if err != nil {
			log.Errorf("replace old file, %s, %s", path, err.Error())
			continue
		}

		var areaIncrease = ((float64(largerImage.Area) - float64(originalImage.Area)) / float64(originalImage.Area)) * 100
		var fileIncrease = ((float64(largerImage.FileSize) - float64(originalImage.FileSize)) / float64(originalImage.FileSize)) * 100

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
	}

	for _, f := range warnings {
		log.WithFields(f).Warn("upsized image is a lot bigger in file size")
	}
}

// getFileList filters the file list
func getFileList(inputPath path.Path, modSince humantime.TimeRange) ([]string, error) {

	var nilTime = time.Time{}
	var err error
	var trimmedFileList = inputPath.Files

	if modSince.From != nilTime {
		trimmedFileList = path.FilterEntities(trimmedFileList, path.NewDateEntitiesFilter(modSince.From, modSince.To))
		if err != nil {
			return nil, fmt.Errorf("unable to filter files by skip map")
		}
	}

	trimmedFileList = path.FilterEntities(trimmedFileList, path.NewRegexEntitiesFilter(regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")))

	// these are all the files all the way down the dir tree
	return path.OnlyNames(trimmedFileList), nil
}
