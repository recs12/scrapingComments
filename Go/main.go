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

func get_comments_from_file(pathBNCFile string) (string, error) {

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

	/*
	    [0]'Tabellenidentifikator'
	    [1]'ID'
	    [2]'Zeichnungsnummer'
	    [3]'Bezeichnung'
	    [4]'Variantenbezeichnung'
	    [5]'Programmversion'
	    [6]'Erstellungsdatum'
	    [7]'Aenderungsdatum'
	    [8]'Produktionsdatum'
	    [9]'Bearbeiter'
	    [10]'Klassifizierung'
	   *[11]'Kommentar'
	    [12]'Material'
	    [13]'Blechdicke'
	    [14]'Gewicht'
	    [15]'UmschreibendesRechteckX'
	    [16]'UmschreibendesRechteckY'
	    [17]'Herkunft'
	   *[18]'AnzahlBiegeschritte'
	    [19]'ToPs-Programmname'
	    [20]'Zuordnung VorschauID'
	    [21]'Bearbeitungszeit'
	    [22]'TeilabmessungX'
	    [23]'TeilabmessungY'
	    [24]'TeilabmessungZ'
	*/

	//SPLIT TO ARRAY
	fieldsSequence := strings.Split(res, ",")
	if len(fieldsSequence) > 18 {
		return strings.TrimSpace(fieldsSequence[18]), nil
	} else {
		return "READING ERROR", errors.New("reading error")
	}
}

func run() ([]string, error) {

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

			id := strings.TrimSuffix(filepath.Base(BNCPath), ".BNC")

			commentsFromFab, errScraping := get_comments_from_file(BNCPath)

			if errScraping == nil {
				appendCsv(LogFileName, BNCPath, id, commentsFromFab)
			} else {
				appendCsv(LogFileName, BNCPath, id, "READING ERROR")
			}
		}
	}

	return fileList, nil
}

func main() {
	start := time.Now()
	run()
	elapsed := time.Since(start)
	log.Printf("Scraping timer: %s", elapsed)
	log.Printf("%s created.", LogFileName)
}
