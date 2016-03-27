package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	filename := flag.String("f", "", "file to parse")
	flag.Parse()

	fd, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	data := make([]int, 96)
	identifiers := make([]string, 32)

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()

		cells := strings.Split(line, ",")

		switch determineLineType(cells) {
		case Data:
			parsedData, row := parseDataRow(cells)
			for i, d := range parsedData {
				data[row*12+i] = d
			}
		case Identifiers:
			parsedIdentifiers, row := parseIdentifierRow(cells)
			for i, id := range parsedIdentifiers {
				identifiers[row*4+i] = id
			}
		default:
			continue
		}
	}
	fd.Close()

	e := NewExperiment(data, identifiers)
	fmt.Println(e)

	me := NewAdjustedExperiment(e)

	fd, err = os.Create("raw.csv")
	if err != nil {
		log.Fatal(err)
	}
	err = e.WriteCSV(fd)
	if err != nil {
		log.Fatal(err)
	}
	err = fd.Close()
	if err != nil {
		log.Fatal(err)
	}

	fd, err = os.Create("adjusted.csv")
	if err != nil {
		log.Fatal(err)
	}

	err = me.WriteCSV(fd)
	if err != nil {
		log.Fatal(err)
	}
	err = fd.Close()
	if err != nil {
		log.Fatal(err)
	}
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
