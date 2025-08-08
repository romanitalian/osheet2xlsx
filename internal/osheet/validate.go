package osheet

import (
	"archive/zip"
	"path"
)

// ValidationIssue represents a structural problem discovered during validation.
type ValidationIssue struct {
	Code    string // stable machine code, e.g., "not_zip", "doc_json_invalid"
	Message string // human-readable summary
}

// ValidateStructure inspects the ZIP and reports structural issues (legacy API).
// Empty slice means acceptable structure for MVP.
func ValidateStructure(zipPath string) ([]string, error) {
	issues, err := ValidateStructureDetailed(zipPath)
	if err != nil {
		return nil, err
	}
	var out []string
	for i := 0; i < len(issues); i++ {
		out = append(out, issues[i].Message)
	}
	return out, nil
}

// ValidateStructureDetailed provides machine-readable issues with codes.
func ValidateStructureDetailed(zipPath string) ([]ValidationIssue, error) {
	var issues []ValidationIssue
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		issues = append(issues, ValidationIssue{Code: "not_zip", Message: "cannot open as zip"})
		return issues, nil
	}
	defer zr.Close()

	// document.json present and parseable?
	if shs, ok := tryParseDocumentJSON(zr.File); ok && len(shs) > 0 {
		return issues, nil
	} else {
		// document.json present but invalid?
		var hasDoc bool
		for i := 0; i < len(zr.File); i++ {
			if path.Base(zr.File[i].Name) == "document.json" {
				hasDoc = true
				break
			}
		}
		if hasDoc {
			issues = append(issues, ValidationIssue{Code: "doc_json_invalid", Message: "document.json present but invalid or empty"})
			return issues, nil
		}
	}

	// Any sheets/*.json parseable?
	var anySheet bool
	for i := 0; i < len(zr.File); i++ {
		f := zr.File[i]
		if !isRegularFile(f) {
			continue
		}
		if path.Dir(f.Name) == "sheets" && path.Ext(f.Name) == ".json" {
			anySheet = true
			if _, ok := tryParseSheetJSON(f); ok {
				return issues, nil
			}
		}
	}
	if anySheet {
		issues = append(issues, ValidationIssue{Code: "sheets_json_invalid", Message: "sheets/*.json present but not parseable"})
		return issues, nil
	}

	// Else: accept archive with files under sheets/ (text fallback)
	for i := 0; i < len(zr.File); i++ {
		f := zr.File[i]
		if !isRegularFile(f) {
			continue
		}
		if path.Dir(f.Name) == "sheets" {
			return issues, nil
		}
	}

	issues = append(issues, ValidationIssue{Code: "no_sheets", Message: "no sheets or document.json found"})
	return issues, nil
}
