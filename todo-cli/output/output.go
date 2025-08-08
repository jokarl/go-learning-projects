package output

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"text/tabwriter"
)

// Emojier is an interface that requires the Emoji method to return a string.
// The returned string should be a Unicode emoji.
type Emojier interface {
	Emoji() string
}

// Table prints the data in a tabular format.
// It expects a slice of structs and prints the headers and values in a tabular format.
// The headers are derived from the struct field names.
func Table(data interface{}) error {
	v, t, err := validate(data)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	defer func(w *tabwriter.Writer) {
		if err := w.Flush(); err != nil {
			fmt.Printf("tabwriter flush error: %v", err)
		}
	}(w)

	// Print headers
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if _, err := fmt.Fprintf(w, "%s\t", name); err != nil {
			return fmt.Errorf("error writing header %s: %v\n", name, err)
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("error printing header: %v\n", err)
	}

	// Print rows
	for i := 0; i < v.Len(); i++ {
		e := v.Index(i)
		for j := 0; j < e.NumField(); j++ {
			field := e.Field(j).Interface()
			// Check if the field implements the Emojier interface
			if em, ok := field.(Emojier); ok {
				if _, err := fmt.Fprintf(w, "%v\t", em.Emoji()); err != nil {
					return fmt.Errorf("error writing field %d of row %d: %v\n", j, i, err)
				}
			} else {
				if _, err := fmt.Fprintf(w, "%v\t", field); err != nil {
					return fmt.Errorf("error writing field %d of row %d: %v\n", j, i, err)
				}
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("error printing row data: %v\n", err)
		}
	}

	return nil
}

func JSON(data interface{}) error {
	_, _, err := validate(data)
	if err != nil {
		return err
	}
	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(d))
	return nil
}

func validate(data interface{}) (reflect.Value, reflect.Type, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return reflect.Value{}, nil, fmt.Errorf("expected a slice, got %v", v.Kind())
	}
	if v.Len() == 0 {
		fmt.Println("There are no items to display")
		os.Exit(0)
	}

	t := v.Index(0).Type()
	if t.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("slice elements are not structs")
	}
	return v, t, nil
}
