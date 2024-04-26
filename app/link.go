package app

import (
	"os"
	"path/filepath"
)

type Link struct {
	newname string
	oldname string
	sym     bool
}

func (l Link) Write() error {
	// remove file if it exists already
	os.Remove(l.newname)
	// make parent directories if necessary
	os.MkdirAll(filepath.Dir(l.newname), 0755)

	if l.sym {
		return os.Symlink(l.oldname, l.newname)
	}

	return os.Link(l.oldname, l.newname)
}
