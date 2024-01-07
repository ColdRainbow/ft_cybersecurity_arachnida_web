package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	//"github.com/dsoprea/go-exif/v3"
	"github.com/rwcarlsen/goexif/exif"
)

func printMetaData(name string, metaData *exif.Exif, creationDate time.Time) {
	fmt.Printf("Info for image %s:\n", name)
	fmt.Println("Creation date is:")
	fmt.Println(creationDate)
	fmt.Printf("Other metadata:")
	fmt.Print(metaData)
}

func parseInfo() error {
	args := os.Args
	if len(args) < 2 {
		return errors.New("Not enough arguments")
	}
	correctExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}
	for _, img := range args[1:] {
		valid := false
		for _, ext := range correctExtensions {
			if strings.HasSuffix(img, ext) {
				valid = true
				break
			}
		}
		if valid == false {
			continue
		}

		f, err := os.Open(img)
		if err != nil {
			return err
		}

		metaData, err := exif.Decode(f)
		if err != nil {
			if err == io.EOF {
				return errors.New("No exif data")
			}
			return err
		}

		creationDate, err := metaData.DateTime()
		if err != nil {
			return err
		}

		printMetaData(img, metaData, creationDate)
	}
	return nil
}
