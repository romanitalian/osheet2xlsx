package fs

import (
	stdfs "io/fs"
	"path/filepath"
)

// ListInputs returns a list of files matching the pattern.
// If recursive is true, it walks subdirectories; otherwise matches only top-level.
func ListInputs(root string, pattern string, recursive bool) ([]string, error) {
	var results []string
	if root == "" {
		root = "."
	}
	if !recursive {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			return nil, err
		}
		return append(results, matches...), nil
	}

	err := filepath.WalkDir(root, func(path string, d stdfs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ok, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}
		if ok {
			results = append(results, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}
