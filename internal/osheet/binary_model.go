package osheet

// BinarySheet represents a parsed binary .osheet file structure
type BinarySheet struct {
	Title  string
	Cells  map[string]map[string]CellData
	Cols   map[string]ColData
	Styles map[string]StyleData
}

// CellData represents a single cell in binary format
type CellData struct {
	Value string `json:"v"`
	Style int    `json:"s,omitempty"`
}

// ColData represents column metadata in binary format
type ColData struct {
	Width float64 `json:"w"`
}

// StyleData represents style information (simplified for MVP)
type StyleData struct {
	// For future expansion - currently not used in conversion
}
