package files

import "io/fs"

type FileType byte

const (
	TypeNormal FileType = iota
	TypeDir
	TypeLink
	TypeSymlink
	TypeOther
)

type File struct {
	Name     string
	LinkName string
	Mode     fs.FileMode
	Type     FileType
}

func (f File) Dir() bool {
	return f.Type == TypeDir
}
