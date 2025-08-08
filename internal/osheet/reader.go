package osheet

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"
)

// ReadBook parses an Osheet ZIP and extracts basic sheet-like data.
// MVP: creates one sheet per file under "sheets/" by embedding up to 32KB of text content into cell A1.
// If no such files found, creates a single sheet "archive" listing entry names.
func ReadBook(zipPath string) (*Book, error) {
	if !IsLikelyOsheet(zipPath) {
		return nil, errors.New("unsupported osheet layout or not a zip")
	}

	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var sheets []Sheet

	if shs, ok := tryParseDocumentJSON(rc.File); ok && len(shs) > 0 {
		sheets = append(sheets, shs...)
	}

	for _, f := range rc.File {
		if !isRegularFile(f) {
			continue
		}
		// Try JSON-based sheets first
		if len(sheets) == 0 && strings.HasPrefix(f.Name, "sheets/") && strings.HasSuffix(strings.ToLower(f.Name), ".json") {
			if sh, ok := tryParseSheetJSON(f); ok {
				sheets = append(sheets, sh)
				continue
			}
		}
	}

	// If no JSON sheets found, fallback: for any file in sheets/ embed text; else list entries
	if len(sheets) == 0 {
		for _, f := range rc.File {
			if !isRegularFile(f) {
				continue
			}
			if strings.HasPrefix(f.Name, "sheets/") {
				content, readErr := readLimitedFile(f, 32*1024)
				if readErr != nil {
					content = "<read error>"
				}
				base := path.Base(f.Name)
				sheet := Sheet{
					Name:   base,
					Width:  1,
					Height: 1,
					Cells:  [][]Cell{{{StringValue: content, Type: ValueString}}},
				}
				sheets = append(sheets, sheet)
			}
		}
	}

	if len(sheets) == 0 {
		// Fallback: list archive entries
		var rows [][]Cell
		for _, f := range rc.File {
			rows = append(rows, []Cell{{StringValue: f.Name, Type: ValueString}})
		}
		sheets = append(sheets, Sheet{
			Name:   "archive",
			Width:  1,
			Height: len(rows),
			Cells:  rows,
		})
	}

	b := &Book{Title: zipPath, Sheets: sheets}
	return b, nil
}

func readLimitedFile(f *zip.File, limit int) (string, error) {
	r, err := f.Open()
	if err != nil {
		return "", err
	}
	defer r.Close()
	var builder strings.Builder
	buf := make([]byte, 4096)
	total := 0
	for {
		if total >= limit {
			builder.WriteString("\n<trimmed>")
			break
		}
		n, rerr := r.Read(buf)
		if n > 0 {
			// ensure not exceed limit
			remain := limit - total
			if n > remain {
				n = remain
			}
			builder.Write(buf[:n])
			total += n
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return builder.String(), rerr
		}
	}
	return builder.String(), nil
}

func isRegularFile(f *zip.File) bool {
	return !strings.HasSuffix(f.Name, "/") && f.FileInfo().Mode().IsRegular()
}

// tryParseSheetJSON attempts to parse a sheet JSON file with simple, predefined shapes.
// Supported shapes:
// 1) {"name":"Sheet1","rows":[["a","b"],["c","d"]]}
// 2) [["a","b"],["c","d"]]
// 3) {"rows":[["a","b"]]}
func tryParseSheetJSON(f *zip.File) (Sheet, bool) {
	r, err := f.Open()
	if err != nil {
		return Sheet{}, false
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return Sheet{}, false
	}

	type rowsNamed struct {
		Name string     `json:"name"`
		Rows [][]string `json:"rows"`
	}
	var rn rowsNamed
	if json.Unmarshal(data, &rn) == nil && len(rn.Rows) > 0 {
		return sheetFromRows(defaultName(rn.Name, path.Base(f.Name)), rn.Rows, nil, nil, nil), true
	}
	var rowsOnly [][]string
	if json.Unmarshal(data, &rowsOnly) == nil && len(rowsOnly) > 0 {
		return sheetFromRows(defaultName("", path.Base(f.Name)), rowsOnly, nil, nil, nil), true
	}
	var r2 struct {
		Rows [][]string `json:"rows"`
	}
	if json.Unmarshal(data, &r2) == nil && len(r2.Rows) > 0 {
		return sheetFromRows(defaultName("", path.Base(f.Name)), r2.Rows, nil, nil, nil), true
	}
	_ = rn // silence unused in some toolchains
	_ = r2
	return Sheet{}, false
}

