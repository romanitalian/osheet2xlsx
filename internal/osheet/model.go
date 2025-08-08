package osheet

// Book represents a parsed Osheet workbook (minimal model for MVP).
type Book struct {
	Title  string
	Sheets []Sheet
}

// Sheet represents a single sheet with cell values.
type Sheet struct {
	Name   string
	Width  int
	Height int
	Cells  [][]Cell
	Merges []Merge
	Cols   []ColSpec
	Rows   []RowSpec
}

// Cell represents a single cell value.
type Cell struct {
	StringValue string
	NumberValue float64
	BoolValue   bool
	DateEpoch   float64 // Excel-style serial when Type=ValueDateTime
	// Formula, when non-empty, takes precedence over value fields
	// and will be written as an Excel formula (e.g. "SUM(A1:B2)").
	Formula string
	Type    ValueType
}

// ValueType enumerates supported cell value kinds.
type ValueType int

const (
	ValueEmpty ValueType = iota
	ValueString
	ValueNumber
	ValueBool
	ValueDateTime
)

// Merge represents a merged cell range inclusive.
type Merge struct {
	StartRow int
	StartCol int
	EndRow   int
	EndCol   int
}

// ColSpec describes explicit column width.
type ColSpec struct {
	Index int
	Width float64
}

// RowSpec describes explicit row height.
type RowSpec struct {
	Index  int
	Height float64
}
