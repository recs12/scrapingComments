package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var LogFileName string = "BNC.csv"

//Linux:
// var SearchDir string := "/home/r3s2/Documents/BNC/"

//Windows
// var SearchDir string := "C:\\Users\\recs\\Documents\\ACTIF"
var SearchDir string = "C:\\Users\\recs\\Documents\\ARCHIVE"

// Headers in csv files.
var Headers string = "path, part_id, number_bends,\n"

// compile the regex patterns
var TagsRegx *regexp.Regexp = regexp.MustCompile("BEGIN_BIEGETEILSTAMM" + `([^"]*)` + "ENDE_BIEGETEILSTAMM")
var TagsFields *regexp.Regexp = regexp.MustCompile("ZA,DA,1" + `([^"]*)` + "C")
var TagsSubFields *regexp.Regexp = regexp.MustCompile("DA," + `([^"]*)` + `\z`) // from DA to the end of the file.

func createCsvAndHeaders(fileName string) {
	// Create a csv file in the same location of the script.
	err := ioutil.WriteFile(fileName, []byte(Headers), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func appendCsv(fileName, path, id, datafield string) {

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()
	if _, err := file.WriteString(path + ", " + id + ", " + datafield + ",\n"); err != nil {
		log.Fatal(err)
	}
}

func getContentBetween(content string, regx regexp.Regexp) string {

	var modified_content string = ""
	matches := regx.FindAllStringSubmatch(content, -1)
	for _, v := range matches {
		modified_content = modified_content + v[1]
	}
	return modified_content
}

func get_comments_from_file(pathBNCFile string, out chan<- string) {

	fmt.Println(pathBNCFile)
	//Get content of the BNCPath
	content, err := ioutil.ReadFile(pathBNCFile)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	textfromBNC := string(content)
	biegeteilstamm := getContentBetween(textfromBNC, *TagsRegx)
	fields := getContentBetween(biegeteilstamm, *TagsFields)
	subFields := getContentBetween(fields, *TagsSubFields)

	//CARRIAGES:
	regx := regexp.MustCompile("\n")
	fieldswhitoutCarriage := regx.ReplaceAllString(subFields, "")

	//ASTERISK
	fieldsWithoutAsteriks := strings.Replace(fieldswhitoutCarriage, "*", "", -1)

	//REMOVE THE QUOTATION MARKS
	res := strings.ReplaceAll(fieldsWithoutAsteriks, "'", "")

	//SPLIT TO ARRAY
	fieldsSequence := strings.Split(res, ",")

	out <- strings.TrimSpace(fieldsSequence[18])

}

func main() {
	start := time.Now()

	out := make(chan string)

	fileList := make([]string, 0)
	e := filepath.Walk(SearchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})

	if e != nil {
		panic(e)
	}

	createCsvAndHeaders(LogFileName)

	for _, BNCPath := range fileList {

		// we filter only the .BNC files.
		if ".BNC" == filepath.Ext(BNCPath) {

			// id := strings.TrimSuffix(filepath.Base(BNCPath), ".BNC")

			// commentsFromFab, errScraping := get_comments_from_file(BNCPath)
			go get_comments_from_file(BNCPath, out)

			// if errScraping == nil {
			// 	appendCsv(LogFileName, BNCPath, id, commentsFromFab)
			// } else {
			// 	appendCsv(LogFileName, BNCPath, id, "READING ERROR")
			// }
			fmt.Println(<-out)
		}
	}

	elapsed := time.Since(start)
	log.Printf("Scraping timer: %s", elapsed)
	log.Printf("%s created.", LogFileName)
}
