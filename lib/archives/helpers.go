package archives

import (
	"archive/tar"

	"github.com/permafrost-dev/eget/lib/files"
)

func tarFileType(typ byte) files.FileType {
	switch typ {
	case tar.TypeReg:
		return files.TypeNormal
	case tar.TypeDir:
		return files.TypeDir
	case tar.TypeLink:
		return files.TypeLink
	case tar.TypeSymlink:
		return files.TypeSymlink
	}

	return files.TypeOther
}
