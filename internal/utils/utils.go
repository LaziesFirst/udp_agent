package utils

import (
	"bufio"
	"encoding/csv"
	"os"
)

func MakeDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func MakeFileIfNotExists(fileName string, headers interface{}) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		if err = CreateCsvFile(fileName, headers); err != nil {
			return err
		}
	}
	return nil
}

func CreateCsvFile(fileName string, headers interface{}) error {
	stringHeaders := []string{}
	for _, val := range headers.([]interface{}) {
		stringHeaders = append(stringHeaders, val.(string))
	}
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		return err
	}
	wr := csv.NewWriter(bufio.NewWriter(file))
	wr.Write(stringHeaders)
	wr.Flush()
	return nil
}
