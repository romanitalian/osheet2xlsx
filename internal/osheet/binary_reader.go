package osheet

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// IsBinaryOsheet checks if the file is a binary .osheet format
func IsBinaryOsheet(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read entire file to check for JSON content
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

// ParseBinaryOsheet parses a binary .osheet file and extracts sheet data
func ParseBinaryOsheet(path string) (*BinarySheet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read entire file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Convert to string for JSON extraction
	text := string(data)

	// Find the main JSON object that contains sheets and cells
	// Look for the pattern that starts with {"gcVer":... and contains sheets
	jsonStart := strings.Index(text, `{"gcVer"`)
	if jsonStart == -1 {
		return nil, fmt.Errorf("no main JSON content found")
	}

	// Find the complete JSON object
	jsonContent, err := extractCompleteJSON(text[jsonStart:])
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON: %w", err)
	}

	// Parse the JSON structure
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract sheet information
	sheets, ok := jsonData["sheets"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no sheets found in JSON")
	}

	// Find the sheet data in the text/sh_1 section
	// The actual sheet data is in a separate JSON object
	sheetDataStart := strings.Index(text, `text/sh_1`)
	if sheetDataStart == -1 {
		return nil, fmt.Errorf("no sheet data section found")
	}

	// Find the JSON object after text/sh_1
	sheetJSONStart := strings.Index(text[sheetDataStart:], "{")
	if sheetJSONStart == -1 {
		return nil, fmt.Errorf("no sheet JSON found")
	}

	sheetJSONContent, err := extractCompleteJSON(text[sheetDataStart+sheetJSONStart:])
	if err != nil {
		return nil, fmt.Errorf("failed to extract sheet JSON: %w", err)
	}

	// Parse the sheet JSON
	var sheetJSON map[string]interface{}
	if err := json.Unmarshal([]byte(sheetJSONContent), &sheetJSON); err != nil {
		return nil, fmt.Errorf("failed to parse sheet JSON: %w", err)
	}

	// Extract title from the first sheet info
	title := "Sheet"
	for _, data := range sheets {
		if sheet, ok := data.(map[string]interface{}); ok {
			if titleVal, ok := sheet["title"].(string); ok {
				title = titleVal
				break
			}
		}
	}

	// Extract cells data from sheetJSON
	cells, ok := sheetJSON["cells"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no cells data found")
	}

	// Parse cells structure
	parsedCells := make(map[string]map[string]CellData)
	for rowKey, rowData := range cells {
		if rowMap, ok := rowData.(map[string]interface{}); ok {
			parsedCells[rowKey] = make(map[string]CellData)
			for colKey, cellData := range rowMap {
				if cellMap, ok := cellData.(map[string]interface{}); ok {
					cell := CellData{}
					if val, ok := cellMap["v"].(string); ok {
						cell.Value = val
					}
					if style, ok := cellMap["s"].(float64); ok {
						cell.Style = int(style)
					}
					parsedCells[rowKey][colKey] = cell
				}
			}
		}
	}

	// Extract columns data
	cols := make(map[string]ColData)
	if colsData, ok := sheetJSON["cols"].(map[string]interface{}); ok {
		for colKey, colData := range colsData {
			if colMap, ok := colData.(map[string]interface{}); ok {
				if width, ok := colMap["w"].(float64); ok {
					cols[colKey] = ColData{Width: width}
				}
			}
		}
	}

	return &BinarySheet{
		Title:  title,
		Cells:  parsedCells,
		Cols:   cols,
		Styles: make(map[string]StyleData), // Simplified for MVP
	}, nil
}

// extractCompleteJSON finds the complete JSON object from the given text
func extractCompleteJSON(text string) (string, error) {
	braceCount := 0
	endIndex := -1

	for i, char := range text {
		if char == '{' {
			braceCount++
		} else if char == '}' {
			braceCount--
			if braceCount == 0 {
				endIndex = i + 1
				break
			}
		}
	}

	if endIndex == -1 {
		return "", fmt.Errorf("incomplete JSON object")
	}

	return text[:endIndex], nil
}
