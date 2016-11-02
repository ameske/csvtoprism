package csvtoprism

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"strings"
)

var (
	rowMapping = map[string]int{"A": 0, "B": 1, "C": 2, "D": 3, "E": 4, "F": 5, "G": 6, "H": 7}
)

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

type LineType int

const (
	Data LineType = iota
	Identifiers
	Discard
)

func determineLineType(cells []string) LineType {
	if isDataLine(cells) {
		return Data
	}

	if isIdentifierLine(cells) {
		return Identifiers
	}

	// If it isn't one of the following lines we can just ditch it
	return Discard
}

// A data line is going to start off with a capital letter A-H FOLLOWED by a Number
func isDataLine(cells []string) bool {
	foundLetter := false

	for _, c := range cells {
		if len(c) == 0 {
			continue
		}

		// Crude, but once we have found a letter we need to inspect the next character
		if foundLetter {
			_, err := strconv.ParseInt(c, 10, 64)
			if err != nil {
				return false
			}
			return true

		}

		if len(c) == 1 && int(c[0]) >= 65 && int(c[0]) <= 72 {
			foundLetter = true
			continue
		}
	}

	return false
}

// An identifier line is going to start off with a capital A-H FOLLOWED by a String
func isIdentifierLine(cells []string) bool {
	foundLetter := false

	for _, c := range cells {
		if len(c) == 0 {
			continue
		}

		// Even more crude....if we found the letter then the next thing neds to NOT be a number
		if foundLetter {
			_, err := strconv.ParseInt(c, 10, 64)
			if err == nil {
				return false
			}
			return true

		}
		if len(c) == 1 && int(c[0]) >= 65 && int(c[0]) <= 72 {
			foundLetter = true
			continue

		}
	}

	return false

}

// A data row has a capital A-H followed by 12 integers.
func parseDataRow(cells []string) ([]int, int) {
	data := make([]int, 0, 12)

	var values []string

	// Find the identifier
	var identifier string
	for i, c := range cells {
		if isRowIdentifier(c) {
			values = cells[i+1:]
			identifier = c
			break
		}
	}

	// Parse out the values
	for _, v := range values {
		value, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			continue
		}

		data = append(data, int(value))
	}

	// Make sure we have all our data points

	if len(data) != 12 {
		log.Fatal("we didn't get a full row of data, make this graceful soon")
	}

	return data, rowMapping[identifier]
}

// A An identifier row has a capital A-H followed by 4 strings
func parseIdentifierRow(cells []string) ([]string, int) {
	data := make([]string, 0, 4)

	var identifiers []string
	var identifier string

	for i, c := range cells {
		if isRowIdentifier(c) {
			identifiers = cells[i+1:]
			identifier = c
			break
		}
	}

	for _, s := range identifiers {
		if strings.TrimSpace(s) != "" {
			data = append(data, s)
		}
	}

	if len(data) != 4 {
		log.Fatal("We didn't get a full row of identifiers, make this graceful soon")
	}

	return data, rowMapping[identifier]
}

func isRowIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	if int(s[0]) >= 64 && int(s[0]) <= 72 {
		return true
	}

	return false
}
