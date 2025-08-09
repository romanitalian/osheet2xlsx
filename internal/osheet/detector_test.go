package osheet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFormat_ZIP(t *testing.T) {
	// Test with a ZIP .osheet file
	testFile := "examples/sample.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	format, err := DetectFormat(testFile)
	if err != nil {
		t.Fatalf("DetectFormat failed: %v", err)
	}

	if format != FormatZIP {
		t.Errorf("Expected FormatZIP, got %s", format)
	}
}

func TestDetectFormat_Binary(t *testing.T) {
	// Test with a binary .osheet file
	testFile := "osheet-converter/Cursor-step1.osheet"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file %s not found, skipping test", testFile)
	}

	format, err := DetectFormat(testFile)
	if err != nil {
		t.Fatalf("DetectFormat failed: %v", err)
	}

	if format != FormatBinary {
		t.Errorf("Expected FormatBinary, got %s", format)
	}
}

func TestDetectFormat_Unknown(t *testing.T) {
	// Test with a non-.osheet file
	tempFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tempFile, []byte("not an osheet file"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	format, err := DetectFormat(tempFile)
	if err != nil {
		t.Fatalf("DetectFormat failed: %v", err)
	}

	if format != FormatUnknown {
		t.Errorf("Expected FormatUnknown, got %s", format)
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{FormatUnknown, "Unknown"},
		{FormatZIP, "ZIP"},
		{FormatBinary, "Binary"},
	}

	for _, tt := range tests {
		if got := tt.format.String(); got != tt.want {
			t.Errorf("Format.String() = %v, want %v", got, tt.want)
		}
	}
}