func defaultName(name string, fallback string) string {
	if name != "" {
		return name
	}
	return strings.TrimSuffix(fallback, ".json")
}

func sheetFromRows(name string, rows [][]string, merges []Merge, cols []ColSpec, rowSpecs []RowSpec) Sheet {
	height := len(rows)
	width := 0
	for i := 0; i < len(rows); i++ {
		if len(rows[i]) > width {
			width = len(rows[i])
		}
	}
	cells := make([][]Cell, height)
	for r := 0; r < height; r++ {
		row := rows[r]
		cells[r] = make([]Cell, len(row))
		for c := 0; c < len(row); c++ {
			cells[r][c] = inferCell(row[c])
		}
	}
	return Sheet{Name: name, Width: width, Height: height, Cells: cells, Merges: merges, Cols: cols, Rows: rowSpecs}
}

// tryParseDocumentJSON parses document.json with an expected shape.
func tryParseDocumentJSON(files []*zip.File) ([]Sheet, bool) {
	var doc *zip.File
	for _, f := range files {
		if strings.EqualFold(path.Base(f.Name), "document.json") {
			doc = f
			break
		}
	}
	if doc == nil {
		return nil, false
	}
	r, err := doc.Open()
	if err != nil {
		return nil, false
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, false
	}

	// Flexible parsing strategy: support multiple sheet schemas
	var docGeneric struct {
		Sheets []json.RawMessage `json:"sheets"`
	}
	if json.Unmarshal(data, &docGeneric) != nil || len(docGeneric.Sheets) == 0 {
		return nil, false
	}
	var out []Sheet
	for i := 0; i < len(docGeneric.Sheets); i++ {
		sh, ok := parseDocumentSheet(docGeneric.Sheets[i])
		if ok {
			out = append(out, sh)
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// parseDocumentSheet tries several schema variants for a single sheet JSON value.
func parseDocumentSheet(raw json.RawMessage) (Sheet, bool) {
	// Base variants of metadata
	type (
		mergeJSON struct{ SR, SC, ER, EC int }
		colJSON   struct {
			Index int
			Width float64
		}
		rowJSON struct {
			Index  int
			Height float64
		}
		sheetV1 struct {
			Name       string      `json:"name"`
			Rows       [][]string  `json:"rows"`
			Merges     []mergeJSON `json:"merges"`
			Cols       []colJSON   `json:"cols"`
			RowHeights []rowJSON   `json:"rowHeights"`
		}
		sheetV2 struct {
			Name       string          `json:"name"`
			Rows       [][]interface{} `json:"rows"`
			Merges     interface{}     `json:"merges"`
			Cols       []colJSON       `json:"cols"`
			RowHeights []rowJSON       `json:"rowHeights"`
		}
		sheetV3 struct {
			Name       string          `json:"name"`
			Cells      [][]interface{} `json:"cells"`
			Merges     interface{}     `json:"merges"`
			Cols       []colJSON       `json:"cols"`
			RowHeights []rowJSON       `json:"rowHeights"`
		}
	)
	// Try V1: rows as [][]string
	var v1 sheetV1
	if json.Unmarshal(raw, &v1) == nil && (len(v1.Rows) > 0 || len(v1.Merges) > 0 || len(v1.Cols) > 0) {
		var merges []Merge
		for j := 0; j < len(v1.Merges); j++ {
			mj := v1.Merges[j]
			merges = append(merges, Merge{StartRow: mj.SR, StartCol: mj.SC, EndRow: mj.ER, EndCol: mj.EC})
		}
		var cols []ColSpec
		for j := 0; j < len(v1.Cols); j++ {
			cj := v1.Cols[j]
			col := ColSpec{Index: cj.Index, Width: cj.Width}
			cols = append(cols, col)
		}
		var rowsSpec []RowSpec
		for j := 0; j < len(v1.RowHeights); j++ {
			rj := v1.RowHeights[j]
			row := RowSpec{Index: rj.Index, Height: rj.Height}
			rowsSpec = append(rowsSpec, row)
		}
		return sheetFromRows(defaultName(v1.Name, "Sheet"), v1.Rows, merges, cols, rowsSpec), true
	}
	// Try V2: rows as [][]interface{}
	var v2 sheetV2
	if json.Unmarshal(raw, &v2) == nil && len(v2.Rows) > 0 {
		rows := convertAnyRowsToStrings(v2.Rows)
		merges := parseFlexibleMerges(v2.Merges)
		var cols []ColSpec
		for j := 0; j < len(v2.Cols); j++ {
			cj := v2.Cols[j]
			cols = append(cols, ColSpec{Index: cj.Index, Width: cj.Width})
		}
		var rowsSpec []RowSpec
		for j := 0; j < len(v2.RowHeights); j++ {
			rj := v2.RowHeights[j]
			rowsSpec = append(rowsSpec, RowSpec{Index: rj.Index, Height: rj.Height})
		}
		return sheetFromRows(defaultName(v2.Name, "Sheet"), rows, merges, cols, rowsSpec), true
	}
	// Try V3: cells as [][]interface{}
	var v3 sheetV3
	if json.Unmarshal(raw, &v3) == nil && len(v3.Cells) > 0 {
		// Build cells with types and formulas
		height := len(v3.Cells)
		width := 0
		for i := 0; i < len(v3.Cells); i++ {
			if len(v3.Cells[i]) > width {
				width = len(v3.Cells[i])
			}
		}
		cells := make([][]Cell, height)
		for r := 0; r < height; r++ {
			row := v3.Cells[r]
			cells[r] = make([]Cell, len(row))
			for c := 0; c < len(row); c++ {
				cells[r][c] = parseAnyCell(row[c])
			}
		}
		merges := parseFlexibleMerges(v3.Merges)
		var cols []ColSpec
		for j := 0; j < len(v3.Cols); j++ {
			cj := v3.Cols[j]
			cols = append(cols, ColSpec{Index: cj.Index, Width: cj.Width})
		}
		var rowsSpec []RowSpec
		for j := 0; j < len(v3.RowHeights); j++ {
			rj := v3.RowHeights[j]
			rowsSpec = append(rowsSpec, RowSpec{Index: rj.Index, Height: rj.Height})
		}
		return Sheet{Name: defaultName(v3.Name, "Sheet"), Width: width, Height: height, Cells: cells, Merges: merges, Cols: cols, Rows: rowsSpec}, true
	}
	return Sheet{}, false
}

// convertAnyRowsToStrings converts [][]interface{} to [][]string preserving string representations.
func convertAnyRowsToStrings(rows [][]interface{}) [][]string {
	out := make([][]string, len(rows))
	for r := 0; r < len(rows); r++ {
		row := rows[r]
		out[r] = make([]string, len(row))
		for c := 0; c < len(row); c++ {
			out[r][c] = anyToString(row[c])
		}
	}
	return out
}

func anyToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "TRUE"
		}
		return "FALSE"
	case nil:
		return ""
	default:
		// Best-effort JSON roundtrip
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	}
}

