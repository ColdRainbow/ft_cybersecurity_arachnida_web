package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Flags struct {
	url     *url.URL
	baseUrl string
	rFlag   *bool
	lFlag   *int
	pFlag   *string
}

func initFlags() (*Flags, error) {
	progFlags := Flags{}
	progFlags.rFlag = flag.Bool("r", false, "recursion")
	progFlags.lFlag = flag.Int("l", 0, "depth")
	progFlags.pFlag = flag.String("p", "./data", "path")
	flag.Parse()
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "l" {
			found = true
		}
	})
	if *progFlags.rFlag == false && found {
		return nil, errors.New("Wrong args")
	}
	if *progFlags.lFlag == 0 && !found {
		*progFlags.lFlag = 5
	}
	if *progFlags.lFlag < 0 {
		return nil, errors.New("Wrong value of l")
	}
	var err error
	if flag.NArg() != 1 {
		return nil, errors.New("Wrong number of arguments")
	}

	if flag.Arg(0) == "" {
		return nil, errors.New("No URL specified")
	}
	progFlags.url, err = url.Parse(flag.Arg(0))
	if err != nil {
		return nil, err
	}
	progFlags.baseUrl = progFlags.url.Scheme + "://" + progFlags.url.Host
	return &progFlags, nil
}

func DownloadImages(subMatchSlice [][]string, currentArgs Flags) error {
	for _, item := range subMatchSlice {
		if strings.HasPrefix(item[1], "http") || strings.HasPrefix(item[1], "//") {
			continue
		}
		log.Println("Image found : ", item[1])

		curImagePath, err := url.JoinPath(currentArgs.baseUrl, item[1])
		if err != nil {
			return err
		}
		imageResp, err := http.Get(curImagePath)
		if err != nil {
			return err
		}
		if imageResp.StatusCode != 200 {
			return fmt.Errorf("Unable to get data for an image %s", curImagePath)
		}

		imagePath := filepath.Join(*currentArgs.pFlag, filepath.Base(item[1]))

		valid := false
		correctExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}
		for _, ext := range correctExtensions {
			if strings.HasSuffix(imagePath, ext) == true {
				valid = true
				break
			}
		}

		if valid == false {
			continue
		}

		bodyData, err := ioutil.ReadAll(imageResp.Body)
		if err != nil {
			return fmt.Errorf("Cannot read body of the image %s", imagePath)
		}

		err = ioutil.WriteFile(imagePath, bodyData, 0644)
		if err != nil {
			return fmt.Errorf("Cannot download image %s", imagePath)
		}
		imageResp.Body.Close()
	}
	return nil
}

func SearchForImageLinks(currentArgs Flags) error {
	log.Println("Parsing : ", currentArgs.url)

	// Request the HTML page.
	resp, err := http.Get(currentArgs.url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Unable to get URL with status code error: %d %s", resp.StatusCode, resp.Status)
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	bodyCopy := bytes.NewReader(htmlData)
	imageRegExp := regexp.MustCompile(`<img[^>]+\bsrc=["']([^"']+)["']`)

	subMatchSlice := imageRegExp.FindAllStringSubmatch(string(htmlData), -1)

	if err := DownloadImages(subMatchSlice, currentArgs); err != nil {
		return err
	}

	if *currentArgs.rFlag == true && *currentArgs.lFlag != 0 {
		file, err := goquery.NewDocumentFromReader(bodyCopy)
		if err != nil {
			return errors.New("Cannot get document from reader")
		}

		childUrl := file.Find("a")
		for i := range childUrl.Nodes {
			single := childUrl.Nodes[i]
			for j := range single.Attr {
				//fmt.Printf("Current Attr is %v\n", single.Attr[j])
				if single.Attr[j].Key == "href" && single.Attr[j].Val[0] == '/' {
					newL := *currentArgs.lFlag - 1
					newArgs := Flags{
						url:   currentArgs.url.JoinPath(single.Attr[j].Val),
						pFlag: currentArgs.pFlag,
						rFlag: currentArgs.rFlag,
						lFlag: &newL,
					}
					SearchForImageLinks(newArgs)
				}
			}
		}
	}
	return nil
}
