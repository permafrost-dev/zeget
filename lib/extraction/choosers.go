package extraction

import (
	"io/fs"
)

// A Chooser selects a file. It may list the file as a direct match (should be
// immediately extracted if found), or a possible match (only extract if it is
// the only match, or if the user manually requests it).
type Chooser interface {
	Choose(name string, dir bool, mode fs.FileMode) (direct bool, possible bool)
}
