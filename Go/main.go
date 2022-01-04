package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func GenerateCSV(fileName string) {
	// Create a csv file in the same location of the script.
	err := ioutil.WriteFile(fileName, []byte("Path, number_of_bends,\n"), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func AppendCSV(fileName, path, comments string) {

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()
	if _, err := file.WriteString(path + ", " + comments + ",\n"); err != nil {
		log.Fatal(err)
	}
}

func get_content_between_(content, starting, ending string) string {

	var modified_content string = ""
	regx := regexp.MustCompile(starting + `([^"]*)` + ending)
	matches := regx.FindAllStringSubmatch(content, -1)
	for _, v := range matches {
		modified_content = modified_content + v[1]
	}
	return modified_content
}

func get_comments_from_file(pathBNCFile string) (string, error) {

	fmt.Println(pathBNCFile)
	//Get content of the BNCPath
	content, err := ioutil.ReadFile(pathBNCFile)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	textfromBNC := string(content)
	biegeteilstamm := get_content_between_(textfromBNC, "BEGIN_BIEGETEILSTAMM", "ENDE_BIEGETEILSTAMM")
	fields := get_content_between_(biegeteilstamm, "ZA,DA,1", "C")
	subFields := get_content_between_(fields, "DA,", `\z`) // from DA to the end of the file.

	//CARRIAGES:
	regx := regexp.MustCompile("\n")
	fieldswhitoutCarriage := regx.ReplaceAllString(subFields, "")

	//ASTERISK
	fieldsWithoutAsteriks := strings.Replace(fieldswhitoutCarriage, "*", "", -1)

	//REMOVE THE QUOTATION MARKS
	res := strings.ReplaceAll(fieldsWithoutAsteriks, "'", "")

	//SPLIT TO ARRAY
	fieldsSequence := strings.Split(res, ",")
	if len(fieldsSequence) == 25 {
		return strings.TrimSpace(fieldsSequence[18]), nil
	} else {
		return "READING ERROR", errors.New("reading error")
	}
}

func run() ([]string, error) {

	logFileName := "BNC.csv"

	// Linux machine:
	// searchDir := "/home/r3s2/Documents/BNC/"

	//RDL machine:
	// searchDir := "C:\\Users\\recs\\OneDrive - Premier Tech\\Documents\\PT\\cmf\\BNC\\"
	//Windows machine:
	// searchDir := "C:\\Users\\recs\\Documents\\ACTIF"
	searchDir := "C:\\Users\\recs\\Documents\\ARCHIVE"



	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})

	if e != nil {
		panic(e)
	}

	GenerateCSV(logFileName)
	for _, BNCPath := range fileList {

		// we filter only the .BNC files.
		if ".BNC" == filepath.Ext(BNCPath) {

			commentsFromFab, errScraping := get_comments_from_file(BNCPath)

			if errScraping == nil {
				AppendCSV(logFileName, BNCPath, commentsFromFab)
			} else {
				AppendCSV(logFileName, BNCPath, "READING ERROR")
			}
		}
	}

	return fileList, nil
}

func main() {
	start := time.Now()
	run()
	elapsed := time.Since(start)
	log.Printf("BNC files scraping took %s", elapsed)
	log.Printf("BNC.csv created.")
}
