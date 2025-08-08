package cmd

import (
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
func getOutputWriter() *os.File {
	if quiet {
		return os.NewFile(0, os.DevNull)
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

// isWithinDir returns true if file is within baseDir after cleaning paths.
func isWithinDir(file string, baseDir string) bool {
	f := file
	b := baseDir
	if !filepath.IsAbs(f) {
		if abs, err := filepath.Abs(f); err == nil {
			f = abs
		}
	}
	if !filepath.IsAbs(b) {
		if abs, err := filepath.Abs(b); err == nil {
			b = abs
		}
	}
	rel, err := filepath.Rel(b, f)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
