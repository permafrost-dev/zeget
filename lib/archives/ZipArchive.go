package archives

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/permafrost-dev/zeget/lib/files"
)

type ZipArchive struct {
	r   *zip.Reader
	idx int
}

// decompressor does nothing for a zip archive because it already has built-in
// compression.
func NewZipArchive(data []byte, _ DecompressFunc) (Archive, error) {
	r := bytes.NewReader(data)
	zr, err := zip.NewReader(r, int64(len(data)))
	return &ZipArchive{
		r:   zr,
		idx: -1,
	}, err
}

func (z *ZipArchive) Next() (files.File, error) {
	z.idx++

	if z.idx < 0 || z.idx >= len(z.r.File) {
		return files.File{}, io.EOF
	}

	f := z.r.File[z.idx]

	typ := files.TypeNormal
	if strings.HasSuffix(f.Name, "/") {
		typ = files.TypeDir
	}

	return files.File{
		Name: f.Name,
		Mode: f.Mode(),
		Type: typ,
	}, nil
}

func (z *ZipArchive) ReadAll() ([]byte, error) {
	if z.idx < 0 || z.idx >= len(z.r.File) {
		return nil, io.EOF
	}
	f := z.r.File[z.idx]
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("zip extract: %w", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	return data, err
}
