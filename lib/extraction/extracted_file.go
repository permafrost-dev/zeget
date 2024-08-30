package extraction

import (
	"io/fs"

	"github.com/permafrost-dev/zeget/lib/utilities"
)

// An ExtractedFile contains the data, name, and permissions of a file in the archive.
type ExtractedFile struct {
	Name        string // name to extract to
	ArchiveName string // name in archive
	mode        fs.FileMode
	Extract     func(to string) error
	Dir         bool
}

// Mode returns the filemode of the extracted file.
func (e ExtractedFile) Mode() fs.FileMode {
	return utilities.ModeFrom(e.Name, e.mode)
}

func (e ExtractedFile) SetMode(mode fs.FileMode) {
	e.mode = mode
}

// String returns the archive name of this extracted file
func (e ExtractedFile) String() string {
	return e.ArchiveName
}
