package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	applog "github.com/romanitalian/osheet2xlsx/v2/internal/log"
	"github.com/romanitalian/osheet2xlsx/v2/internal/osheet"
)

func newValidateCmd() *cobra.Command {
	// ErrValidateStructure signals structural validation failure for exit-code mapping
	// exposed from cmd for main's error mapping.
	// Note: keep message generic; details are printed to stdout/stderr.
	ErrValidateStructure = errors.New("validate: structure invalid")
	cmd := &cobra.Command{
		Use:   "validate <path>",
		Short: "Validate .osheet structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			applog.Get().Info("validate: " + path)
			detailed, err := osheet.ValidateStructureDetailed(path)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			if len(detailed) == 0 {
				if jsonLog {
					fmt.Fprintln(getOutputWriter(), "{\"event\":\"validate\",\"ok\":true}")
				} else {
					fmt.Fprintln(getOutputWriter(), "OK: structure looks valid")
				}
				return nil
			}
			if jsonLog {
				// Print first issue for brevity
				fmt.Fprintf(getOutputWriter(), "{\"event\":\"validate\",\"ok\":false,\"code\":\"%s\",\"issue\":\"%s\"}\n", detailed[0].Code, detailed[0].Message)
			}
			return ErrValidateStructure
		},
	}
	return cmd
}
