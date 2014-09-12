package main

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/andlabs/ui"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	w         ui.Window
	fileLabel ui.Label
	fileURL   string
)

func initGUI() {

	fileLabel = ui.NewStandaloneLabel("File...")
	b := ui.NewButton("Browse")
	b.OnClicked(func() {

		ui.OpenFile(w, func(filename string) {

			fileLabel.SetText(filename)
			fileURL = filename
		})
	})
	s := ui.NewButton("Start")
	s.OnClicked(func() {

		ParseTweetFile(fileURL)

	})

	h := ui.NewHorizontalStack(fileLabel, b, s)
	h.SetStretchy(0)

	w = ui.NewWindow("golang twitpic", 500, 50, h)
	w.OnClosing(func() bool {
		ui.Stop()
		return true
	})
	w.Show()
}

func ParseTweetFile(zipFile string) {
	if zipFile == "" {
		fmt.Printf("Please provide the path to the tweets.zip\n")
		fileLabel.SetText("Could not proceed file")
		return
	}

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		fmt.Printf("Could not open provided zip :(")
		fileLabel.SetText("Could not open file. Please provide another one")
		return
	}

	defer r.Close()

	re := regexp.MustCompile("http://twitpic.com/[^, ]*")
	for _, f := range r.File {

		if f.Name == "tweets.csv" {
			fc, err := f.Open()
			if err != nil {
				fmt.Printf("Could not open the tweets.csv")
				fileLabel.SetText("Could not open the tweets.csv file")
				return
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
				usr, _ := user.Current()

				downloadDir := filepath.Join(usr.HomeDir, "twitpicsave")
				err := os.Mkdir(downloadDir, os.ModeDir|os.ModePerm)
				if err != nil {
					fileLabel.SetText(fmt.Sprintf("Could not create directory at %s", downloadDir))
					return
				}

				go DownloadImages(foundURL, downloadDir)

			}
			break

		}

	}
}
func main() {

	go ui.Do(initGUI)
	err := ui.Go()
	if err != nil {
		panic(err)
	}
}

func DownloadImages(imageURLs []string, outputDir string) {

	for _, url := range imageURLs {

		var fullUrl string
		if strings.HasSuffix(url, "/full") {
			fullUrl = url
		} else {
			fullUrl = fmt.Sprintf("%s/full", url)
		}
		fileLabel.SetText(fmt.Sprintf("Downloading HTML at %s\n", fullUrl))

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

		fileLabel.SetText("Finished")
	}
}
