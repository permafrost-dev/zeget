package files

import (
	"os"
	"path/filepath"
)

type Link struct {
	Newname string
	Oldname string
	Sym     bool
}

func (l Link) Write() error {
	// remove file if it exists already
	os.Remove(l.Newname)
	// make parent directories if necessary
	os.MkdirAll(filepath.Dir(l.Newname), 0755)

	if l.Sym {
		return os.Symlink(l.Oldname, l.Newname)
	}

	return os.Link(l.Oldname, l.Newname)
}
