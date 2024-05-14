package archives

import (
	"io"

	"github.com/permafrost-dev/zeget/lib/files"
)

type ArchiveFunc func(data []byte, decomp DecompressFunc) (Archive, error)
type DecompressFunc func(r io.Reader) (io.Reader, error)

type Archive interface {
	Next() (files.File, error)
	ReadAll() ([]byte, error)
}
