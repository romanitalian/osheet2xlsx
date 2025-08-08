package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	appcmd "github.com/romanitalian/osheet2xlsx/v2/cmd"
)

func main() {
	if err := appcmd.Execute(); err != nil {
		code := mapErrorToExitCode(err)
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(code)
	}
}

func mapErrorToExitCode(err error) int {
	if err == nil {
		return 0
	}
	// specific typed errors first
	if errors.Is(err, appcmd.ErrValidateStructure) {
		return 4 // structural issues in input
	}
	// simple mapping heuristic; specific errors could be wrapped/types later
	msg := err.Error()
	if msg == "partial failure" {
		return 5
	}
	if contains(msg, []string{"argument", "usage", "no inputs"}) {
		return 2
	}
	if contains(msg, []string{"unsupported osheet", "parse", "invalid"}) {
		return 4
	}
	if contains(msg, []string{"permission", "exists", "io", "read", "write"}) {
		return 3
	}
	return 1
}

func contains(s string, subs []string) bool {
	for i := 0; i < len(subs); i++ {
		if strings.Contains(s, subs[i]) {
			return true
		}
	}
	return false
}
