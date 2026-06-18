package donation

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// maxImportRows caps how many data rows a single file may contain. This is a
// guard against oversized / malicious sheets (decompression bombs) and keeps
// memory bounded.
const maxImportRows = 50000

// Decompression guards (zip-bomb defense). A 10MB upload must never expand to
// gigabytes: cap total unzipped size and per-worksheet XML size.
const (
	maxUnzipSize    = 100 << 20 // 100MB total decompressed
	maxUnzipXMLSize = 50 << 20  // 50MB per worksheet XML
)

// maxHeaderScanRows bounds how far down we look for the metadata block and the
// column-header row.
const maxHeaderScanRows = 25

// requiredColumns are the column headers the template MUST contain. Matching is
// case-insensitive and space-insensitive.
var requiredColumns = []string{
	"id donatur",
	"nama donatur",
	"jenis sumbangan",
	"catatan",
	"jumlah",
}

// headerMeta holds the per-file metadata block (komisariat / period).
type headerMeta struct {
	KomisariatID   string
	KomisariatName string
	Period         string
}

// parsedSheet is the full result of reading a donation import file.
type parsedSheet struct {
	Meta headerMeta
	Rows []ImportDonationRequest
}

// parseDonationSheet reads the first worksheet, extracts the metadata block,
// locates and validates the column header, and returns one
// ImportDonationRequest per non-blank data row.
//
// Errors returned here are template/structure errors meant to be surfaced to
// the uploader verbatim.
func parseDonationSheet(r io.Reader) (*parsedSheet, error) {
	// Cap decompression to defend against zip bombs: even a small (<=10MB)
	// upload could otherwise expand to gigabytes in memory.
	f, err := excelize.OpenReader(r, excelize.Options{
		UnzipSizeLimit:    maxUnzipSize,
		UnzipXMLSizeLimit: maxUnzipXMLSize,
	})
	if err != nil {
		return nil, fmt.Errorf("file is not a valid .xlsx workbook: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel file has no worksheet")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read worksheet rows: %w", err)
	}

	meta := extractHeaderMeta(rows)

	headerIdx, colIndex, err := locateHeader(rows)
	if err != nil {
		return nil, err
	}

	out := make([]ImportDonationRequest, 0)
	for i := headerIdx + 1; i < len(rows); i++ {
		row := rows[i]
		req := ImportDonationRequest{
			Row:          i + 1, // 1-based spreadsheet row
			DonaturID:    strings.TrimSpace(valueAt(row, colIndex, "id donatur")),
			DonaturName:  strings.TrimSpace(valueAt(row, colIndex, "nama donatur")),
			CategoryName: strings.TrimSpace(valueAt(row, colIndex, "jenis sumbangan")),
			Note:         strings.TrimSpace(valueAt(row, colIndex, "catatan")),
			Amount:       strings.TrimSpace(valueAt(row, colIndex, "jumlah")),
		}
		if isBlankRow(req) {
			continue
		}
		out = append(out, req)
		if len(out) > maxImportRows {
			return nil, fmt.Errorf("file exceeds the maximum of %d rows", maxImportRows)
		}
	}

	return &parsedSheet{Meta: meta, Rows: out}, nil
}

// locateHeader scans the top rows for the column-header row (the one containing
// "id donatur") and validates that all required columns are present. It returns
// the header row index and a map of normalized column name -> column index.
func locateHeader(rows [][]string) (int, map[string]int, error) {
	limit := min(len(rows), maxHeaderScanRows)
	for i := 0; i < limit; i++ {
		colIndex := make(map[string]int)
		for j, cell := range rows[i] {
			key := normalizeHeader(cell)
			if key == "" {
				continue
			}
			if _, exists := colIndex[key]; !exists {
				colIndex[key] = j
			}
		}
		if _, ok := colIndex["id donatur"]; !ok {
			continue // not the header row
		}

		var missing []string
		for _, col := range requiredColumns {
			if _, ok := colIndex[col]; !ok {
				missing = append(missing, col)
			}
		}
		if len(missing) > 0 {
			return 0, nil, fmt.Errorf(
				"invalid template: missing required column(s): %s. Expected columns: %s",
				strings.Join(missing, ", "), strings.Join(requiredColumns, ", "))
		}
		return i, colIndex, nil
	}

	return 0, nil, fmt.Errorf(
		"invalid template: could not find the header row. Expected columns: %s",
		strings.Join(requiredColumns, ", "))
}

