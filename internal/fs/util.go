package fs

import (
	"errors"
	"os"
	"path/filepath"
)

// EnsureParentDir creates the parent directory of the provided file path if needed.
// Returns nil when the directory exists or was created successfully.
func EnsureParentDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return nil
}

// FileExists returns true when the given path exists and is a file.
func FileExists(path string) (bool, error) {
	st, err := os.Stat(path)
	if err == nil {
		if st.IsDir() {
			return false, errors.New("path is directory")
		}
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
