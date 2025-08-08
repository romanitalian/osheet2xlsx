package osheet

import (
	"archive/zip"
)

// IsLikelyOsheet performs a lightweight detection whether the file is a ZIP container.
func IsLikelyOsheet(path string) bool {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	defer rc.Close()
	return len(rc.File) > 0
}
