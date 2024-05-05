package extraction

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/permafrost-dev/eget/lib/utilities"
)

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

	possible := !mode.IsDir() && utilities.IsExec(name, mode.Perm())
	return fmatch && possible, possible
}

func (b *BinaryChooser) String() string {
	return fmt.Sprintf("exe `%s`", b.Tool)
}
