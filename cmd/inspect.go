package cmd

import (
	"archive/zip"
	"fmt"

	"github.com/spf13/cobra"

	applog "github.com/romanitalian/osheet2xlsx/v2/internal/log"
	"github.com/romanitalian/osheet2xlsx/v2/internal/osheet"
)

func newInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <path>",
		Short: "Inspect .osheet metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			applog.Get().Info("inspect: " + path)
			if !osheet.IsLikelyOsheet(path) {
				return fmt.Errorf("not an osheet (zip) or unsupported: %s", path)
			}
			// Prefer real parse for sheet names
			if b, err := osheet.ReadBook(path); err == nil && len(b.Sheets) > 0 {
				if jsonLog {
					// print structured info to stdout
					fmt.Fprintf(getOutputWriter(), "{\"event\":\"inspect\",\"sheets\":%d,\"names\":[", len(b.Sheets))
					for i := 0; i < len(b.Sheets); i++ {
						if i > 0 {
							fmt.Fprint(getOutputWriter(), ",")
						}
						fmt.Fprintf(getOutputWriter(), "\"%s\"", b.Sheets[i].Name)
					}
					fmt.Fprintln(getOutputWriter(), "]}")
				} else {
					fmt.Fprintf(getOutputWriter(), "sheets: %d\n", len(b.Sheets))
					for i := 0; i < len(b.Sheets); i++ {
						fmt.Fprintf(getOutputWriter(), "- %s\n", b.Sheets[i].Name)
					}
				}
				return nil
			}
			// Fallback: open zip and count entries
			zr, err := zip.OpenReader(path)
			if err != nil {
				return err
			}
			defer zr.Close()
			if jsonLog {
				fmt.Fprintf(getOutputWriter(), "{\"event\":\"inspect\",\"entries\":%d}\n", len(zr.File))
			} else {
				fmt.Fprintf(getOutputWriter(), "osheet: entries=%d\n", len(zr.File))
			}
			return nil
		},
	}
	return cmd
}