// extractHeaderMeta pulls ID Komisariat / Nama Komisariat / Periode from the
// top rows. It tolerates both "Label : Value" in a single cell and
// "Label | Value" across adjacent cells.
func extractHeaderMeta(rows [][]string) headerMeta {
	var meta headerMeta
	limit := min(len(rows), maxHeaderScanRows)
	for i := 0; i < limit; i++ {
		for j, cell := range rows[i] {
			label, inline := splitLabelValue(cell)
			norm := normalizeHeader(label)
			value := inline
			if value == "" {
				value = nextCellValue(rows[i], j+1)
			}
			if value == "" {
				continue
			}
			switch {
			case norm == "id komisariat" && meta.KomisariatID == "":
				meta.KomisariatID = value
			case norm == "nama komisariat" && meta.KomisariatName == "":
				meta.KomisariatName = value
			case (norm == "periode" || norm == "period") && meta.Period == "":
				meta.Period = value
			}
		}
	}
	return meta
}

// splitLabelValue splits "Label : Value" into ("Label", "Value"). When there is
// no separator it returns (cell, "").
func splitLabelValue(cell string) (label, value string) {
	if idx := strings.Index(cell, ":"); idx >= 0 {
		return cell[:idx], strings.TrimSpace(cell[idx+1:])
	}
	return cell, ""
}

func nextCellValue(row []string, from int) string {
	for k := from; k < len(row); k++ {
		if v := strings.TrimSpace(row[k]); v != "" {
			return v
		}
	}
	return ""
}

// valueAt returns the cell for the given normalized column name, tolerating
// short rows (excelize trims trailing empty cells).
func valueAt(row []string, colIndex map[string]int, col string) string {
	idx, ok := colIndex[col]
	if !ok || idx >= len(row) {
		return ""
	}
	return row[idx]
}

// normalizeHeader lower-cases, trims and collapses internal whitespace so that
// "ID  Donatur", "id donatur" and " ID Donatur " all match.
func normalizeHeader(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func isBlankRow(req ImportDonationRequest) bool {
	return req.DonaturID == "" &&
		req.DonaturName == "" &&
		req.CategoryName == "" &&
		req.Note == "" &&
		req.Amount == ""
}

// parseAmount converts an Indonesian-formatted amount into a numeric value.
// Thousand separators (".") and currency decoration are dropped:
//
//	"20.000"      -> 20000
//	"1.899.000"   -> 1899000
//	"Rp 500.000"  -> 500000
//	""            -> 0
//
// A value that contains no digits at all (e.g. "abc") is rejected.
func parseAmount(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, nil
	}

	var digits strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	if digits.Len() == 0 {
		return 0, fmt.Errorf("invalid amount %q", raw)
	}

	value, err := strconv.ParseFloat(digits.String(), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", raw)
	}
	return value, nil
}

// indonesianMonths maps lower-cased Indonesian month names to their value.
var indonesianMonths = map[string]time.Month{
	"januari": time.January, "februari": time.February, "maret": time.March,
	"april": time.April, "mei": time.May, "juni": time.June,
	"juli": time.July, "agustus": time.August, "september": time.September,
	"oktober": time.October, "november": time.November, "desember": time.December,
}

// parsePeriod interprets the donation period carried in the "Catatan" column.
// It accepts common explicit date layouts and Indonesian "Month Year" text:
//
//	"Desember 2025" -> 2025-12-01
//	"2025-12-01"    -> 2025-12-01
//	"12/2025"       -> 2025-12-01
//	"2025"          -> 2025-01-01
//
// Returns nil when empty or unrecognised (period is optional, never fails a row).
func parsePeriod(raw string) *time.Time {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		"01/2006",
		"2006-01",
		"2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}

	// "Month Year" (Indonesian), e.g. "Desember 2025".
	fields := strings.Fields(strings.ToLower(s))
	if len(fields) == 2 {
		if month, ok := indonesianMonths[fields[0]]; ok {
			if year, err := strconv.Atoi(fields[1]); err == nil {
				t := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
				return &t
			}
		}
	}

	return nil
}
