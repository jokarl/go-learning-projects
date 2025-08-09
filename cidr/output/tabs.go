package output

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

type TabFormatter struct {
	minWidth   int
	tabWidth   int
	padding    int
	padChar    byte
	flags      uint
	labelColor color
}

const (
	defaultMinWidth   = 0
	defaultTabWidth   = 0
	defaultPadding    = 2
	defaultPadChar    = ' '
	defaultFlags      = 0
	defaultLabelColor = Blue
)

// Option is a functional option for TabFormatter.
type Option func(*TabFormatter)

func WithMinWidth(v int) Option {
	return func(t *TabFormatter) { t.minWidth = v }
}
func WithTabWidth(v int) Option {
	return func(t *TabFormatter) { t.tabWidth = v }
}
func WithPadding(v int) Option {
	return func(t *TabFormatter) { t.padding = v }
}
func WithPadChar(v byte) Option {
	return func(t *TabFormatter) { t.padChar = v }
}
func WithFlags(v uint) Option {
	return func(t *TabFormatter) { t.flags = v }
}

// newTabFormatter builds a TabFormatter, applying any options provided.
// If no options are passed, defaults are used.
func newTabFormatter(opts ...Option) *TabFormatter {
	tf := &TabFormatter{
		minWidth:   defaultMinWidth,
		tabWidth:   defaultTabWidth,
		padding:    defaultPadding,
		padChar:    defaultPadChar,
		flags:      defaultFlags,
		labelColor: defaultLabelColor,
	}
	for _, opt := range opts {
		opt(tf)
	}
	return tf
}

type tabWriter struct {
	*tabwriter.Writer
	err error
}

func (tw *tabWriter) write(s string) {
	if tw.err != nil {
		return
	}
	_, tw.err = fmt.Fprint(tw.Writer, s)
}

func (tw *tabWriter) tab() {
	tw.write("\t")
}

func (tw *tabWriter) newline() {
	tw.write("\n")
}

func (tf *TabFormatter) Print(data any) error {
	return tf.Fprint(os.Stdout, data)
}

func (tf *TabFormatter) Fprint(w io.Writer, data any) error {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return fmt.Errorf("tab formatter: nil data")
	}

	// Unwrap pointers
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Errorf("tab formatter: nil pointer")
		}
		v = v.Elem()
	}

	tw := &tabWriter{
		Writer: tabwriter.NewWriter(w, tf.minWidth, tf.tabWidth, tf.padding, tf.padChar, tf.flags),
	}
	defer func() {
		if flushErr := tw.Writer.Flush(); flushErr != nil && tw.err == nil {
			tw.err = flushErr
		}
	}()

	switch v.Kind() {
	case reflect.Struct:
		// One object -> rows "Label:\tValue"
		fields := collectTabFields(v)
		for _, f := range fields {
			tw.write(fmt.Sprintf("%s%s:%s\t%s", tf.labelColor, f.label, Reset, f.value))
			tw.newline()
		}

	case reflect.Slice, reflect.Array:
		// Table: header row + data rows
		if v.Len() == 0 {
			return nil // nothing to print
		}
		row0 := deref(v.Index(0))
		if row0.Kind() != reflect.Struct {
			return fmt.Errorf("tab formatter: slice element must be struct, got %s", row0.Kind())
		}
		fields := collectTabFields(row0)

		// Header
		for i, f := range fields {
			if i > 0 {
				tw.tab()
			}
			tw.write(f.label)
		}
		tw.newline()

		// Rows
		for i := 0; i < v.Len(); i++ {
			elem := deref(v.Index(i))
			fields := collectTabFields(elem)
			for j, f := range fields {
				if j > 0 {
					tw.tab()
				}
				tw.write(f.value)
			}
			tw.newline()
		}
	default:
		return fmt.Errorf("tab formatter: unsupported kind %s", v.Kind())
	}

	return tw.err
}

// helpers

type tabField struct {
	label string
	value string
}

func collectTabFields(v reflect.Value) []tabField {
	t := v.Type()
	out := make([]tabField, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		if sf.PkgPath != "" { // unexported
			continue
		}

		tag := parseTabTag(sf)
		if tag.skip {
			continue
		}

		fv := v.Field(i)
		if tag.omit && isZeroValue(fv) {
			continue
		}

		valStr := stringifyField(fv)
		out = append(out, tabField{label: tag.label, value: valStr})
	}
	return out
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}
	return v
}

func stringifyField(v reflect.Value) string {
	// nil ptr/interface
	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && v.IsNil() {
		return ""
	}
	// Prefer direct Stringer
	if v.CanInterface() {
		if s, ok := v.Interface().(fmt.Stringer); ok {
			return s.String()
		}
	}
	// Try pointer receiver Stringer (e.g., big.Int)
	if v.CanAddr() {
		if s, ok := v.Addr().Interface().(fmt.Stringer); ok {
			return s.String()
		}
	}

	// Deref pointers for basic handling
	v = deref(v)

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Slice, reflect.Array:
		var b strings.Builder
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(stringifyField(v.Index(i)))
		}
		return b.String()
	case reflect.Struct:
		// Fallback for structs without Stringer
		if v.CanInterface() {
			return fmt.Sprintf("%v", v.Interface())
		}
	}
	if v.CanInterface() {
		return fmt.Sprintf("%v", v.Interface())
	}
	return ""
}

type tabTag struct {
	label string
	omit  bool // omit field if zero-value
	skip  bool // skip unconditionally (tab:"-")
}

func parseTabTag(sf reflect.StructField) tabTag {
	// Prefer `tab`, fall back to `tabs`
	raw := sf.Tag.Get("tab")
	if raw == "" {
		raw = sf.Tag.Get("tabs")
	}
	if raw == "-" {
		return tabTag{skip: true}
	}
	parts := strings.Split(raw, ",")
	label := strings.TrimSpace(parts[0])

	var omit bool
	for _, p := range parts[1:] {
		switch strings.TrimSpace(p) {
		case "omit", "omitempty": // support both spellings
			omit = true
		}
	}
	if label == "" {
		label = sf.Name
	}
	return tabTag{label: label, omit: omit}
}

func isZeroValue(v reflect.Value) bool {
	// Handle nil pointers/interfaces
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	// Prefer reflect.Value.IsZero where possible
	if v.IsValid() {
		return v.IsZero()
	}
	return true
}
