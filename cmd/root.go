package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	appcfg "github.com/romanitalian/osheet2xlsx/v2/internal/config"
	applog "github.com/romanitalian/osheet2xlsx/v2/internal/log"
)

var (
	logLevel string
	quiet    bool
	noColor  bool
	jsonLog  bool
)

var rootCmd = &cobra.Command{
	Use:           "osheet2xlsx",
	Short:         "Convert Osheet files to .xlsx",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := appcfg.Load()
		if err != nil {
			// Use default config if loading fails
			cfg = &appcfg.Config{}
		}
		// flags override config
		if logLevel == "" && cfg.LogLevel != "" {
			logLevel = cfg.LogLevel
		}
		if !cmd.Flags().Changed("json") && cfg.JSON {
			jsonLog = true
		}
		if !cmd.Flags().Changed("quiet") && cfg.Quiet {
			quiet = true
		}
		if !cmd.Flags().Changed("no-color") && cfg.NoColor {
			noColor = true
		}
		return applog.SetupLogger(logLevel, quiet, noColor, jsonLog)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level: error|warn|info|debug")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress non-error output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&jsonLog, "json", false, "enable JSON logs")
}

// Execute runs the root command.
func Execute() error {
	rootCmd.AddCommand(newConvertCmd())
	rootCmd.AddCommand(newInspectCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newCompletionCmd())
	// Helper used by convert for safe outDir joins
	rootCmd.SilenceUsage = true
	return rootCmd.Execute()
}

// getOutputWriter returns the appropriate writer for standard output respecting quiet mode.
func getOutputWriter() io.Writer {
	if quiet {
		return io.Discard
	}
	return os.Stdout
}

// isTerminal attempts to detect if stdout is a terminal (best-effort, no external deps).
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	mode := fi.Mode()
	return (mode & os.ModeCharDevice) != 0
}

// isWithinDir returns true if file is within baseDir using only path cleaning and Rel.
// Both paths are cleaned; no reliance on current working directory resolution.
func isWithinDir(file string, baseDir string) bool {
	b := filepath.Clean(baseDir)
	f := filepath.Clean(file)
	rel, err := filepath.Rel(b, f)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
