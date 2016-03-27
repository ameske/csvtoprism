package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	filename := flag.String("f", "", "file to parse")
	flag.Parse()

	inputFD, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFD.Close()

	data, identifiers, err := parseInputData(inputFD)
	if err != nil {
		log.Fatal(err)
	}

	e := NewExperiment(data, identifiers)
	fmt.Println(e)

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
}

func parseInputData(r io.Reader) (data []int, identifiers []string, err error) {
	data = make([]int, 96)
	identifiers = make([]string, 32)

	scanner := bufio.NewScanner(r)
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
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return data, identifiers, nil
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
