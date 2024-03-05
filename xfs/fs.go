package xfs

import (
	"io/fs"
	"os"
)

//go:generate moq -out fs_mock.go . FS
type FS = interface {
	fs.FS
	fs.StatFS
	fs.ReadDirFS
	fs.ReadFileFS
}

//go:generate moq -out dir_entry_mock.go . DirEntry
type DirEntry = fs.DirEntry

//go:generate moq -out file_mock.go . File
type File = fs.File

//go:generate moq -out file_info_mock.go . FileInfo
type FileInfo = fs.FileInfo

func Dir(name string) FS {
	return os.DirFS(name).(FS)
}
