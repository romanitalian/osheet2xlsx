package osheet

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type sheetV1 struct {
	Name   string                         `json:"name"`
	Rows   [][]string                     `json:"rows"`
	Merges []struct{ SR, SC, ER, EC int } `json:"merges"`
	Cols   []struct {
		Index int
		Width float64
	} `json:"cols"`
	RowHeights []struct {
		Index  int
		Height float64
	} `json:"rowHeights"`
}

type docV1 struct {
	Sheets []sheetV1 `json:"sheets"`
}

func writeZip(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zw.Create: %v", err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zw.Close: %v", err)
	}
}

func TestReadBook_DocumentJSON_V1(t *testing.T) {
	d := t.TempDir()
	zipPath := filepath.Join(d, "sample.osheet")
	doc := docV1{Sheets: []sheetV1{{
		Name: "Test",
		Rows: [][]string{{"1", "2"}, {"3", "=SUM(A1:B1)"}},
	}}}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	writeZip(t, zipPath, map[string][]byte{"document.json": b})

	book, err := ReadBook(zipPath)
	if err != nil {
		t.Fatalf("ReadBook: %v", err)
	}
	if len(book.Sheets) != 1 || book.Sheets[0].Name != "Test" {
		t.Fatalf("sheets parse fail")
	}
	if got := book.Sheets[0].Cells[1][1].StringValue; got != "=SUM(A1:B1)" {
		t.Fatalf("formula cell value kept as string (will be formula only if schema provides): %q", got)
	}
}

func TestReadBook_DocumentJSON_V3_CellsWithFormula(t *testing.T) {
	d := t.TempDir()
	zipPath := filepath.Join(d, "sample.osheet")
	// V3 schema with typed cells and formula
	cells := [][]interface{}{
		{map[string]interface{}{"t": "n", "v": 1}, map[string]interface{}{"t": "n", "v": 2}},
		{map[string]interface{}{"f": "SUM(A1:B1)"}, "02.01.2024"},
	}
	v3 := map[string]interface{}{
		"sheets": []interface{}{map[string]interface{}{
			"name":  "Typed",
			"cells": cells,
		}},
	}
	b, err := json.Marshal(v3)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	writeZip(t, zipPath, map[string][]byte{"document.json": b})

	book, err := ReadBook(zipPath)
	if err != nil {
		t.Fatalf("ReadBook: %v", err)
	}
	if len(book.Sheets) != 1 {
		t.Fatalf("sheets != 1")
	}
	c := book.Sheets[0].Cells
	if c[1][0].Formula == "" {
		t.Fatalf("formula not parsed")
	}
	if c[1][1].Type != ValueDateTime {
		t.Fatalf("date not parsed")
	}
}

func TestValidateStructure_NoSheets(t *testing.T) {
	d := t.TempDir()
	zipPath := filepath.Join(d, "empty.osheet")
	writeZip(t, zipPath, map[string][]byte{"foo.txt": []byte("hi")})
	issues, err := ValidateStructure(zipPath)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(issues) == 0 {
		t.Fatalf("expected issues for no sheets")
	}
}
