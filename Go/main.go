package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var LogFileName string = "BNC.csv"

//Windows
var SearchDir string = "C:\\Users\\recs\\Documents\\ACTIF"

// var SearchDir string = "C:\\Users\\recs\\Documents\\ARCHIVE"

// Headers in csv files.
var Headers string = "path;part_id;number_bends;time_for_bends;with_adapter_bnc;comments\n"

// compile the regex patterns
var TagsRegx *regexp.Regexp = regexp.MustCompile("BEGIN_BIEGETEILSTAMM" + `((.|\n)*)` + "ENDE_BIEGETEILSTAMM")
var TagsFields *regexp.Regexp = regexp.MustCompile("ZA,DA,1" + `((.|\n)*)` + `(^C$\s)*`)
var TagsSubFields *regexp.Regexp = regexp.MustCompile("DA," + `((.|\n)*\d{1,2})` + `(.|\n)*`) // from DA to the end of the file.
var commas *regexp.Regexp = regexp.MustCompile(`([^0-9,']),([^0-9,'])`)

func createCsvAndHeaders(fileName string) {
	// Create a csv file in the same location of the script.
	err := ioutil.WriteFile(fileName, []byte(Headers), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func appendCsv(fileName, path, idNumber, bendsNumber, bendsTime, hasAdapter, comments string) {

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()
	if _, err := file.WriteString(path + ";" + idNumber + ";" + bendsNumber + ";" + bendsTime + ";" + hasAdapter + ";" + comments + "\n"); err != nil {
		log.Fatal(err)
	}
}

func getContentBetween(content string, regx regexp.Regexp) string {

	var bucket string = ""
	matches := regx.FindAllStringSubmatch(content, -1)
	for _, v := range matches {
		bucket = bucket + v[1]
	}
	return bucket
}

func findAdapterModufixInBnc(text, keyWord string) string {
	isAdapter := strings.Contains(text, keyWord)
	return strconv.FormatBool(isAdapter)
}

func getFieldsFromBNC(pathBNCFile string) ([4]string, error) {

	fmt.Println(pathBNCFile)

	//Get content of the bncPath
	content, err := ioutil.ReadFile(pathBNCFile)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	textfromBNC := string(content)

	// Check for Adapter
	hasAdapter := findAdapterModufixInBnc(textfromBNC, "Adapter Modufix") //Check if with adapter.

	biegeteilstamm := getContentBetween(textfromBNC, *TagsRegx)
	fields := getContentBetween(biegeteilstamm, *TagsFields)
	subFields := getContentBetween(fields, *TagsSubFields)

	//CARRIAGES:
	fieldswhitoutCarriage := strings.Replace(subFields, "\r", "", -1)
	fieldswhitoutNewLine := strings.Replace(fieldswhitoutCarriage, "\n", "", -1)

	//ASTERISK
	fieldsWithoutAsteriks := strings.Replace(fieldswhitoutNewLine, "*", "", -1)

	//ASTERISK
	fieldsSepCleaned := strings.Replace(fieldsWithoutAsteriks, "*", "", -1)

	//CLEANING QUOTATION MARKS & COMMA
	fieldsSepCleaned2 := strings.ReplaceAll(fieldsSepCleaned, "',  '", "','")
	fieldsSepCleaned3 := strings.ReplaceAll(fieldsSepCleaned2, ",  ", ",")

	//REMOVE COMMA FROM COMMENTS
	withoutCommaInComments := commas.ReplaceAllString(fieldsSepCleaned3, " virgule ")

	//SPLIT TO ARRAY
	fieldsSequence := strings.Split(withoutCommaInComments, ",")

	var sequence []string
	for _, i := range fieldsSequence {
		sequence = append(sequence, strings.TrimSpace(i))
	}

	var colData [4]string
	if len(sequence) > 18 {
		colData[0] = strings.TrimSpace(sequence[18])
		colData[1] = strings.TrimSpace(sequence[21])
		colData[2] = hasAdapter
		commentWithoutNewLine := strings.Replace(sequence[11], "\n", "", -1)            // remove new lines
		colData[3] = strings.Replace(commentWithoutNewLine, ";", " point-virgule ", -1) // remove new lines
		return colData, nil

	} else {
		return colData, errors.New("no data could scraped from this file")
	}

}

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(filepath.Base(fn), ".BNC")
}

func appendFile(logFile string, filePath string, data [4]string) {
	// get basename
	idNumber := FilenameWithoutExtension(filePath)
	appendCsv(logFile, filePath, idNumber, data[0], data[1], data[2], data[3])
}

func main() {
	LogFileName := os.Args[1]
	SearchDir := os.Args[2]

	fmt.Println("CSV Name:", LogFileName)
	fmt.Println("Target Folder:", SearchDir)
	start := time.Now()

	fileList := make([]string, 0)
	e := filepath.Walk(SearchDir, func(path string, f os.FileInfo, err error) error {
		// we filter only the .BNC files.
		if filepath.Ext(path) == ".BNC" {
			fileList = append(fileList, path)
		}
		return err
	})

	if e != nil {
		panic(e)
	}

	createCsvAndHeaders(LogFileName)

	for _, bncPath := range fileList {
		// skip files that are unvalide because of the commas in the comments
		bendData, err := getFieldsFromBNC(bncPath)
		if err != nil || bendData[1] == "" {
			appendFile(LogFileName, bncPath, [4]string{"failed", "failed", "failed", "failed"})
		} else {
			appendFile(LogFileName, bncPath, bendData)
		}

	}

	elapsed := time.Since(start)
	log.Printf("Scraping timer: %s", elapsed)
	log.Printf("%s created.", LogFileName)
}
