package archives

import (
	"archive/tar"

	"github.com/permafrost-dev/zeget/lib/files"
)

func TarFileType(typ byte) files.FileType {
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
