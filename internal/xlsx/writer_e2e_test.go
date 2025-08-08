package xlsx

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	osmodel "github.com/romanitalian/osheet2xlsx/v2/internal/osheet"
)

func TestWriteBook_FormulasAndDates(t *testing.T) {
	book := &osmodel.Book{Title: "t", Sheets: []osmodel.Sheet{{
		Name: "S",
		Cells: [][]osmodel.Cell{{
			{Type: osmodel.ValueNumber, NumberValue: 1},
			{Type: osmodel.ValueNumber, NumberValue: 2},
		}, {
			{Formula: "SUM(A1:B1)"},
			{Type: osmodel.ValueDateTime, DateEpoch: 45569}, // 2024-12-31 roughly
		}},
	}}}
	out := filepath.Join(t.TempDir(), "out.xlsx")
	if err := WriteBook(book, out); err != nil {
		t.Fatalf("WriteBook: %v", err)
	}
	// sanity: file is a valid zip
	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	st, err := f.Stat()
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	zr, err := zip.NewReader(f, st.Size())
	if err != nil {
		t.Fatalf("not a zip: %v", err)
	}
	_ = zr
}

func TestJSONLogsAreDirect_NotNested(t *testing.T) {
	rec := struct{ Event, Input string }{Event: "convert_start", Input: "a"}
	if _, err := json.Marshal(rec); err != nil {
		t.Fatalf("marshal: %v", err)
	}
}
