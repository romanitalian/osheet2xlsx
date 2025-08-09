package osheet

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ConvertBinaryToSheet converts a BinarySheet to our standard Sheet format
func ConvertBinaryToSheet(binary *BinarySheet) (*Sheet, error) {
	if binary == nil {
		return nil, fmt.Errorf("binary sheet is nil")
	}

	// Find maximum row and column indices
	maxRow := 0
	maxCol := 0

	for rowKey, rowData := range binary.Cells {
		rowIndex, err := strconv.Atoi(rowKey)
		if err != nil {
			continue
		}
		if rowIndex > maxRow {
			maxRow = rowIndex
		}

		for colKey := range rowData {
			colIndex, err := strconv.Atoi(colKey)
			if err != nil {
				continue
			}
			if colIndex > maxCol {
				maxCol = colIndex
			}
		}
	}

	// Create matrix with proper dimensions (convert to 1-based)
	height := maxRow + 1
	width := maxCol + 1

	cells := make([][]Cell, height)
	for i := range cells {
		cells[i] = make([]Cell, width)
	}

	// Fill cells from binary format
	for rowKey, rowData := range binary.Cells {
		rowIndex, err := strconv.Atoi(rowKey)
		if err != nil {
			continue
		}

		for colKey, cellData := range rowData {
			colIndex, err := strconv.Atoi(colKey)
			if err != nil {
				continue
			}

			// Convert cell data to our format
			cell := inferCell(cellData.Value)
			cells[rowIndex][colIndex] = cell
		}
	}

	// Convert column specifications
	cols := make([]ColSpec, 0, len(binary.Cols))
	for colKey, colData := range binary.Cols {
		colIndex, err := strconv.Atoi(colKey)
		if err != nil {
			continue
		}
		cols = append(cols, ColSpec{
			Index: colIndex,
			Width: colData.Width,
		})
	}

	return &Sheet{
		Name:   binary.Title,
		Width:  width,
		Height: height,
		Cells:  cells,
		Cols:   cols,
		Merges: nil, // Not supported in binary format for MVP
		Rows:   nil, // Not supported in binary format for MVP
	}, nil
}

// GenerateDocumentJSON creates a document.json in our expected format
func GenerateDocumentJSON(sheet *Sheet) ([]byte, error) {
	// Convert cells to our v3 format with short keys
	cells := make([][]interface{}, len(sheet.Cells))
	for i, row := range sheet.Cells {
		cells[i] = make([]interface{}, len(row))
		for j, cell := range row {
			cellData := make(map[string]interface{})

			// Set value based on cell type
			switch cell.Type {
			case ValueString:
				cellData["t"] = "s"
				cellData["v"] = cell.StringValue
			case ValueNumber:
				cellData["t"] = "n"
				cellData["v"] = cell.NumberValue
			case ValueBool:
				cellData["t"] = "b"
				cellData["v"] = cell.BoolValue
			case ValueDateTime:
				cellData["t"] = "d"
				cellData["v"] = cell.DateEpoch
			default:
				cellData["v"] = cell.StringValue
			}

			// Add formula if present
			if cell.Formula != "" {
				cellData["f"] = cell.Formula
			}

			cells[i][j] = cellData
		}
	}

	// Create document structure
	document := map[string]interface{}{
		"sheets": []interface{}{
			map[string]interface{}{
				"name":  sheet.Name,
				"cells": cells,
			},
		},
	}

	// Add column specifications if present
	if len(sheet.Cols) > 0 {
		cols := make([]map[string]interface{}, 0, len(sheet.Cols))
		for _, col := range sheet.Cols {
			cols = append(cols, map[string]interface{}{
				"index": col.Index,
				"width": col.Width,
			})
		}
		if sheets, ok := document["sheets"].([]interface{}); ok && len(sheets) > 0 {
			if firstSheet, ok := sheets[0].(map[string]interface{}); ok {
				firstSheet["cols"] = cols
			}
		}
	}

	result, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}
	return result, nil
}
