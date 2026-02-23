package xcontainer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	t.Run("intersection", func(t *testing.T) {
		require.Equal(
			t,
			[]int{4},
			Set[int].Intersection(
				ToSet([]int{1, 2, 3, 4}),
				ToSet([]int{3, 4, 5, 6}),
				ToSet([]int{4, 5, 6, 7, 8}),
			).Collect(),
		)
	})
}
