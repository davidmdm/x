//go:build !race

package xruntime

func RaceEnabled() bool { return false }
