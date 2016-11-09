package csvtoprism

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

func round(num float64) int {
	if num < 0.0 {
		return int(num - 0.5)
	}

	return int(num + 0.5)
}

// A RawSample is a set of data points associated with some identifier
type RawSample struct {
	Name string `json:"name"`
	Data [3]int `json:"data"`
}

// ParseRawSamplesXLS constructs a set of RawSamples from the given filename.
// It is epxected that the stream provided by the io.ReaderAt is in the format of an XLS file.
func ParseRawSamplesXLS(r io.ReaderAt, size int64) ([]RawSample, error) {
	var buf bytes.Buffer

	err := generateCSVFromXLSXReader(r, size, 0, &buf)
	if err != nil {
		return nil, err
	}

	return ParseRawSamplesCSV(&buf)
}

// ParseRawSamplesCSV constructs a set of RawSamples using data provided by the given io.Reader.
// It is expected that the stream provided by the io.Reader is in the format of a CSV file.
func ParseRawSamplesCSV(r io.Reader) ([]RawSample, error) {
	data, identifiers, err := parseInputData(r)
	if err != nil {
		return nil, err
	}

	return parseRawSamplesFromRawData(data, identifiers)
}

// Adjust returns a new RawSample with the data points adjusted by the
// provided mean
func (r RawSample) Adjust(mean int) RawSample {
	var adjusted RawSample

	adjusted.Name = r.Name

	for i := 0; i < 3; i++ {
		adjusted.Data[i] = r.Data[i] - mean
	}

	return adjusted
}

// A ControlledSample is a set of RawSamples associated with a particular RawSample
// that represents the control group.
type ControlledSample struct {
	Control      RawSample   `json:"control"`
	Experimental []RawSample `json:"experimental"`
}

// Adjust returns a new ControlledSample where all of the data points
// are adjusted by the mean of the Control's data points.
func (c ControlledSample) Adjust() ControlledSample {
	var adjusted ControlledSample

	fmean := float64(c.Control.Data[0]+c.Control.Data[1]+c.Control.Data[2]) / 3.0

	mean := round(fmean)

	adjusted.Control = c.Control.Adjust(mean)

	for _, e := range c.Experimental {
		adjusted.Experimental = append(adjusted.Experimental, e.Adjust(mean))
	}

	return adjusted
}

// An Experiment consists of a name, generally representing the donor source,
// and a set of ControlledSamples.
type Experiment struct {
	Name    string             `json:"name"`
	Samples []ControlledSample `json:"samples"`
}

// Adjusted produces a new Experiment where all of the samples have had their data
// points adjusted by the mean of their control.
func (e Experiment) Adjusted() Experiment {
	var adjusted Experiment

	adjusted.Name = fmt.Sprintf("%s_Adjusted", e.Name)

	for _, s := range e.Samples {
		adjusted.Samples = append(adjusted.Samples, s.Adjust())
	}

	return adjusted
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
	for _, sample := range e.Samples {
		// First write the control
		outputFormat[0][col] = sample.Control.Name
		outputFormat[1][col] = strconv.Itoa(sample.Control.Data[0])
		outputFormat[2][col] = strconv.Itoa(sample.Control.Data[1])
		outputFormat[3][col] = strconv.Itoa(sample.Control.Data[2])
		col++

		// Now write the data poitns
		for _, s := range sample.Experimental {
			outputFormat[0][col] = s.Name
			outputFormat[1][col] = strconv.Itoa(s.Data[0])
			outputFormat[2][col] = strconv.Itoa(s.Data[1])
			outputFormat[3][col] = strconv.Itoa(s.Data[2])
			col++
		}
	}

	cw := csv.NewWriter(w)

	err := cw.WriteAll(outputFormat)
	if err != nil {
		return err
	}

	cw.Flush()

	return nil
}

// ParseRawSamplesFromRawData takes flat slice representations of the data and
// sample identifiers and joins them together.
func parseRawSamplesFromRawData(data []int, identifiers []string) ([]RawSample, error) {
	if len(data)%3 != 0 {
		return nil, errors.New("must have a multiple of 3 data points")
	}

	samples := make([]RawSample, 0)

	dCount := 0

	for _, identifier := range identifiers {
		var s RawSample
		s.Name = identifier
		for j := 0; j < 3; j++ {
			s.Data[j] = data[dCount]
			dCount++
		}
		samples = append(samples, s)
	}

	return samples, nil
}

func generateCSVFromXLSXReader(r io.ReaderAt, size int64, sheetIndex int, w io.Writer) error {
	xlFile, err := xlsx.OpenReaderAt(r, size)
	if err != nil {
		return err
	}

	return generateCSV(xlFile, sheetIndex, w)
}

func generateCSVFromXLSXFile(excelFileName string, sheetIndex int, w io.Writer) error {
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		return err
	}

	return generateCSV(xlFile, sheetIndex, w)
}

func generateCSV(f *xlsx.File, sheetIndex int, w io.Writer) error {
	sheetLen := len(f.Sheets)
	switch {
	case sheetLen == 0:
		return errors.New("This XLSX file contains no sheets.")
	case sheetIndex >= sheetLen:
		return fmt.Errorf("No sheet %d available, please select a sheet between 0 and %d\n", sheetIndex, sheetLen-1)
	}
	sheet := f.Sheets[sheetIndex]
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
