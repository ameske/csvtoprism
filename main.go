package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Sample struct {
	Name string
	Data [3]int
}

type Experiment []Sample

func NewAdjustedExperiment(e Experiment) Experiment {
	c, err := e.Control()
	if err != nil {
		log.Fatal(err)
	}

	mean := float64((c.Data[0] + c.Data[1] + c.Data[2])) / 3.0

	var adjusted Experiment
	for i, _ := range e {
		var s Sample
		s.Name = e[i].Name
		for j := 0; j < 3; j++ {
			s.Data[j] = int(math.Floor(float64(e[i].Data[j]) - mean))
		}
	}

	return adjusted
}

// Sort moves the negative control samples to the front
func (e Experiment) Sort() {
	control := -1
	for i, s := range e {
		if strings.Contains(strings.ToLower(s.Name), "unpulsed") {
			control = i
			break
		}
	}

	if control != -1 {
		var tmp Sample
		tmp = e[0]
		e[0] = e[control]
		e[control] = tmp
	}
}

func (e Experiment) String() string {
	var b bytes.Buffer
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(&b, 0, 8, 0, '\t', 0)

	for i := 0; i < len(e); i++ {
		fmt.Fprintf(w, "%s\t", e[i].Name)
	}
	fmt.Fprintf(w, "\n")

	for i := 0; i < 3; i++ {
		for j := 0; j < len(e); j++ {
			fmt.Fprintf(w, "%d\t", e[j].Data[i])
		}
		fmt.Fprintf(w, "\n")
	}

	w.Flush()

	return b.String()
}

func (e Experiment) Control() (s Sample, err error) {
	control := -1
	for i, s := range e {
		if strings.Contains(strings.ToLower(s.Name), "unpulsed") {
			control = i
			break
		}
	}

	if control != -1 {
		return e[control], nil
	}

	return s, errors.New("no control found")
}

var rowMapping = map[string]int{"A": 0, "B": 1, "C": 2, "D": 3, "E": 4, "F": 5, "G": 6, "H": 7}

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

	e := reconstructExperiment(data, identifiers)
	fmt.Println(e)

	me := NewAdjustedExperiment(e)

	fd, err = os.Create("raw.csv")
	if err != nil {
		log.Fatal(err)
	}
	err = dumpExperiment(fd, e)
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
	err = dumpExperiment(fd, me)
	if err != nil {
		log.Fatal(err)
	}
	err = fd.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func reconstructExperiment(data []int, identifiers []string) Experiment {
	var e Experiment

	dCount := 0

	for _, identifier := range identifiers {
		var s Sample
		s.Name = identifier
		for j := 0; j < 3; j++ {
			s.Data[j] = data[dCount]
			dCount++
		}
		e = append(e, s)
	}

	return e
}

func dumpExperiment(w io.Writer, e Experiment) error {
	outputFormat := make([][]string, 4)
	outputFormat[0] = make([]string, 32)
	outputFormat[1] = make([]string, 32)
	outputFormat[2] = make([]string, 32)
	outputFormat[3] = make([]string, 32)

	col := 0
	for _, s := range e {
		outputFormat[0][col] = s.Name
		outputFormat[1][col] = strconv.Itoa(s.Data[0])
		outputFormat[2][col] = strconv.Itoa(s.Data[1])
		outputFormat[3][col] = strconv.Itoa(s.Data[2])
		col++
	}

	cw := csv.NewWriter(w)

	err := cw.WriteAll(outputFormat)
	if err != nil {
		return err
	}

	cw.Flush()

	return nil
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
