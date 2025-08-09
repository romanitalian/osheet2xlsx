package osheet

import (
	"archive/zip"
	"io"
	"os"
	"regexp"
	"strings"
)

// Format represents the detected format of an .osheet file
type Format int

const (
	FormatUnknown Format = iota
	FormatZIP
	FormatBinary
)

// String returns the string representation of the format
func (f Format) String() string {
	switch f {
	case FormatZIP:
		return "ZIP"
	case FormatBinary:
		return "Binary"
	default:
		return "Unknown"
	}
}

// DetectFormat automatically detects the format of an .osheet file
func DetectFormat(path string) (Format, error) {
	file, err := os.Open(path)
	if err != nil {
		return FormatUnknown, err
	}
	defer file.Close()

	// First, try to detect ZIP format (fast check)
	if isZIPFormat(file) {
		return FormatZIP, nil
	}

	// Reset file position for binary format check
	if _, err := file.Seek(0, 0); err != nil {
		return FormatUnknown, err
	}

	// Check for binary format
	if isBinaryFormat(file) {
		return FormatBinary, nil
	}

	return FormatUnknown, nil
}

// isZIPFormat checks if the file is a valid ZIP archive
func isZIPFormat(file *os.File) bool {
	// Check ZIP magic number
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return false
	}

	// ZIP files start with PK\x03\x04
	if header[0] != 0x50 || header[1] != 0x4B || header[2] != 0x03 || header[3] != 0x04 {
		return false
	}

	// Try to open as ZIP to verify it's valid
	if _, err := file.Seek(0, 0); err != nil {
		return false
	}

	rc, err := zip.OpenReader(file.Name())
	if err != nil {
		return false
	}
	defer rc.Close()

	// Check if it has any files (basic validation)
	return len(rc.File) > 0
}

// isBinaryFormat checks if the file is a binary .osheet format
func isBinaryFormat(file *os.File) bool {
	// Read entire file to check for binary format patterns
	data, err := io.ReadAll(file)
	if err != nil {
		return false
	}

	// Convert to string for pattern matching
	text := string(data)

	// Check for binary header pattern (schema, enc, id, ver)
	if !strings.Contains(text, "schema") {
		return false
	}

	// Look for JSON structure with sheets and cells
	jsonPattern := regexp.MustCompile(`"sheets":\s*\{[^}]*\}`)
	if !jsonPattern.MatchString(text) {
		return false
	}

	cellsPattern := regexp.MustCompile(`"cells":\s*\{[^}]*\}`)
	return cellsPattern.MatchString(text)
}
