package app

import "io/fs"

// An ExtractedFile contains the data, name, and permissions of a file in the
// archive.
type ExtractedFile struct {
	Name        string // name to extract to
	ArchiveName string // name in archive
	mode        fs.FileMode
	Extract     func(to string) error
	Dir         bool
}

// Mode returns the filemode of the extracted file.
func (e ExtractedFile) Mode() fs.FileMode {
	return ModeFrom(e.Name, e.mode)
}

// String returns the archive name of this extracted file
func (e ExtractedFile) String() string {
	return e.ArchiveName
}