// parseFlexibleMerges supports multiple shapes: array of {SR,SC,ER,EC} or [[sr,sc,er,ec],...]
func parseFlexibleMerges(m interface{}) []Merge {
	if m == nil {
		return nil
	}
	var out []Merge
	// Try known struct array
	switch vv := m.(type) {
	case []interface{}:
		for _, item := range vv {
			switch it := item.(type) {
			case map[string]interface{}:
				sr := toInt(it["SR"]) + toInt(it["sr"]) + toInt(it["r1"]) + toInt(it["startRow"]) // first non-zero wins
				sc := toInt(it["SC"]) + toInt(it["sc"]) + toInt(it["c1"]) + toInt(it["startCol"])
				er := toInt(it["ER"]) + toInt(it["er"]) + toInt(it["r2"]) + toInt(it["endRow"])
				ec := toInt(it["EC"]) + toInt(it["ec"]) + toInt(it["c2"]) + toInt(it["endCol"])
				if sr > 0 && sc > 0 && er >= sr && ec >= sc {
					out = append(out, Merge{StartRow: sr, StartCol: sc, EndRow: er, EndCol: ec})
				}
			case []interface{}:
				if len(it) == 4 {
					sr := toInt(it[0])
					sc := toInt(it[1])
					er := toInt(it[2])
					ec := toInt(it[3])
					if sr > 0 && sc > 0 && er >= sr && ec >= sc {
						out = append(out, Merge{StartRow: sr, StartCol: sc, EndRow: er, EndCol: ec})
					}
				}
			}
		}
	}
	return out
}

func toInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}

// inferCell attempts to parse a string into number, bool, or datetime; falls back to string
func inferCell(s string) Cell {
	t := strings.TrimSpace(s)
	if t == "" {
		return Cell{Type: ValueEmpty}
	}
	// bool
	if t == "true" || t == "TRUE" || t == "True" {
		return Cell{Type: ValueBool, BoolValue: true, StringValue: "TRUE"}
	}
	if t == "false" || t == "FALSE" || t == "False" {
		return Cell{Type: ValueBool, BoolValue: false, StringValue: "FALSE"}
	}
	// epoch detection on pure integers first (seconds or milliseconds)
	digitsOnly := true
	for i := 0; i < len(t); i++ {
		ch := t[i]
		if ch < '0' || ch > '9' {
			digitsOnly = false
			break
		}
	}
	if digitsOnly && len(t) >= 10 { // plausibly a timestamp
		if i, err := strconv.ParseInt(t, 10, 64); err == nil {
			if len(t) >= 13 { // assume milliseconds
				tm := time.UnixMilli(i).UTC()
				return Cell{Type: ValueDateTime, DateEpoch: toExcelSerial(tm), StringValue: t}
			}
			tm := time.Unix(i, 0).UTC()
			return Cell{Type: ValueDateTime, DateEpoch: toExcelSerial(tm), StringValue: t}
		}
	}
	// number with locales, percents, currency, negatives
	if f, ok := parseNumber(t); ok {
		return Cell{Type: ValueNumber, NumberValue: f, StringValue: t}
	}
	// datetime: robust parsing across common variants
	if tm, ok := parseDate(t); ok {
		return Cell{Type: ValueDateTime, DateEpoch: toExcelSerial(tm), StringValue: t}
	}
	// epoch seconds
	if i, err := strconv.ParseInt(t, 10, 64); err == nil {
		// Heuristic: treat large values as milliseconds
		if i > 1_000_000_000_000 { // > ~2001-09 in ms
			tm := time.UnixMilli(i).UTC()
			return Cell{Type: ValueDateTime, DateEpoch: toExcelSerial(tm), StringValue: t}
		}
		tm := time.Unix(i, 0).UTC()
		return Cell{Type: ValueDateTime, DateEpoch: toExcelSerial(tm), StringValue: t}
	}
	return Cell{Type: ValueString, StringValue: s}
}

// parseAnyCell converts various JSON cell encodings into Cell.
// Supported forms:
// - primitive: string/number/bool => inferred via inferCell on string or direct mapping
// - object: {"type":"string|number|bool|date|datetime","value":..., "formula":"..."}
// - object short keys: {"t":"n|s|b|d","v":..., "f":"..."}
func parseAnyCell(v interface{}) Cell {
	switch t := v.(type) {
	case string:
		return inferCell(t)
	case float64:
		return Cell{Type: ValueNumber, NumberValue: t, StringValue: strconv.FormatFloat(t, 'f', -1, 64)}
	case bool:
		if t {
			return Cell{Type: ValueBool, BoolValue: true, StringValue: "TRUE"}
		}
		return Cell{Type: ValueBool, BoolValue: false, StringValue: "FALSE"}
	case map[string]interface{}:
		// formula first
		formula := ""
		if f, ok := t["formula"].(string); ok {
			formula = f
		} else if f, ok := t["f"].(string); ok {
			formula = f
		}
		// value and type
		var (
			cType string
			val   interface{}
		)
		if s, ok := t["type"].(string); ok {
			cType = strings.ToLower(s)
		} else if s, ok := t["t"].(string); ok {
			cType = strings.ToLower(s)
		}
		if v, ok := t["value"]; ok {
			val = v
		} else if v, ok := t["v"]; ok {
			val = v
		}
		if cType == "" {
			// infer from val; if no value but formula present, return formula-only cell
			if val == nil {
				return Cell{Type: ValueEmpty, Formula: formula}
			}
			c := parseAnyCell(val)
			c.Formula = formula
			return c
		}
		switch cType {
		case "s", "str", "string":
			s := anyToString(val)
			return Cell{Type: ValueString, StringValue: s, Formula: formula}
		case "n", "num", "number":
			switch vv := val.(type) {
			case float64:
				return Cell{Type: ValueNumber, NumberValue: vv, StringValue: anyToString(val), Formula: formula}
			case string:
				nstr := normalizeNumberString(vv)
				if f, err := strconv.ParseFloat(nstr, 64); err == nil {
					return Cell{Type: ValueNumber, NumberValue: f, StringValue: vv, Formula: formula}
				}
				// fallback to infer
				c := inferCell(vv)
				c.Formula = formula
				return c
			default:
				c := inferCell(anyToString(val))
				c.Formula = formula
				return c
			}
		case "b", "bool", "boolean":
			switch vv := val.(type) {
			case bool:
				return Cell{Type: ValueBool, BoolValue: vv, StringValue: anyToString(val), Formula: formula}
			case string:
				s := strings.TrimSpace(strings.ToLower(vv))
				return Cell{Type: ValueBool, BoolValue: s == "true" || s == "1", StringValue: strings.ToUpper(s), Formula: formula}
			default:
				return Cell{Type: ValueBool, BoolValue: false, StringValue: "FALSE", Formula: formula}
			}
		case "d", "date", "datetime", "time":
			switch vv := val.(type) {
			case string:
				c := inferCell(vv)
				c.Formula = formula
				if c.Type == ValueDateTime {
					return c
				}
			case float64:
				// Treat as excel serial if reasonable
				if vv > 10 && vv < 1000000 {
					return Cell{Type: ValueDateTime, DateEpoch: vv, StringValue: anyToString(val), Formula: formula}
				}
			}
			// Fallback to string inference
			c := inferCell(anyToString(val))
			c.Formula = formula
			return c
		default:
			c := inferCell(anyToString(val))
			c.Formula = formula
			return c
		}
	default:
		return Cell{Type: ValueEmpty}
	}
}

