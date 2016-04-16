package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/tealeg/xlsx"
)

func main() {
	filename := flag.String("f", "", "file to parse")
	flag.Parse()

	var csv bytes.Buffer

	fmt.Printf("Input File: %s\n", *filename)

	err := generateCSVFromXLSXFile(*filename, 0, &csv)
	if err != nil {
		log.Fatal(err)
	}

	e, err := NewExperiment(&csv)
	if err != nil {
		log.Fatal(err)
	}

	me := NewAdjustedExperiment(e)

	rawFD, err := os.Create("raw.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer rawFD.Close()

	err = e.WriteCSV(rawFD)
	if err != nil {
		log.Fatal(err)
	}

	adjFD, err := os.Create("adjusted.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer adjFD.Close()

	err = me.WriteCSV(adjFD)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Succesfully generated raw and adjusted data sets.")
}

func dumpDataTable(data []int) {
	for i, d := range data {
		if i%12 == 0 {
			fmt.Printf("\n%d ", d)
		} else {
			fmt.Printf("%d ", d)
		}
	}
}

// Adapted from https://github.com/tealeg/xlsx2csv
func generateCSVFromXLSXFile(excelFileName string, sheetIndex int, w io.Writer) error {
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		return err
	}
	sheetLen := len(xlFile.Sheets)
	switch {
	case sheetLen == 0:
		return errors.New("This XLSX file contains no sheets.")
	case sheetIndex >= sheetLen:
		return fmt.Errorf("No sheet %d available, please select a sheet between 0 and %d\n", sheetIndex, sheetLen-1)
	}
	sheet := xlFile.Sheets[sheetIndex]
	for _, row := range sheet.Rows {
		var vals []string
		if row != nil {
			for _, cell := range row.Cells {
				str, err := cell.String()
				if err != nil {
					vals = append(vals, err.Error())
				}
				vals = append(vals, fmt.Sprintf("%s", str))
			}
			fmt.Fprintf(w, strings.Join(vals, ",")+"\n")
		}
	}
	return nil
}
