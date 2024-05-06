package files

import (
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"
)

type Link struct {
	Newname string
	Oldname string
	Sym     bool
	Fs      vfs.FS
	Cleanup func()
}

func NewLink(oldname, newname string) *Link {
	return &Link{
		Oldname: oldname,
		Newname: newname,
		Sym:     true,
		Fs:      vfs.OSFS,
	}
}

func (lnk *Link) Write() error {
	if _, err := lnk.Fs.Stat(lnk.Newname); err == nil {
		lnk.Fs.Remove(lnk.Newname)
	}

	lnk.Fs.Mkdir(filepath.Dir(lnk.Newname), 0755)

	return lnk.Fs.Symlink(lnk.Oldname, lnk.Newname)
}
