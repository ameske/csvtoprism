package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Sample struct {
	Name string
	Data [3]int
}

type Experiment []Sample

func NewExperiment(data []int, identifiers []string, sortOrder []string) Experiment {
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

	e.Sort(sortOrder)

	return e

}

func NewAdjustedExperiment(e Experiment) Experiment {
	c, err := e.Control()
	if err != nil {
		log.Fatal(err)
	}

	mean := (c.Data[0] + c.Data[1] + c.Data[2]) / 3

	var adjusted Experiment
	for i, _ := range e {
		var s Sample
		s.Name = e[i].Name
		for j := 0; j < 3; j++ {
			s.Data[j] = e[i].Data[j] - mean
		}
		adjusted = append(adjusted, s)
	}

	return adjusted
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

// Sort re-arranges the samples based on a given sort order
func (e Experiment) Sort(order []string) error {
	if len(e) != len(order) {
		return errors.New("sort order is not complete")
	}

	for _, s := range e {
		found := false
		for _, o := range order {
			if s.Name == o {
				found = true
				break
			}

			if !found {
				return errors.New("sort order missing experiment identifier")
			}
		}
	}

	// Use an in-place insertion sort to rearrange the experiments
	var tmp Sample
	for i, o := range order {
		idx := -1
		for j, s := range e {
			if s.Name == o {
				idx = j
			}
		}

		tmp = e[i]
		e[i] = e[idx]
		e[idx] = tmp
	}

	return nil
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

func (e Experiment) WriteCSV(w io.Writer) error {
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
