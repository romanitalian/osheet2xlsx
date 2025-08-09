package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	appcfg "github.com/romanitalian/osheet2xlsx/v3/internal/config"
	appconvert "github.com/romanitalian/osheet2xlsx/v3/internal/convert"
	applog "github.com/romanitalian/osheet2xlsx/v3/internal/log"
)

var (
	logLevel string
	quiet    bool
	noColor  bool
	jsonLog  bool
)

var rootCmd = &cobra.Command{
	Use:           "osheet2xlsx [file.osheet] [flags]",
	Short:         "Convert Osheet files to .xlsx",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no args or first arg is not a .osheet file, show help
		if len(args) == 0 {
			return cmd.Help()
		}

		// Check if first arg ends with .osheet
		if !strings.HasSuffix(args[0], ".osheet") {
			return fmt.Errorf("expected .osheet file, got: %s", args[0])
		}

		// Run conversion logic
		return runConvert(cmd, args[0])
	},
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

	// Add conversion flags for direct file input
	rootCmd.Flags().String("out", "", "output .xlsx file path")
	rootCmd.Flags().Bool("overwrite", false, "overwrite existing output files")
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

// runConvert handles the conversion logic for direct file input
func runConvert(cmd *cobra.Command, inputPath string) error {
	// Get flags from root command
	outFlag, err := cmd.Flags().GetString("out")
	if err != nil {
		return fmt.Errorf("failed to get out flag: %w", err)
	}
	overwriteFlag, err := cmd.Flags().GetBool("overwrite")
	if err != nil {
		return fmt.Errorf("failed to get overwrite flag: %w", err)
	}

	// Create default options for single file conversion
	opts := &convertOptions{
		inputPath: inputPath,
		pattern:   "*.osheet",
		parallel:  1,
		out:       outFlag,
		overwrite: overwriteFlag,
	}

	// Load config for defaults
	cfg, err := appcfg.Load()
	if err != nil {
		cfg = &appcfg.Config{}
	}

	// Apply config defaults (only if flags not set)
	if !overwriteFlag && cfg.Convert.Overwrite {
		opts.overwrite = true
	}

	logger := applog.Get()
	logger.Info(fmt.Sprintf("convert: input=%q out=%q outDir=%q pattern=%q recursive=%t overwrite=%t parallel=%d dryRun=%t progress=%t failFast=%t",
		opts.inputPath, opts.out, opts.outDir, opts.pattern, opts.recursive, opts.overwrite, opts.parallel, opts.dryRun, opts.progress, opts.failFast,
	))

	// Single file conversion

	// Generate output path
	outPath := opts.out
	if outPath == "" {
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		outPath = name + ".xlsx"
	}

	if jsonLog {
		fmt.Fprintf(getOutputWriter(), `{"event":"convert_start","input":"%s","output":"%s"}`+"\n", inputPath, outPath)
	}

	produced, err := appconvert.ConvertSingle(inputPath, outPath, opts.overwrite)
	if err != nil {
		if jsonLog {
			fmt.Fprintf(getOutputWriter(), `{"event":"convert_error","input":"%s","error":"%v"}`+"\n", inputPath, err)
		} else {
			logger.Error(fmt.Sprintf("convert failed for %s: %v", inputPath, err))
		}
		return err
	}

	if jsonLog {
		fmt.Fprintf(getOutputWriter(), `{"event":"convert_ok","input":"%s","output":"%s"}`+"\n", inputPath, produced)
	} else {
		fmt.Fprintf(getOutputWriter(), "OK: %s -> %s\n", inputPath, produced)
	}

	return nil
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
