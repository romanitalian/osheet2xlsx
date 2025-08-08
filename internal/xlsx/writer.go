package xlsx

import (
	"fmt"
	"regexp"

	"github.com/xuri/excelize/v2"

	"github.com/romanitalian/osheet2xlsx/v2/internal/osheet"
)

// Helper functions to handle excelize errors gracefully
func safeSetSheetName(f *excelize.File, oldName, newName string) {
	if err := f.SetSheetName(oldName, newName); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to rename sheet %s to %s: %v\n", oldName, newName, err)
	}
}

func safeNewSheet(f *excelize.File, name string) string {
	if sheetIndex, err := f.NewSheet(name); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to create sheet %s: %v\n", name, err)
		return ""
	} else {
		// Convert sheet index to name
		sheetNames := f.GetSheetList()
		if sheetIndex < len(sheetNames) {
			return sheetNames[sheetIndex]
		}
		return name
	}
}

func safeCoordinatesToCellName(col, row int) string {
	if cellName, err := excelize.CoordinatesToCellName(col, row); err != nil {
		// Fallback to simple conversion
		return fmt.Sprintf("%c%d", 'A'+col-1, row)
	} else {
		return cellName
	}
}

func safeSetCellFormula(f *excelize.File, sheet, cell, formula string) {
	if err := f.SetCellFormula(sheet, cell, formula); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to set formula in %s!%s: %v\n", sheet, cell, err)
	}
}

func safeSetCellStr(f *excelize.File, sheet, cell, value string) {
	if err := f.SetCellStr(sheet, cell, value); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to set string in %s!%s: %v\n", sheet, cell, err)
	}
}

func safeSetCellFloat(f *excelize.File, sheet, cell string, value float64, precision int, bitSize int) {
	if err := f.SetCellFloat(sheet, cell, value, precision, bitSize); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to set float in %s!%s: %v\n", sheet, cell, err)
	}
}

func safeSetCellBool(f *excelize.File, sheet, cell string, value bool) {
	if err := f.SetCellBool(sheet, cell, value); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to set bool in %s!%s: %v\n", sheet, cell, err)
	}
}

func safeNewStyle(f *excelize.File, style *excelize.Style) int {
	if styleID, err := f.NewStyle(style); err != nil {
		// Return default style ID on error
		return 0
	} else {
		return styleID
	}
}

func safeSetCellStyle(f *excelize.File, sheet, startCell, endCell string, styleID int) {
	if err := f.SetCellStyle(sheet, startCell, endCell, styleID); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to set style in %s!%s:%s: %v\n", sheet, startCell, endCell, err)
	}
}

func safeMergeCell(f *excelize.File, sheet, startCell, endCell string) {
	if err := f.MergeCell(sheet, startCell, endCell); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to merge cells %s!%s:%s: %v\n", sheet, startCell, endCell, err)
	}
}

func safeSetColWidth(f *excelize.File, sheet, startCol, endCol string, width float64) {
	if err := f.SetColWidth(sheet, startCol, endCol, width); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to set column width in %s!%s:%s: %v\n", sheet, startCol, endCol, err)
	}
}

func safeSetRowHeight(f *excelize.File, sheet string, row int, height float64) {
	if err := f.SetRowHeight(sheet, row, height); err != nil {
		// Log error but continue - this is not critical
		fmt.Printf("Warning: failed to set row height in %s!%d: %v\n", sheet, row, err)
	}
}

// WriteEmptyBook creates a minimal xlsx file at the given path.
func WriteEmptyBook(path string) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	return f.SaveAs(path)
}

// WriteBook writes a parsed Osheet book into an XLSX file.
func WriteBook(book *osheet.Book, outPath string) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Remove default sheet
	defaultSheet := f.GetSheetName(0)
	if defaultSheet == "" {
		defaultSheet = "Sheet1"
	}

	// Create sheets in order
	for i, s := range book.Sheets {
		name := sanitizeSheetName(s.Name)
		if name == "" {
			name = fmt.Sprintf("Sheet%d", i+1)
		}
		if i == 0 {
			// rename default sheet
			safeSetSheetName(f, defaultSheet, name)
		} else {
			safeNewSheet(f, name)
		}
		// Write cells
		for r := 0; r < len(s.Cells); r++ {
			row := s.Cells[r]
			for c := 0; c < len(row); c++ {
				cell := row[c]
				axis := safeCoordinatesToCellName(c+1, r+1)
				// If formula present, prefer writing formula
				if cell.Formula != "" {
					safeSetCellFormula(f, name, axis, cell.Formula)
					continue
				}
				switch cell.Type {
				case osheet.ValueString:
					if len(cell.StringValue) > 0 && cell.StringValue[0] == '=' {
						safeSetCellFormula(f, name, axis, cell.StringValue)
					} else {
						safeSetCellStr(f, name, axis, cell.StringValue)
					}
				case osheet.ValueNumber:
					safeSetCellFloat(f, name, axis, cell.NumberValue, -1, 64)
				case osheet.ValueBool:
					if cell.BoolValue {
						safeSetCellBool(f, name, axis, true)
					} else {
						safeSetCellBool(f, name, axis, false)
					}
				case osheet.ValueDateTime:
					safeSetCellFloat(f, name, axis, cell.DateEpoch, -1, 64)
					// Apply a basic date-time style
					styleID := safeNewStyle(f, &excelize.Style{NumFmt: 22})
					safeSetCellStyle(f, name, axis, axis, styleID)
				default:
					safeSetCellStr(f, name, axis, cell.StringValue)
				}
			}
		}
		// Apply merges
		for _, m := range s.Merges {
			if m.StartRow <= 0 || m.StartCol <= 0 || m.EndRow < m.StartRow || m.EndCol < m.StartCol {
				continue
			}
			ax1 := safeCoordinatesToCellName(m.StartCol, m.StartRow)
			ax2 := safeCoordinatesToCellName(m.EndCol, m.EndRow)
			safeMergeCell(f, name, ax1, ax2)
		}
		// Apply column widths
		for _, c := range s.Cols {
			if c.Index <= 0 || c.Width <= 0 {
				continue
			}
			safeSetColWidth(f, name, columnName(c.Index), columnName(c.Index), c.Width)
		}
		// Apply row heights
		for _, rh := range s.Rows {
			if rh.Index <= 0 || rh.Height <= 0 {
				continue
			}
			safeSetRowHeight(f, name, rh.Index, rh.Height)
		}
	}

	return f.SaveAs(outPath)
}

var invalidSheetChars = regexp.MustCompile(`[\[\]\*\?/\\:]`)

func sanitizeSheetName(in string) string {
	if in == "" {
		return in
	}
	// Replace invalid characters
	cleaned := invalidSheetChars.ReplaceAllString(in, "_")
	// Trim to Excel's 31-char sheet name limit
	if len(cleaned) > 31 {
		cleaned = cleaned[:31]
	}
	return cleaned
}

// columnName converts 1-based column index to Excel column label.
func columnName(idx int) string {
	// Simple conversion without recursion
	s := ""
	for idx > 0 {
		idx--
		s = string([]byte{byte('A' + (idx % 26))}) + s
		idx /= 26
	}
	return s
}
