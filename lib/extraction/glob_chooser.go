package extraction

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/gobwas/glob"
)

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
