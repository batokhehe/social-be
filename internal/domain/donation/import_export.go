package donation

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// buildErrorReport renders the failed rows of a batch into an .xlsx file with
// the columns: Row Number, Donator ID, Category, Error. Returns the file bytes
// and a suggested download filename.
func buildErrorReport(log ImportLog, rows []ImportError) ([]byte, string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := f.GetSheetName(0)

	header := []interface{}{"Row Number", "Donator ID", "Category", "Error"}
	if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
		return nil, "", fmt.Errorf("failed to write report header: %w", err)
	}

	for i, row := range rows {
		cellRef, _ := excelize.CoordinatesToCellName(1, i+2)
		record := []interface{}{row.RowNumber, row.DonaturID, row.CategoryName, row.ErrorMessage}
		if err := f.SetSheetRow(sheet, cellRef, &record); err != nil {
			return nil, "", fmt.Errorf("failed to write report row: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", fmt.Errorf("failed to render report: %w", err)
	}

	filename := fmt.Sprintf("import_errors_%s.xlsx", log.BatchID)
	return buf.Bytes(), filename, nil
}
