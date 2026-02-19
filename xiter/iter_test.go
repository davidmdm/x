package xiter

import (
	"iter"
	"reflect"
	"slices"
	"testing"
)

func TestJoin(t *testing.T) {
	cases := []struct {
		Name   string
		Input  [][]int
		Output []int
	}{
		{
			Name:   "no sequences",
			Input:  [][]int{},
			Output: nil,
		},
		{
			Name:   "single sequence",
			Input:  [][]int{{1, 2, 3}},
			Output: []int{1, 2, 3},
		},
		{
			Name:   "multiple sequences",
			Input:  [][]int{{1, 2, 3}, {4, 5}, {6}},
			Output: []int{1, 2, 3, 4, 5, 6},
		},
	}

	asSequence := func(values []int) iter.Seq[int] {
		return func(yield func(int) bool) {
			for _, value := range values {
				if !yield(value) {
					return
				}
			}
		}
	}

	asSequences := func(listOfValues [][]int) []iter.Seq[int] {
		if listOfValues == nil {
			return nil
		}
		ret := make([]iter.Seq[int], len(listOfValues))
		for i, values := range listOfValues {
			ret[i] = asSequence(values)
		}
		return ret
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			output := slices.Collect(Join(asSequences(tc.Input)...))
			if !reflect.DeepEqual(output, tc.Output) {
				t.Fatalf("expected output to be %v but got %v", tc.Output, output)
			}
		})
	}
}
