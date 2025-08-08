package slice

import (
	"slices"
	"testing"
)

// TestDiff calls slice.Diff with a name, checking
// for a valid return value.
func TestDiff(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		// arrange
		a := []int{1, 2, 3, 4, 5}
		b := []int{4, 5}
		want := []int{1, 2, 3}

		// act
		r := Diff(a, b)

		// assert
		if !slices.Equal(r, want) {
			t.Errorf(`Diff(%v, %v) = %v, want match for %v, nil`, a, b, r, want)
		}
	})

	t.Run("string", func(t *testing.T) {
		// arrange
		a := []string{"1", "2", "3", "4", "5"}
		b := []string{"4", "5"}
		want := []string{"1", "2", "3"}

		// act
		r := Diff(a, b)

		// assert
		if !slices.Equal(r, want) {
			t.Errorf(`Diff(%v, %v) = %v, want match for %v, nil`, a, b, r, want)
		}
	})

	t.Run("struct", func(t *testing.T) {
		// arrange
		type testStruct struct {
			ID   int
			Name string
		}
		a := []testStruct{{ID: 1, Name: "a1"}, {ID: 2, Name: "a2"}}
		b := []testStruct{{ID: 2, Name: "a2"}}
		want := []testStruct{{ID: 1, Name: "a1"}}

		// act
		r := Diff(a, b)

		// assert
		if !slices.Equal(r, want) {
			t.Errorf(`Diff(%v, %v) = %v, want match for %v, nil`, a, b, r, want)
		}
	})

}
