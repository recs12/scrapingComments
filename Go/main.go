package main

import (
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

//Linux:
// var SearchDir string = "/home/r3s2/Documents/BNC/"

//Windows
var SearchDir string = "C:\\Users\\recs\\Documents\\ACTIF"

// var SearchDir string = "C:\\Users\\recs\\Documents\\ARCHIVE"

// Headers in csv files.
var Headers string = "path, part_id, number_bends, time_for_bends, with_adapter_bnc,\n"

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

func appendCsv(fileName, path, idNumber, bendsNumber, bendsTime, hasAdapter string) {

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()
	if _, err := file.WriteString(path + ", " + idNumber + ", " + bendsNumber + ", " + bendsTime + ", " + hasAdapter + ",\n"); err != nil {
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

func findAdapterModufixInBnc(paragraph, element string) string {
	i := strings.Contains(paragraph, element)
	return strings.Title(strconv.FormatBool(i))
}

func TrimC(element string) string {
	// 'if the last letter is a C remove the character
	lc := strings.TrimSpace(element)
	lc = lc[len(lc)-1:]
	if lc == "C" {
		return lc[:len(lc)-1]
	} else {
		return lc
	}

}

func getCommentsFromFile(in <-chan string, out chan<- [3]string) {

	pathBNCFile := <-in
	fmt.Println(pathBNCFile)

	//Get content of the BNCPath
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
	regx := regexp.MustCompile("\n")
	fieldswhitoutCarriage := regx.ReplaceAllString(subFields, "")

	//ASTERISK
	fieldsWithoutAsteriks := strings.Replace(fieldswhitoutCarriage, "*", "", -1)

	//REMOVE THE QUOTATION MARKS
	fieldsWithoutQuotes := strings.ReplaceAll(fieldsWithoutAsteriks, "'", "")

	//SPLIT TO ARRAY
	fieldsSequence := strings.Split(fieldsWithoutQuotes, ",")
	var colData [3]string
	if len(fieldsSequence) > 21 {
		colData[0] = strings.TrimSpace(fieldsSequence[18])
		colData[1] = strings.TrimSpace(TrimC(fieldsSequence[21]))
		colData[2] = hasAdapter //Capitalize false -> False

	} else {
		colData[0] = "error reading"
		colData[1] = "error reading"
		colData[2] = "error reading"
	}
	out <- colData

}

func chunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for {
		if len(slice) == 0 {
			break
		}

		// necessary check to avoid slicing beyond
		// slice capacity
		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}

func FilenameWithoutExtension(fn string) string {

	return strings.TrimSuffix(filepath.Base(fn), ".BNC")
}

func appendFile(logFile string, filePath string, data [3]string) {
	// get basename
	idNumber := FilenameWithoutExtension(filePath)

	if "error reading" != "" {
		appendCsv(logFile, filePath, idNumber, data[0], data[1], data[2])
	} else {
		appendCsv(logFile, filePath, idNumber, "error", "error", "error")
	}
}

func main() {
	start := time.Now()

	out := make(chan [3]string)
	in := make(chan string)

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

	chunkedFileList := chunkSlice(fileList, 3)

	//chunk the array here
	for _, BNCPath := range chunkedFileList {

		// commentsFromFab, errScraping := getCommentsFromFile(BNCPath)
		go getCommentsFromFile(in, out)
		go getCommentsFromFile(in, out)
		go getCommentsFromFile(in, out)

		go func() {
			in <- BNCPath[0] //chunk[0]
			in <- BNCPath[1] //chunk[1]
			in <- BNCPath[2] //chunk[1]
		}()

		appendFile(LogFileName, BNCPath[0], <-out)
		appendFile(LogFileName, BNCPath[1], <-out)
		appendFile(LogFileName, BNCPath[2], <-out)
	}

	elapsed := time.Since(start)
	log.Printf("Scraping timer: %s", elapsed)
	log.Printf("%s created.", LogFileName)
}
