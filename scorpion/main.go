package main

import "log"

func main() {
	if err := parseInfo(); err != nil {
		log.Fatal(err)
	}
}
