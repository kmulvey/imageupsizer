package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"
)

func main() {
	var upsizedImagesEntry path.Entry
	var originalImagesEntry path.Entry
	var v bool
	var help bool
	flag.Var(&upsizedImagesEntry, "upsized-images", "directory of images from cmd/imageupsizer")
	flag.Var(&originalImagesEntry, "original-images", "directory of source images")
	flag.BoolVar(&help, "help", false, "print help")
	flag.BoolVar(&v, "version", false, "print version")
	flag.BoolVar(&v, "v", false, "print version")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if v {
		var verPrinter = printer.New()
		var info = version.Get()
		if err := verPrinter.PrintInfo(os.Stdout, info); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	var upsizedFiles, err = upsizedImagesEntry.Flatten()
	if err != nil {
		log.Fatal("error getting upsized files: %s", err)
		os.Exit(1)
	}

	for i, upsizedImage := range upsizedFiles {
		var originalImage = filepath.Join(originalImagesEntry.String(), filepath.Base(upsizedImage.AbsolutePath))

		// we could has already deleted one of them, so just go around
		if !fileExists(upsizedImage.AbsolutePath) {
			fmt.Println(upsizedImage.AbsolutePath, " already deleted")
			continue
		}
		if !fileExists(originalImage) {
			fmt.Println(originalImage, " already deleted")
			continue
		}

		var viewerCmd string
		var goos = runtime.GOOS
		switch goos {
		case "windows":
		case "darwin":
			viewerCmd = "preview"
		case "linux":
			viewerCmd = "eog" // eog -- GNOME Image Viewer 41.1
		default:
			log.Fatalf("unsupported os: %s", goos)
		}
		// open both images with image viewer
		cmd := exec.Command(viewerCmd, upsizedImage.AbsolutePath)
		var err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		cmdS := exec.Command(viewerCmd, originalImage)
		err = cmdS.Run()
		if err != nil {
			log.Fatal(err)
		}

		// ask the user if we should delete
		var del string
		fmt.Printf("[%d/%d]	accept: %s ?", i+1, len(upsizedFiles), originalImage)
		fmt.Scanln(&del)
		if del == "n" {
			err = os.Remove(upsizedImage.AbsolutePath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("deleted", upsizedImage.AbsolutePath)
		}
	}
}

// fileExists returns true if the file exists
func fileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err == nil {
		return true
	}
	return false
}
