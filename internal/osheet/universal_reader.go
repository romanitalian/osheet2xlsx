package osheet

import (
	"fmt"
)

// ReadBookUniversal automatically detects the format and reads the book
func ReadBookUniversal(path string) (*Book, error) {
	format, err := DetectFormat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	switch format {
	case FormatZIP:
		return ReadBook(path)
	case FormatBinary:
		return ReadBinaryBook(path)
	case FormatUnknown:
		return nil, fmt.Errorf("unsupported or unknown format")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ReadBinaryBook reads a binary .osheet file and returns a Book
func ReadBinaryBook(path string) (*Book, error) {
	binarySheet, err := ParseBinaryOsheet(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse binary .osheet: %w", err)
	}

	sheet, err := ConvertBinaryToSheet(binarySheet)
	if err != nil {
		return nil, fmt.Errorf("failed to convert binary sheet: %w", err)
	}

	return &Book{
		Title:  path,
		Sheets: []Sheet{*sheet},
	}, nil
}
