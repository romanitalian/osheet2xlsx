package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	// These values can be overridden by -ldflags at build time.
	version = "dev"
	commit  = ""
	date    = ""
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			buildTime := date
			if buildTime == "" {
				buildTime = time.Now().UTC().Format(time.RFC3339)
			}
			fmt.Fprintf(getOutputWriter(), "version: %s\ncommit: %s\nbuilt: %s\ngo: %s\n", version, commit, buildTime, runtime.Version())
		},
	}
	return cmd
}
