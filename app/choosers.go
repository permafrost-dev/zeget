package app

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// A Chooser selects a file. It may list the file as a direct match (should be
// immediately extracted if found), or a possible match (only extract if it is
// the only match, or if the user manually requests it).
type Chooser interface {
	Choose(name string, dir bool, mode fs.FileMode) (direct bool, possible bool)
}

// A BinaryChooser selects executable files. If the executable file has the
// name 'Tool' it is considered a direct match. If the file is only executable,
// it is a possible match.
type BinaryChooser struct {
	Tool string
}

func (b *BinaryChooser) Choose(name string, dir bool, mode fs.FileMode) (bool, bool) {
	if dir {
		return false, false
	}

	fmatch := filepath.Base(name) == b.Tool ||
		filepath.Base(name) == b.Tool+".exe" ||
		filepath.Base(name) == b.Tool+".appimage"

	possible := !mode.IsDir() && IsExec(name, mode.Perm())
	return fmatch && possible, possible
}

func (b *BinaryChooser) String() string {
	return fmt.Sprintf("exe `%s`", b.Tool)
}

// LiteralFileChooser selects files with the name 'File'.
type LiteralFileChooser struct {
	File string
}

func (lf *LiteralFileChooser) Choose(name string, _ bool, _ fs.FileMode) (bool, bool) {
	return false, filepath.Base(name) == filepath.Base(lf.File) && strings.HasSuffix(name, lf.File)
}

func (lf *LiteralFileChooser) String() string {
	return fmt.Sprintf("`%s`", lf.File)
}

type GlobChooser struct {
	expr string
	g    glob.Glob
	all  bool
}

func NewGlobChooser(gl string) (*GlobChooser, error) {
	g, err := glob.Compile(gl, '/')
	return &GlobChooser{
		g:    g,
		expr: gl,
		all:  gl == "*" || gl == "/",
	}, err
}

func (gc *GlobChooser) Choose(name string, _ bool, _ fs.FileMode) (bool, bool) {
	if gc.all {
		return true, true
	}
	if len(name) > 0 && name[len(name)-1] == '/' {
		name = name[:len(name)-1]
	}
	return false, gc.g.Match(filepath.Base(name)) || gc.g.Match(name)
}

func (gc *GlobChooser) String() string {
	return fmt.Sprintf("`%s`", gc.expr)
}
