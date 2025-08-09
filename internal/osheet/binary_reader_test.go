package osheet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBinaryOsheet(t *testing.T) {
	// Test with a real binary .osheet file
	testFile := "osheet-converter/Cursor-step1.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	if !IsBinaryOsheet(testFile) {
		t.Errorf("IsBinaryOsheet should return true for %s", testFile)
	}

	// Test with a non-binary file
	tempFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tempFile, []byte("not a binary osheet"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if IsBinaryOsheet(tempFile) {
		t.Errorf("IsBinaryOsheet should return false for %s", tempFile)
	}
}

func TestParseBinaryOsheet(t *testing.T) {
	testFile := "osheet-converter/Cursor-step1.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	binarySheet, err := ParseBinaryOsheet(testFile)
	if err != nil {
		t.Fatalf("ParseBinaryOsheet failed: %v", err)
	}

	if binarySheet == nil {
		t.Fatal("ParseBinaryOsheet returned nil")
	}

	// Check basic properties
	if binarySheet.Title == "" {
		t.Error("Title should not be empty")
	}

	if len(binarySheet.Cells) == 0 {
		t.Error("Cells should not be empty")
	}

	// Check that we have some data
	hasData := false
	for _, row := range binarySheet.Cells {
		if len(row) > 0 {
			hasData = true
			break
		}
	}

	if !hasData {
		t.Error("Should have some cell data")
	}

	t.Logf("Parsed sheet: %s with %d rows", binarySheet.Title, len(binarySheet.Cells))
}

func TestConvertBinaryToSheet(t *testing.T) {
	testFile := "osheet-converter/Cursor-step1.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	binarySheet, err := ParseBinaryOsheet(testFile)
	if err != nil {
		t.Fatalf("ParseBinaryOsheet failed: %v", err)
	}

	sheet, err := ConvertBinaryToSheet(binarySheet)
	if err != nil {
		t.Fatalf("ConvertBinaryToSheet failed: %v", err)
	}

	if sheet == nil {
		t.Fatal("ConvertBinaryToSheet returned nil")
	}

	// Check basic properties
	if sheet.Name == "" {
		t.Error("Sheet name should not be empty")
	}

	if sheet.Width <= 0 {
		t.Error("Sheet width should be positive")
	}

	if sheet.Height <= 0 {
		t.Error("Sheet height should be positive")
	}

	if len(sheet.Cells) == 0 {
		t.Error("Sheet cells should not be empty")
	}

	t.Logf("Converted sheet: %s (%dx%d)", sheet.Name, sheet.Width, sheet.Height)
}

func TestGenerateDocumentJSON(t *testing.T) {
	testFile := "osheet-converter/Cursor-step1.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	binarySheet, err := ParseBinaryOsheet(testFile)
	if err != nil {
		t.Fatalf("ParseBinaryOsheet failed: %v", err)
	}

	sheet, err := ConvertBinaryToSheet(binarySheet)
	if err != nil {
		t.Fatalf("ConvertBinaryToSheet failed: %v", err)
	}

	jsonData, err := GenerateDocumentJSON(sheet)
	if err != nil {
		t.Fatalf("GenerateDocumentJSON failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Generated JSON should not be empty")
	}

	// Check that JSON contains expected structure
	jsonStr := string(jsonData)
	if !contains(jsonStr, `"sheets"`) {
		t.Error("Generated JSON should contain 'sheets'")
	}

	if !contains(jsonStr, `"cells"`) {
		t.Error("Generated JSON should contain 'cells'")
	}

	t.Logf("Generated JSON size: %d bytes", len(jsonData))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}
