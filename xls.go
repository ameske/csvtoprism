package csvtoprism

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tealeg/xlsx"
)

// Adapted from https://github.com/tealeg/xlsx2csv
func CSVFromXLS(excelFileName string, sheetIndex int, w io.Writer) error {
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
