package main

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kmulvey/imageupsizer"
	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

func main() {
	var newFilesEntry path.Entry
	var oldFilesEntry path.Entry
	var logLevel string
	flag.Var(&newFilesEntry, "new-files", "path to files, globbing must be quoted")
	flag.Var(&oldFilesEntry, "old-files", "A directory to put the larger image in")
	flag.StringVar(&logLevel, "log-level", "info", "Set the level of log output: (info, warn, error)")
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

	var newFiles, err = newFilesEntry.Flatten()
	if err != nil {
		log.Fatalf("error flattening newFiles: %s", err)
	}

	if len(newFiles) == 0 {
		log.Error("path not provided")
		flag.PrintDefaults()
		return
	}

	if err := os.MkdirAll(oldFilesEntry.String(), os.ModePerm); err != nil {
		log.Error("output path must be directory: ", oldFilesEntry.String())
		return
	}

	var files = path.FilterEntities(newFiles, path.NewRegexEntitiesFilter(regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")))

	for _, file := range files {
		var newImage, err = imageupsizer.GetImageConfigFromFile(file.AbsolutePath)
		if err != nil {
			log.Errorf("GetImageConfigFromFile, %s, %s", file.AbsolutePath, err.Error())
			continue
		}
		oldImage, err := imageupsizer.GetImageConfigFromFile(filepath.Join(oldFilesEntry.String(), filepath.Base(file.AbsolutePath)))
		if err != nil {
			log.Errorf("GetImageConfigFromFile, %s, %s", file.AbsolutePath, err.Error())
			continue
		}

		if newImage.Area > oldImage.Area {
			err = os.Rename(newImage.LocalPath, filepath.Join(oldFilesEntry.String(), filepath.Base(file.AbsolutePath)))
			if err != nil {
				log.Errorf("rename %s to %s, err: %s", newImage.LocalPath, filepath.Join(oldFilesEntry.String(), filepath.Base(file.AbsolutePath)), err.Error())
				continue
			}
			log.WithFields(log.Fields{
				"old":  oldImage.Area,
				"new":  newImage.Area,
				"from": newImage.LocalPath,
				"to":   filepath.Join(oldFilesEntry.String(), filepath.Base(newImage.LocalPath)),
			}).Info("move")
		}
	}
}
