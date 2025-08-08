package convert

import (
	"errors"
	"path/filepath"
	"strings"

	appfs "github.com/romanitalian/osheet2xlsx/v2/internal/fs"
	"github.com/romanitalian/osheet2xlsx/v2/internal/osheet"
	"github.com/romanitalian/osheet2xlsx/v2/internal/xlsx"
)

// ConvertSingle is a placeholder that writes an empty XLSX next to the input file.
func ConvertSingle(inputPath string, outputPath string, overwrite bool) (string, error) {
	out := outputPath
	if out == "" {
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		out = name + ".xlsx"
	}

	if err := appfs.EnsureParentDir(out); err != nil {
		return "", err
	}

	// Path traversal protection when output directory is set by caller
	// Note: outputPath == "" && outputPath != out is always false, so this check is redundant
	// The logic is handled above where we set out = name + ".xlsx" when outputPath is empty

	// batch mode may set outDir upstream; ensure that when caller provides absolute out in opts, it is intended.
	// No additional checks needed here as the logic is handled in the calling code
	// If caller provided an out path under a directory, ensure that when using outDir externally, they validate.
	// Additionally, guard against attempts like name with path separators (should be stripped by Base())
	if strings.ContainsAny(filepath.Base(out), string([]rune{filepath.Separator})) {
		return "", errors.New("invalid output file name")
	}

	if !overwrite {
		if ok, err := appfs.FileExists(out); err != nil {
			return "", err
		} else if ok {
			return "", errors.New("output exists; use --overwrite to replace")
		}
	}

	book, err := osheet.ReadBook(inputPath)
	if err != nil {
		return "", err
	}
	if err := xlsx.WriteBook(book, out); err != nil {
		return "", err
	}
	return out, nil
}
