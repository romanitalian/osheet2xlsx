package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"

	appcfg "github.com/romanitalian/osheet2xlsx/v2/internal/config"
	appconvert "github.com/romanitalian/osheet2xlsx/v2/internal/convert"
	appfs "github.com/romanitalian/osheet2xlsx/v2/internal/fs"
	applog "github.com/romanitalian/osheet2xlsx/v2/internal/log"
)

type convertOptions struct {
	inputPath string
	out       string
	outDir    string
	recursive bool
	pattern   string
	overwrite bool
	parallel  int
	dryRun    bool
	progress  bool
	failFast  bool
}

func newConvertCmd() *cobra.Command {
	opts := &convertOptions{}
	cmd := &cobra.Command{
		Use:   "convert [path]",
		Short: "Convert .osheet files to .xlsx",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// hydrate defaults from config when flags not set
			cfg, err := appcfg.Load()
			if err != nil {
				// Use default config if loading fails
				cfg = &appcfg.Config{}
			}
			if !cmd.Flags().Changed("pattern") && cfg.Convert.Pattern != "" {
				opts.pattern = cfg.Convert.Pattern
			}
			if !cmd.Flags().Changed("recursive") && cfg.Convert.Recursive {
				opts.recursive = true
			}
			if !cmd.Flags().Changed("out-dir") && cfg.Convert.OutDir != "" {
				opts.outDir = cfg.Convert.OutDir
			}
			if !cmd.Flags().Changed("overwrite") && cfg.Convert.Overwrite {
				opts.overwrite = true
			}
			if !cmd.Flags().Changed("parallel") && cfg.Convert.Parallel != 0 {
				opts.parallel = cfg.Convert.Parallel
			}
			if !cmd.Flags().Changed("dry-run") && cfg.Convert.DryRun {
				opts.dryRun = true
			}
			if !cmd.Flags().Changed("progress") && cfg.Convert.Progress {
				opts.progress = true
			}
			if !cmd.Flags().Changed("fail-fast") && cfg.Convert.FailFast {
				opts.failFast = true
			}
			if len(args) == 1 {
				opts.inputPath = args[0]
			}
			logger := applog.Get()
			logger.Info(fmt.Sprintf("convert: input=%q out=%q outDir=%q pattern=%q recursive=%t overwrite=%t parallel=%d dryRun=%t progress=%t failFast=%t",
				opts.inputPath, opts.out, opts.outDir, opts.pattern, opts.recursive, opts.overwrite, opts.parallel, opts.dryRun, opts.progress, opts.failFast,
			))

			// Decide single vs batch
			var inputs []string
			if opts.inputPath != "" {
				st, err := os.Stat(opts.inputPath)
				if err == nil && !st.IsDir() {
					inputs = []string{opts.inputPath}
				}
			}
			if len(inputs) == 0 {
				root := opts.inputPath
				if root == "" {
					root = "."
				}
				found, err := appfs.ListInputs(root, opts.pattern, opts.recursive)
				if err != nil {
					return err
				}
				inputs = found
			}

			if len(inputs) == 0 {
				return fmt.Errorf("no inputs found")
			}

			var hadErrors bool
			var errMu sync.Mutex
			workerCount := opts.parallel
			if workerCount <= 0 {
				workerCount = 1
			}

			type job struct{ in string }
			jobs := make(chan job)
			var wg sync.WaitGroup

			runOne := func(in string) {
				outPath := opts.out
				if outPath == "" {
					base := filepath.Base(in)
					ext := filepath.Ext(base)
					name := base[:len(base)-len(ext)]
					outName := name + ".xlsx"
					if opts.outDir != "" {
						// Protect against path traversal: ensure the final path stays within outDir
						cleanDir := filepath.Clean(opts.outDir)
						candidate := filepath.Join(cleanDir, outName)
						if !isWithinDir(candidate, cleanDir) {
							logger.Error("invalid output path")
							return
						}
						outPath = candidate
					} else {
						outPath = outName
					}
				}
				if opts.dryRun {
					fmt.Fprintf(getOutputWriter(), "DRY-RUN: would convert %s -> %s\n", in, outPath)
					return
				}
				if jsonLog {
					fmt.Fprintf(getOutputWriter(), `{"event":"convert_start","input":"%s","output":"%s"}`+"\n", in, outPath)
				}
				produced, err := appconvert.ConvertSingle(in, outPath, opts.overwrite)
				if err != nil {
					errMu.Lock()
					hadErrors = true
					errMu.Unlock()
					if jsonLog {
						fmt.Fprintf(getOutputWriter(), `{"event":"convert_error","input":"%s","error":"%v"}`+"\n", in, err)
					} else {
						logger.Error(fmt.Sprintf("convert failed for %s: %v", in, err))
					}
					return
				}
				if jsonLog {
					fmt.Fprintf(getOutputWriter(), `{"event":"convert_ok","input":"%s","output":"%s"}`+"\n", in, produced)
				} else {
					fmt.Fprintf(getOutputWriter(), "OK: %s -> %s\n", in, produced)
				}
			}

			// progress bar (simple): only when TTY and not quiet/json
			showProgress := opts.progress && isTerminal() && !jsonLog && !quiet
			var total int
			var totalDone int
			var start time.Time
			if showProgress {
				total = len(inputs)
				start = time.Now()
				fmt.Fprint(getOutputWriter(), "Progress: 0/", total, " (0%) | 0.0 it/s | elapsed 0s | eta --\r")
			}

			incr := func() {
				if showProgress {
					// very simple progress rendering
					errMu.Lock()
					defer errMu.Unlock()
					totalDone++
					elapsed := time.Since(start)
					secs := elapsed.Seconds()
					speed := 0.0
					if secs > 0 {
						speed = float64(totalDone) / secs
					}
					percent := 0.0
					if total > 0 {
						percent = (float64(totalDone) / float64(total)) * 100.0
					}
					etaStr := "--"
					if speed > 0 && totalDone < total {
						remain := float64(total-totalDone) / speed
						etaStr = formatDuration(time.Duration(remain * float64(time.Second)))
					}
					fmt.Fprintf(getOutputWriter(), "Progress: %d/%d (%.1f%%) | %.1f it/s | elapsed %s | eta %s\r",
						totalDone, total, percent, speed, formatDuration(elapsed), etaStr,
					)
				}
			}

			if workerCount == 1 {
				for _, in := range inputs {
					if opts.failFast && hadErrors {
						break
					}
					runOne(in)
					incr()
				}
			} else {
				for i := 0; i < workerCount; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for j := range jobs {
							errMu.Lock()
							failed := hadErrors
							errMu.Unlock()
							if opts.failFast && failed {
								return
							}
							runOne(j.in)
							incr()
						}
					}()
				}
				for _, in := range inputs {
					jobs <- job{in: in}
				}
				close(jobs)
				wg.Wait()
			}

			if hadErrors {
				// distinct error to be mapped by main or caller to exit code 5
				return fmt.Errorf("partial failure")
			}
			if showProgress {
				fmt.Fprint(getOutputWriter(), "\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.out, "out", "", "output .xlsx file path (single input)")
	cmd.Flags().StringVar(&opts.outDir, "out-dir", "", "output directory (batch)")
	cmd.Flags().BoolVar(&opts.recursive, "recursive", false, "scan directories recursively")
	cmd.Flags().StringVar(&opts.pattern, "pattern", "*.osheet", "glob pattern for inputs")
	cmd.Flags().BoolVar(&opts.overwrite, "overwrite", false, "overwrite existing output files")
	cmd.Flags().IntVar(&opts.parallel, "parallel", 0, "parallel workers (0=auto)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "do not write files, only report")
	cmd.Flags().BoolVar(&opts.progress, "progress", false, "show progress bar for TTY")
	cmd.Flags().BoolVar(&opts.failFast, "fail-fast", false, "stop batch on first error")

	return cmd
}

// formatDuration prints durations as H:MM:SS or M:SS or S
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	secs := int(d.Seconds() + 0.5) // round
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%d:%02d", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
