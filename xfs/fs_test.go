package xfs_test

import (
	"testing"

	"github.com/davidmdm/x/xfs"
	"github.com/stretchr/testify/require"
)

func TestDir(t *testing.T) {
	// calls os.DirFS under the hood but casts it to extended interface.
	// This tests makes sure that this cast is successful and doesn't panic.
	fs := xfs.Dir(".")
	require.Equal(t, ".", fs.DirName())

	data, err := fs.ReadFile("go.mod")
	require.NoError(t, err)
	require.Greater(t, len(data), 0)
}
