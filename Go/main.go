// add timer
// add loop through subFields
// add go in jupyther

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func GenerateCSV(fileName string) {
	err := ioutil.WriteFile(fileName, []byte("Path; FileName; Comments;\n"), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func AppendCSV(fileName, path, name, comments string) {
	//Append second line
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer file.Close()
	if _, err := file.WriteString(path + "; " + name + "; " + comments + ";\n"); err != nil {
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

func get_comments_from_file(pathBNCFile string) string {
	//Get content of the BNCPath
	content, err := ioutil.ReadFile(pathBNCFile)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	textfromBNC := string(content)
	biegeteilstamm := get_content_between_(textfromBNC, "BEGIN_BIEGETEILSTAMM", "ENDE_BIEGETEILSTAMM")
	fields := get_content_between_(biegeteilstamm, `ZA,DA,1`, `C`)
	subFields := get_content_between_(fields, `DA,`, `C`)

	//CARRIAGES:
	regx := regexp.MustCompile("\n")
	fieldswhitoutCarriage := regx.ReplaceAllString(subFields, "")

	//ASTERISK
	fieldsWithoutAsteriks := strings.Replace(fieldswhitoutCarriage, "*", "", -1)

	//REMOVE THE GUIMMET
	res := strings.ReplaceAll(fieldsWithoutAsteriks, "'", "")

	// split to array
	fieldsSequence := strings.Split(res, ",")
	return strings.TrimSpace(fieldsSequence[9])
}

func run() ([]string, error) {
	searchDir := "/home/r3s2/Documents/BNCGoLang/"

	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})

	if e != nil {
		panic(e)
	}

	GenerateCSV("actif.csv")
	for _, BNCPath := range fileList {
		// we filter only the .BNC files.
		if ".BNC" == filepath.Ext(BNCPath) {
			// code extract
			commentsFromFab := get_comments_from_file(BNCPath)
			BNCFileName := filepath.Base(BNCPath)
			AppendCSV("actif.csv", BNCPath, BNCFileName, commentsFromFab)
		}
	}

	return fileList, nil
}

func main() {
	run()
}
