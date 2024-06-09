package xruntime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStack(t *testing.T) {
	DoThing := func() {
		stack := GetCallStack(3)

		fnNames := []string{}
		for _, frame := range stack.Frames {
			fnNames = append(fnNames, frame.Function)
		}

		require.Equal(
			t,
			[]string{
				"github.com/davidmdm/x/xruntime.TestStack.func1",
				"github.com/davidmdm/x/xruntime.TestStack.func2",
				"github.com/davidmdm/x/xruntime.TestStack",
			},
			fnNames,
		)
	}

	Run := func() {
		DoThing()
	}

	Run()
}
