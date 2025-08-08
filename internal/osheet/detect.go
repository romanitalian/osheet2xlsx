package osheet

import (
	"archive/zip"
	"os"
)

// IsLikelyOsheet performs a lightweight detection whether the file is a ZIP container.
func IsLikelyOsheet(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	r, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return false
	}
	return len(r.File) > 0
}
