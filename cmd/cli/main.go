package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ameske/csvtoprism"
)

func main() {
	filename := flag.String("f", "", "file to parse")
	flag.Parse()

	var csv bytes.Buffer

	fmt.Printf("Input File: %s\n", *filename)

	filePart := strings.Split(*filename, ".")[0]

	err := csvtoprism.CSVFromXLS(*filename, 0, &csv)
	if err != nil {
		log.Fatal(err)
	}

	e, err := csvtoprism.NewExperiment(&csv)
	if err != nil {
		log.Fatal(err)
	}

	me := csvtoprism.NewAdjustedExperiment(e)

	rawFD, err := os.Create(fmt.Sprintf(filePart + "_Raw.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer rawFD.Close()

	err = e.WriteCSV(rawFD)
	if err != nil {
		log.Fatal(err)
	}

	adjFD, err := os.Create(fmt.Sprintf(filePart + "_Adjusted.csv"))
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
