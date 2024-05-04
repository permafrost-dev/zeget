package archives

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"

	"github.com/permafrost-dev/eget/lib/files"
)

type TarArchive struct {
	r *tar.Reader
}

func NewTarArchive(data []byte, decompress DecompressFunc) (Archive, error) {
	r := bytes.NewReader(data)
	dr, err := decompress(r)
	if err != nil {
		return nil, err
	}
	return &TarArchive{
		r: tar.NewReader(dr),
	}, nil
}

func (t *TarArchive) Next() (files.File, error) {
	for {
		hdr, err := t.r.Next()
		if err != nil {
			return files.File{}, err
		}
		ft := TarFileType(hdr.Typeflag)
		if ft != files.TypeOther {
			return files.File{
				Name:     hdr.Name,
				LinkName: hdr.Linkname,
				Mode:     fs.FileMode(hdr.Mode),
				Type:     ft,
			}, err
		}
	}
}

func (t *TarArchive) ReadAll() ([]byte, error) {
	return io.ReadAll(t.r)
}
