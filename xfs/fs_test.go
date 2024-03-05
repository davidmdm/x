package xfs_test

import (
	"testing"

	"github.com/davidmdm/x/xfs"
)

func TestDir(t *testing.T) {
	// calls os.DirFS under the hood but casts it to extended interface.
	// This tests makes sure that this cast is successful and doesn't panic.
	_ = xfs.Dir(".")
}
