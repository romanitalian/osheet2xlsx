package osheet

import (
	"os"
	"testing"
)

func TestReadBookUniversal_ZIP(t *testing.T) {
	// Test with a ZIP .osheet file
	testFile := "examples/sample.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	book, err := ReadBookUniversal(testFile)
	if err != nil {
		t.Fatalf("ReadBookUniversal failed: %v", err)
	}

	if book == nil {
		t.Fatal("ReadBookUniversal returned nil")
	}

	if len(book.Sheets) == 0 {
		t.Error("Book should have at least one sheet")
	}

	t.Logf("Successfully read ZIP book: %s with %d sheets", book.Title, len(book.Sheets))
}

func TestReadBookUniversal_Binary(t *testing.T) {
	// Test with a binary .osheet file
	testFile := "osheet-converter/Cursor-step1.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	book, err := ReadBookUniversal(testFile)
	if err != nil {
		t.Fatalf("ReadBookUniversal failed: %v", err)
	}

	if book == nil {
		t.Fatal("ReadBookUniversal returned nil")
	}

	if len(book.Sheets) == 0 {
		t.Error("Book should have at least one sheet")
	}

	sheet := book.Sheets[0]
	if sheet.Name == "" {
		t.Error("Sheet should have a name")
	}

	if sheet.Width <= 0 || sheet.Height <= 0 {
		t.Error("Sheet should have positive dimensions")
	}

	t.Logf("Successfully read binary book: %s with sheet %s (%dx%d)",
		book.Title, sheet.Name, sheet.Width, sheet.Height)
}

func TestReadBookUniversal_Unknown(t *testing.T) {
	// Test with a non-.osheet file
	tempFile := "nonexistent.osheet"

	_, err := ReadBookUniversal(tempFile)
	if err == nil {
		t.Error("ReadBookUniversal should fail for non-existent file")
	}
}

func TestReadBinaryBook(t *testing.T) {
	// Test with a binary .osheet file
	testFile := "osheet-converter/Cursor-step1.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	book, err := ReadBinaryBook(testFile)
	if err != nil {
		t.Fatalf("ReadBinaryBook failed: %v", err)
	}

	if book == nil {
		t.Fatal("ReadBinaryBook returned nil")
	}

	if len(book.Sheets) != 1 {
		t.Errorf("Expected 1 sheet, got %d", len(book.Sheets))
	}

	sheet := book.Sheets[0]
	if sheet.Name != "Таблица1" {
		t.Errorf("Expected sheet name 'Таблица1', got '%s'", sheet.Name)
	}

	t.Logf("Successfully read binary book: %s with sheet %s (%dx%d)",
		book.Title, sheet.Name, sheet.Width, sheet.Height)
}
