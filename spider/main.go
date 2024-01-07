package main

import (
	"errors"
	"log"
	"os"
)

func main() {
	programFlags, err := initFlags()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(*programFlags.pFlag); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(*programFlags.pFlag, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := SearchForImageLinks(*programFlags); err != nil {
		log.Fatal(err)
	}
}
