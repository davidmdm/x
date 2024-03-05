package xfs

import (
	"io/fs"
)

//go:generate moq -out fs_mock.go . FS
type FS = fs.FS

//go:generate moq -out file_mock.go . File
type File = fs.File

//go:generate moq -out file_info_mock.go . FileInfo
type FileInfo = fs.FileInfo
