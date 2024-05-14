package extraction

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/permafrost-dev/zeget/lib/archives"
	"github.com/permafrost-dev/zeget/lib/files"
	"github.com/permafrost-dev/zeget/lib/targetfile"
	"github.com/permafrost-dev/zeget/lib/utilities"
	"github.com/twpayne/go-vfs/v5"
	"github.com/ulikunitz/xz"
)

// An Extractor reads in some archive data and extracts a particular file from
// it. If there are multiple candidates it returns a list and an error
// explaining what happened.
type Extractor interface {
	Extract(data []byte, multiple bool) (ExtractedFile, []ExtractedFile, error)
}

// NewExtractor constructs an extractor for the given archive file using the
// given chooser. It will construct extractors for files ending in '.tar.gz',
// '.tar.bz2', '.tar', '.zip'. After these matches, if the file ends with
// '.gz', '.bz2' it will be decompressed and copied. Other files will simply
// be copied without any decompression or extraction.
func NewExtractor(fs vfs.FS, filename string, tool string, chooser Chooser) Extractor {
	if tool == "" {
		tool = filename
	}

	gunzipper := func(r io.Reader) (io.Reader, error) {
		return gzip.NewReader(r)
	}
	b2unzipper := func(r io.Reader) (io.Reader, error) {
		return bzip2.NewReader(r), nil
	}
	xunzipper := func(r io.Reader) (io.Reader, error) {
		return xz.NewReader(bufio.NewReader(r))
	}
	zstdunzipper := func(r io.Reader) (io.Reader, error) {
		return zstd.NewReader(r)
	}
	nounzipper := func(r io.Reader) (io.Reader, error) {
		return r, nil
	}

	switch {
	case strings.HasSuffix(filename, ".tar.gz"), strings.HasSuffix(filename, ".tgz"):
		return NewArchiveExtractor(chooser, archives.NewTarArchive, gunzipper, fs)

	case strings.HasSuffix(filename, ".tar.bz2"), strings.HasSuffix(filename, ".tbz"):
		return NewArchiveExtractor(chooser, archives.NewTarArchive, b2unzipper, fs)

	case strings.HasSuffix(filename, ".tar.xz"), strings.HasSuffix(filename, ".txz"):
		return NewArchiveExtractor(chooser, archives.NewTarArchive, xunzipper, fs)

	case strings.HasSuffix(filename, ".tar.zst"):
		return NewArchiveExtractor(chooser, archives.NewTarArchive, zstdunzipper, fs)

	case strings.HasSuffix(filename, ".tar"):
		return NewArchiveExtractor(chooser, archives.NewTarArchive, nounzipper, fs)

	case strings.HasSuffix(filename, ".zip"):
		return NewArchiveExtractor(chooser, archives.NewZipArchive, nounzipper, fs)

	case strings.HasSuffix(filename, ".gz"):
		return NewSingleFileExtractor(tool, filename, gunzipper, fs)

	case strings.HasSuffix(filename, ".bz2"):
		return NewSingleFileExtractor(tool, filename, b2unzipper, fs)

	case strings.HasSuffix(filename, ".xz"):
		return NewSingleFileExtractor(tool, filename, xunzipper, fs)

	case strings.HasSuffix(filename, ".zst"):
		return NewSingleFileExtractor(tool, filename, zstdunzipper, fs)

	default:
		return NewSingleFileExtractor(tool, filename, nounzipper, fs)
	}
}

type ArchiveExtractor struct {
	File       Chooser
	Ar         archives.ArchiveFunc
	Decompress archives.DecompressFunc
	Fs         vfs.FS
}

func NewArchiveExtractor(file Chooser, ar archives.ArchiveFunc, decompress archives.DecompressFunc, fs interface{}) *ArchiveExtractor {
	if fs == nil {
		fs = vfs.OSFS
	}

	return &ArchiveExtractor{
		File:       file,
		Ar:         ar,
		Decompress: decompress,
		Fs:         fs.(vfs.FS),
	}
}

