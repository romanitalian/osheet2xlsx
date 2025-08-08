package cmd

import (
	"archive/zip"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func goRun(args ...string) *exec.Cmd {
	// This test file lives in package cmd, whose directory is <repo>/cmd at runtime.
	// The main package is one level up at <repo>.
	// Note: This is safe as we control the args in tests
	cmd := exec.Command("go", append([]string{"run", ".."}, args...)...)
	return cmd
}

func TestCLI_Validate_ExitCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	// invalid: run validate on non-zip file path
	cmd := goRun("validate", filepath.Join("..", "go.mod"))
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit (out=%s)", string(out))
	}
	// go run returns exit=1 while embedding program exit in output as "exit status N"
	// Accept either ExitError(4) or output containing "exit status 4"
	if ee, ok := err.(*exec.ExitError); ok {
		if ee.ExitCode() == 4 {
			return
		}
	}
	if !strings.Contains(string(out), "exit status 4") {
		t.Fatalf("validate exit mismatch: err=%v out=%s", err, string(out))
	}
}

func TestCLI_Version_Runs(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skip("not supported")
	}
	cmd := goRun("version")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("version failed: %v (%s)", err, string(out))
	}
}

// no-op

func TestCLI_Convert_OK(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	// Prepare temp osheet with document.json v3 including a formula & date
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.osheet")
	out := filepath.Join(tmp, "out.xlsx")
	makeOsheet(t, in)

	cmd := goRun("convert", in, "--out", out, "--overwrite")
	outb, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("convert failed: %v (%s)", err, string(outb))
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output not created: %v", err)
	}
}

func makeOsheet(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	zw := zip.NewWriter(f)
	// build document.json
	cells := [][]interface{}{
		{map[string]interface{}{"t": "n", "v": 1}, map[string]interface{}{"t": "n", "v": 2}},
		{map[string]interface{}{"f": "SUM(A1:B1)"}, "2024-01-02"},
	}
	v3 := map[string]interface{}{
		"sheets": []interface{}{map[string]interface{}{
			"name":  "S",
			"cells": cells,
		}},
	}
	b, err := json.Marshal(v3)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	w, err := zw.Create("document.json")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := w.Write(b); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}
}

func TestCLI_Validate_UnparseableSheetJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	tmp := t.TempDir()
	in := filepath.Join(tmp, "bad.osheet")
	// build zip with sheets/foo.json invalid content
	file, err := os.Create(in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	zipWriter := zip.NewWriter(file)
	writer, err := zipWriter.Create("sheets/foo.json")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, writeErr := writer.Write([]byte("{invalid json")); writeErr != nil {
		t.Fatalf("write: %v", writeErr)
	}
	if closeErr := zipWriter.Close(); closeErr != nil {
		t.Fatalf("close zip: %v", closeErr)
	}
	if fileCloseErr := file.Close(); fileCloseErr != nil {
		t.Fatalf("close file: %v", fileCloseErr)
	}
	// run validate
	cmd := goRun("validate", in)
	outb, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure, got success: %s", string(outb))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if ee.ExitCode() == 4 {
			return
		}
	}
	if !strings.Contains(string(outb), "exit status 4") {
		t.Fatalf("unexpected output: %s", string(outb))
	}
}
