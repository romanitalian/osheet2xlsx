package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds global and command defaults.
type Config struct {
	LogLevel string        `json:"logLevel"`
	JSON     bool          `json:"json"`
	Quiet    bool          `json:"quiet"`
	NoColor  bool          `json:"noColor"`
	Convert  ConvertConfig `json:"convert"`
}

type ConvertConfig struct {
	Pattern   string `json:"pattern"`
	Recursive bool   `json:"recursive"`
	OutDir    string `json:"outDir"`
	Overwrite bool   `json:"overwrite"`
	Parallel  int    `json:"parallel"`
	DryRun    bool   `json:"dryRun"`
	Progress  bool   `json:"progress"`
	FailFast  bool   `json:"failFast"`
}

var loaded *Config

// Load reads configuration from file and environment variables.
// File search order: $OS2X_CONFIG, ./osheet2xlsx.json, $XDG_CONFIG_HOME/osheet2xlsx/config.json, $HOME/.osheet2xlsx.json
func Load() (*Config, error) {
	if loaded != nil {
		return loaded, nil
	}
	cfg := &Config{}
	// from file
	paths := candidatePaths()
	for _, p := range paths {
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			if err := readJSON(p, cfg); err != nil {
				return nil, err
			}
			break
		}
	}
	// env overrides
	if v := os.Getenv("OS2X_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("OS2X_JSON"); v != "" {
		cfg.JSON = parseBool(v)
	}
	if v := os.Getenv("OS2X_QUIET"); v != "" {
		cfg.Quiet = parseBool(v)
	}
	if v := os.Getenv("OS2X_NO_COLOR"); v != "" {
		cfg.NoColor = parseBool(v)
	}

	if v := os.Getenv("OS2X_CONVERT_PATTERN"); v != "" {
		cfg.Convert.Pattern = v
	}
	if v := os.Getenv("OS2X_CONVERT_RECURSIVE"); v != "" {
		cfg.Convert.Recursive = parseBool(v)
	}
	if v := os.Getenv("OS2X_CONVERT_OUT_DIR"); v != "" {
		cfg.Convert.OutDir = v
	}
	if v := os.Getenv("OS2X_CONVERT_OVERWRITE"); v != "" {
		cfg.Convert.Overwrite = parseBool(v)
	}
	if v := os.Getenv("OS2X_CONVERT_PARALLEL"); v != "" {
		cfg.Convert.Parallel = parseInt(v)
	}
	if v := os.Getenv("OS2X_CONVERT_DRY_RUN"); v != "" {
		cfg.Convert.DryRun = parseBool(v)
	}
	if v := os.Getenv("OS2X_CONVERT_PROGRESS"); v != "" {
		cfg.Convert.Progress = parseBool(v)
	}
	if v := os.Getenv("OS2X_CONVERT_FAIL_FAST"); v != "" {
		cfg.Convert.FailFast = parseBool(v)
	}

	loaded = cfg
	return cfg, nil
}

func candidatePaths() []string {
	var out []string
	if v := os.Getenv("OS2X_CONFIG"); v != "" {
		out = append(out, v)
	}
	out = append(out, "osheet2xlsx.json")
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		out = append(out, filepath.Join(x, "osheet2xlsx", "config.json"))
	}
	if h := os.Getenv("HOME"); h != "" {
		out = append(out, filepath.Join(h, ".osheet2xlsx.json"))
	}
	return out
}

func readJSON(path string, into *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return errors.New("empty config file")
	}
	var tmp Config
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	merge(into, &tmp)
	return nil
}

func merge(dst, src *Config) {
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
	dst.JSON = dst.JSON || src.JSON
	dst.Quiet = dst.Quiet || src.Quiet
	dst.NoColor = dst.NoColor || src.NoColor
	if src.Convert.Pattern != "" {
		dst.Convert.Pattern = src.Convert.Pattern
	}
	dst.Convert.Recursive = dst.Convert.Recursive || src.Convert.Recursive
	if src.Convert.OutDir != "" {
		dst.Convert.OutDir = src.Convert.OutDir
	}
	dst.Convert.Overwrite = dst.Convert.Overwrite || src.Convert.Overwrite
	if src.Convert.Parallel != 0 {
		dst.Convert.Parallel = src.Convert.Parallel
	}
	dst.Convert.DryRun = dst.Convert.DryRun || src.Convert.DryRun
	dst.Convert.Progress = dst.Convert.Progress || src.Convert.Progress
	dst.Convert.FailFast = dst.Convert.FailFast || src.Convert.FailFast
}

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func parseInt(s string) int {
	if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return i
	}
	return 0
}