func (a *ArchiveExtractor) Extract(data []byte, multiple bool) (ExtractedFile, []ExtractedFile, error) {
	var candidates []ExtractedFile
	var dirs []string

	ar, err := a.Ar(data, a.Decompress)
	if err != nil {
		return ExtractedFile{}, nil, err
	}
	for {
		f, err := ar.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ExtractedFile{}, nil, fmt.Errorf("extract: %w", err)
		}
		var hasdir bool
		for _, d := range dirs {
			if strings.HasPrefix(f.Name, d) {
				hasdir = true
				break
			}
		}
		if hasdir {
			continue
		}
		direct, possible := a.File.Choose(f.Name, f.Dir(), f.Mode)
		if !direct && !possible {
			continue
		}

		name := utilities.GetRename(f.Name, f.Name)

		fdata, err := ar.ReadAll()
		if err != nil {
			return ExtractedFile{}, nil, fmt.Errorf("extract: %w", err)
		}

		var extract func(to string) error

		extract = func(to string) error {
			tf := targetfile.GetTargetFile(a.Fs, to, utilities.ModeFrom(name, f.Mode), true)

			if tf.Err != nil {
				return fmt.Errorf("extract: %w", err)
			}

			return tf.Write(fdata, true)
		}

		if f.Dir() {
			extract, dirs = a.handleDirs(f, data, dirs)
		}

		ef := ExtractedFile{
			Name:        name,
			ArchiveName: f.Name,
			mode:        f.Mode,
			Extract:     extract,
			Dir:         f.Dir(),
		}
		if direct && !multiple {
			return ef, nil, err
		}
		if err == nil {
			candidates = append(candidates, ef)
		}

	}

	if len(candidates) == 1 {
		return candidates[0], nil, nil
	}

	if len(candidates) == 0 {
		return ExtractedFile{}, candidates, fmt.Errorf("target %v not found in archive", a.File)
	}

	return ExtractedFile{}, candidates, fmt.Errorf("%d candidates for target %v found", len(candidates), a.File)
}

func (a *ArchiveExtractor) handleDirs(f files.File, data []byte, dirs []string) (func(to string) error, []string) {
	directories := append(dirs, f.Name)

	extract := func(to string) error {
		ar, err := a.Ar(data, a.Decompress)
		if err != nil {
			return err
		}
		var links []files.Link
		for {
			subf, err := ar.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				return fmt.Errorf("extract: %w", err)
			}

			if !strings.HasPrefix(subf.Name, f.Name) {
				continue
			}

			if subf.Dir() {
				// TODO implement MkdirAll
				a.Fs.Mkdir(filepath.Join(to, subf.Name[len(f.Name):]), 0755)
				continue
			}

			if subf.Type == files.TypeLink || subf.Type == files.TypeSymlink {
				newname := filepath.Join(to, subf.Name[len(f.Name):])
				oldname := subf.LinkName
				links = append(links, files.Link{
					Newname: newname,
					Oldname: oldname,
					Sym:     subf.Type == files.TypeSymlink,
				})
				continue
			}

			fdata, err := ar.ReadAll()
			if err != nil {
				return fmt.Errorf("extract: %w", err)
			}
			name := filepath.Join(to, subf.Name[len(f.Name):])

			tf := targetfile.GetTargetFile(a.Fs, name, subf.Mode, true)
			if err = tf.Write(fdata, true); err != nil {
				return fmt.Errorf("extract: %w", err)
			}
		}

		for _, l := range links {
			l.Fs = a.Fs
			if err := l.Write(); err != nil && err != os.ErrExist {
				return fmt.Errorf("extract: %w", err)
			}
		}

		return nil
	}

	return extract, directories
}

// SingleFileExtractor extracts files called 'Name' after decompressing the
// file with 'Decompress'.
type SingleFileExtractor struct {
	Rename     string
	Name       string
	Decompress func(r io.Reader) (io.Reader, error)
	Fs         vfs.FS
}

func (sf *SingleFileExtractor) Extract(data []byte, _ bool) (ExtractedFile, []ExtractedFile, error) {
	name := utilities.GetRename(sf.Rename, sf.Name)

	return ExtractedFile{
		Name:        name,
		ArchiveName: sf.Name,
		mode:        0666,
		Extract: func(to string) error {
			r := bytes.NewReader(data)
			dr, err := sf.Decompress(r)
			if err != nil {
				return err
			}

			decdata, err := io.ReadAll(dr)
			if err != nil {
				return err
			}

			tf := targetfile.GetTargetFile(sf.Fs, to, utilities.ModeFrom(name, 0666), true)
			return tf.Write(decdata, true)
		},
	}, nil, nil
}

func NewSingleFileExtractor(name string, rename string, decompress func(r io.Reader) (io.Reader, error), fs interface{}) *SingleFileExtractor {
	return &SingleFileExtractor{
		Name:       name,
		Rename:     rename,
		Decompress: decompress,
		Fs:         fs.(vfs.FS),
	}
}
