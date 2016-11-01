package csvtoprism

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

// NewExperiment constructs a new Experiement using data provided by the given io.Reader
func NewExperiment(r io.Reader) (Experiment, error) {
	data, identifiers, sortOrder, err := parseInputData(r)
	if err != nil {
		return nil, err
	}

	return NewExperimentFromRawData(data, identifiers, sortOrder), nil
}

// NewExperimentFromRawData takes slice representations of the various parts of an experiment and
// returns an Experiment struct.
func NewExperimentFromRawData(data []int, identifiers []string, sortOrder []int) Experiment {
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

	if len(sortOrder) != 0 {
		fmt.Println("Sort order specified.")
		err := e.Sort(sortOrder)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("New Order: ")
		for i, s := range e {
			if i+1 == len(e) {
				fmt.Printf("%s\n", s.Name)
			} else {
				fmt.Printf("%s, ", s.Name)
			}
		}
	} else {
		fmt.Println("No sort order specified. Retaining original order of identifiers.")
	}

	return e
}

// NewAdjustedExperiment returns a new experiment where the data points are all adjusted
// by the control of the provided experiment.
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

// Copy returns a duplicate of the given experiment.
func (e Experiment) Copy() Experiment {
	copy := make([]Sample, 0, len(e))

	for _, s := range e {
		copy = append(copy, s)
	}

	return copy
}

// Control returns the sample specified by the name "unpulsed". This sample is used
// to calculate an adjusted experiment.
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

// Identifiers returns the list of identifiers contained in the experiment
func (e Experiment) Identifiers() []string {
	idents := make([]string, 0, len(e))

	for _, s := range e {
		idents = append(idents, s.Name)
	}

	return idents
}

// Sort re-arranges the samples based on a given sort order, specified by a list
// of integers representing an identifiers index
func (e Experiment) Sort(order []int) error {
	if len(e) != len(order) {
		return fmt.Errorf("experiment contains %d identifiers, but sort order contains %d", len(e), len(order))
	}

	original := e.Copy()

	for i, o := range order {
		if o-1 < 0 || o > len(e) {
			return fmt.Errorf("invalid identifier index [%d] - index must be between 1-32", o)
		}

		var tmp Sample

		tmp = original[i]
		e[i] = original[o-1]
		e[o-1] = tmp
	}

	return nil
}

// String "pretty-prints" the experiement
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

// WriteCSV dumps the experiment in CSV format suitable for import into Prism using the
// provided io.Writer
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
