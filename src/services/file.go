package services

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	u "gsheet-to-json-csv/src/utils"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Should point to the folder where language json files are kept
const outputPath = "../outputs/"

// Download exported
func Download(url string, filename string, timeout int64) *u.ErrorResponse {
	u.GeneralLogger.Println("Downloading", url, "...")
	client := http.Client{
		Timeout: time.Duration(timeout * int64(time.Second)),
	}
	resp, err := client.Get(url)
	if err != nil {
		u.ErrorLogger.Println("Cannot download file from the given url", err)
		return u.ReturnErrorResponse(err, "Cannot download file from the given url")
	}

	if resp.StatusCode != 200 {
		u.ErrorLogger.Printf("Response from the URL was %d, but expecting 200", resp.StatusCode)
		return u.ReturnErrorResponse(
			errors.New("Response returned with a status different from 200"),
			"Response returned with a status different from 200",
		)
	}
	if resp.Header["Content-Type"][0] != "text/csv" {
		u.ErrorLogger.Printf("The file downloaded has content type '%s', expected 'text/csv'.", resp.Header["Content-Type"])
		return u.ReturnErrorResponse(
			errors.New("Downloaded file didn't contain the expected content-type: 'text/csv'"),
			"Downloaded file didn't contain the expected content-type: 'text/csv'",
		)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		u.ErrorLogger.Println("Cannot read Body of Response", err)
		return u.ReturnErrorResponse(err, "Cannot read Body of Response")
	}

	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		u.ErrorLogger.Println("Cannot write to file", err)
		return u.ReturnErrorResponse(err, "Cannot write to file")
	}

	u.GeneralLogger.Println("Doc downloaded in ", filename)

	return u.ReturnErrorResponse(nil, "")
}

// WriteLanguageFiles exported
func WriteLanguageFiles(csvFilePath string) *u.ErrorResponse {
	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		u.ErrorLogger.Println("Cannot open file:"+csvFilePath, err)
		return u.ReturnErrorResponse(err, "Cannot open file:"+csvFilePath)

	}

	csvFileContent, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return u.ReturnErrorResponse(err, "Cannot read file:"+csvFilePath)
	}
	for i, lang := range csvFileContent[0][1:] {
		absPath, err := filepath.Abs(outputPath + lang + ".json")
		if err != nil {
			u.ErrorLogger.Println("Cannot get path specified: \""+lang+".json\"", err)
			return u.ReturnErrorResponse(err, "Cannot get path specified: \""+lang+".json\"")
		}

		file, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			u.ErrorLogger.Println("Cannot open file: \""+lang+".json\"", err)
			return u.ReturnErrorResponse(err, "Cannot open file: \""+lang+".json\"")
		}
		err = file.Truncate(0)
		if err != nil {
			return u.ReturnErrorResponse(err, "Cannot truncate file: \""+lang+".json\"")
		}
		mapLn := map[string]string{}
		u.GeneralLogger.Println("Language:", lang, i)
		for j, row := range csvFileContent[1:] {
			// fmt.Println(csvFileContent[j+1][0], row[i+1])
			mapLn[csvFileContent[j+1][0]] = row[i+1]
		}
		encodedJSON, _ := json.Marshal(mapLn)
		// u.GeneralLogger.Println(string(encodedJSON))
		_, err = file.Write(encodedJSON)
		if err != nil {
			return u.ReturnErrorResponse(err, "Cannot write to file: \""+lang+".json\"")
		}
		file.Close()
	}
	return u.ReturnErrorResponse(nil, "")
}
