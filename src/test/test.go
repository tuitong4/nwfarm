package main

import (
	"os"
	"encoding/csv"
	"fmt"
)

func readcsv(filename string) ([][]string, error){
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	var records [][]string
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		var _records []string
		for _, _item := range record{
			//_records = append(_records, strings.TrimSpace(_item))
			_records = append(_records, _item)
		}
		records = append(records, _records)
	}

	return records, nil
}

func main(){
	records, _ := readcsv("./d.csv")

	for _, reacord := range records{
		fmt.Println(reacord)
	}
}
