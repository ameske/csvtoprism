package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	filename := flag.String("f", "", "file to parse")
	flag.Parse()

	inputFD, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFD.Close()

	fmt.Printf("Input File: %s\n", *filename)

	e, err := NewExperiment(inputFD)
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
