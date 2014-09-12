package main

import (
	"archive/zip"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	outputDir string
	zipFile   string
)

func main() {

	flag.StringVar(&outputDir, "outputDir", "twitpicsave", "")
	flag.StringVar(&zipFile, "zipFile", "", "")
	flag.Parse()

	if zipFile == "" {
		fmt.Printf("Please provide the path to the tweets.zip\n")
		flag.Usage()
		return
	}

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		fmt.Printf("Could not open provided zip :(")
	}

	defer r.Close()

	re := regexp.MustCompile("http://twitpic.com/[^, ]*")
	for _, f := range r.File {

		if f.Name == "tweets.csv" {
			fc, err := f.Open()
			if err != nil {
				fmt.Printf("Could not open the tweets.csv")
			}

			csvReader := csv.NewReader(fc)

			var foundURL []string
			for {
				record, err := csvReader.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					fmt.Println(err)
					return
				}

				foundURL = append(foundURL, re.FindAllString(record[5], -1)...)
				foundURL = append(foundURL, re.FindAllString(record[9], -1)...)

			}

			if len(foundURL) > 0 {
				fmt.Println("Length:", len(foundURL))
				err := os.Mkdir(outputDir, os.ModeDir|os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}

				DownloadImages(foundURL)

			}
			break

		}

	}

}

func DownloadImages(imageURLs []string) {

	for _, url := range imageURLs {

		var fullUrl string
		if strings.HasSuffix(url, "/full") {
			fullUrl = url
		} else {
			fullUrl = fmt.Sprintf("%s/full", url)
		}
		fmt.Printf("Downloading HTML at %s\n", fullUrl)

		doc, err := goquery.NewDocument(fullUrl)
		if err != nil {
			fmt.Println("Could not get the HTML at %v \n", err)
			continue
		}

		s := doc.Find("#media-full img").First()
		if val, exist := s.Attr("src"); exist {

			resp, err := http.Get(val)
			if err != nil {

				fmt.Printf("Could not retrieve twitpic at %s - Error:%v\n", val, err)
				continue
			}

			if data, err := ioutil.ReadAll(resp.Body); err == nil {

				interrogativIndex := strings.LastIndex(val, "?")

				outputName := filepath.Join(outputDir, filepath.Base(val[0:interrogativIndex]))

				fmt.Printf("Writing %s\n", outputName)
				err = ioutil.WriteFile(outputName, data, os.ModePerm)
				if err != nil {
					fmt.Printf("Could note write at %s\n - %v", outputName, err)
				}

			} else {
				fmt.Println("Could not retrieve file:", err)
			}

		}

	}
}
