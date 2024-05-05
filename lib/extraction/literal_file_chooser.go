package extraction

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

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