func normalizeNumberString(in string) string {
	s := strings.TrimSpace(in)
	// remove spaces and apostrophes (thousands separators)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u00A0", "")
	s = strings.ReplaceAll(s, "'", "")
	hasDot := strings.Contains(s, ".")
	hasComma := strings.Contains(s, ",")
	if hasDot && hasComma {
		// consider last symbol as decimal if it's dot, drop commas
		lastDot := strings.LastIndex(s, ".")
		lastComma := strings.LastIndex(s, ",")
		if lastDot > lastComma {
			s = strings.ReplaceAll(s, ",", "")
			return s
		}
		// else decimal is comma, remove dots and replace comma with dot
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
		return s
	}
	if hasComma && !hasDot {
		// likely decimal comma
		s = strings.ReplaceAll(s, ",", ".")
		return s
	}
	return s
}

// parseNumber handles locales, currency symbols, percent, and parentheses negatives
func parseNumber(in string) (float64, bool) {
	s := strings.TrimSpace(in)
	if s == "" {
		return 0, false
	}
	// percent
	isPercent := false
	if strings.HasSuffix(s, "%") {
		isPercent = true
		s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	}
	// parentheses negative
	sign := 1.0
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		sign = -1.0
		s = strings.TrimPrefix(strings.TrimSuffix(s, ")"), "(")
	}
	// leading plus/minus
	if strings.HasPrefix(s, "+") {
		s = strings.TrimPrefix(s, "+")
	} else if strings.HasPrefix(s, "-") {
		sign = -1.0 * sign
		s = strings.TrimPrefix(s, "-")
	}
	// strip currency symbols
	currency := []string{"$", "€", "£", "₽", "¥", "₴", "₺", "₹", "zł", "PLN", "USD", "EUR", "RUB"}
	for i := 0; i < len(currency); i++ {
		s = strings.ReplaceAll(s, currency[i], "")
	}
	s = strings.TrimSpace(s)
	s = normalizeNumberString(s)
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		if isPercent {
			f = f / 100.0
		}
		return sign * f, true
	}
	return 0, false
}

// parseDate tries multiple layouts and returns UTC time on success
func parseDate(in string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.000Z07:00",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		"02.01.2006",
		"02.01.2006 15:04:05",
		"02/01/2006",
		"02/01/2006 15:04:05",
	}
	for i := 0; i < len(layouts); i++ {
		if tm, err := time.Parse(layouts[i], in); err == nil {
			return tm.UTC(), true
		}
	}
	return time.Time{}, false
}

func toExcelSerial(t time.Time) float64 {
	origin := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	diff := t.UTC().Sub(origin)
	return diff.Hours() / 24.0
}
