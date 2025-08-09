package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Formatter interface {
	Fprint(w io.Writer, data any) error
	Print(data any) error
}

const DefaultFormat = "tab"

var availableFormats = map[string]Formatter{
	"tab":  newTabFormatter(),
	"json": newJSONFormatter(),
}

func Formats() []string {
	keys := make([]string, 0, len(availableFormats))
	for k := range availableFormats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func GetFormatter(name string) (Formatter, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		n = DefaultFormat
	}
	if f, ok := availableFormats[n]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("unsupported format %q (choose one of: %s)",
		name, strings.Join(Formats(), ", "))
}
