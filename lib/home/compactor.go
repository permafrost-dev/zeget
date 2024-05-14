package home

import "strings"

type PathCompactor struct {
	HomePath string
}

func NewPathCompactor() *PathCompactor {
	path, _ := NewPathExpander().HomeDirectory()

	return &PathCompactor{
		HomePath: path,
	}
}

func (pc *PathCompactor) Compact(path string) string {
	if strings.HasPrefix(path, pc.HomePath) {
		path = strings.Replace(path, pc.HomePath, "~", 1)
	}
	return path
}
