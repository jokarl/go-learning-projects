package output

import (
	"encoding/json"
	"io"
	"os"
)

type JSONFormatter struct{}

func newJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (tf *JSONFormatter) Print(data any) error {
	return tf.Fprint(os.Stdout, data)
}

func (tf *JSONFormatter) Fprint(w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return err
	}
	return nil
}
