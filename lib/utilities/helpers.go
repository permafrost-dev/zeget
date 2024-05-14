package utilities

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	. "github.com/permafrost-dev/eget/lib/assets"
)

// async get the modification time of a filesystem file:
// Bintime returns the modification time of a file or directory.
func Bintime(bin string, to string) (t time.Time) {
	file := ""
	dir := "."

	if to != "" && IsDirectory(to) {
		// direct directory
		dir = to
	} else if ebin := os.Getenv("EGET_BIN"); ebin != "" {
		dir = ebin
	}

	if to != "" && !strings.ContainsRune(to, os.PathSeparator) {
		// path joined possible with eget bin
		bin = to
	} else if to != "" && !IsDirectory(to) {
		// direct path
		file = to
	}

	if file == "" {
		file = filepath.Join(dir, bin)
	}

	fi, err := os.Stat(file)
	if err != nil {
		return
	}

	return fi.ModTime()
}

// IsURL returns true if s is a valid URL.
func IsURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// Cut is strings.Cut
func Cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

// IsGithubURL returns true if s is a URL with github.com as the host.
func IsGithubURL(s string) bool {
	var ghrgx = regexp.MustCompile(`^(http(s)?://)?github\.com/[\w,\-,_]+/[\w,\-,_]+(.git)?(/)?$`)
	return ghrgx.MatchString(s)
}

func IsInvalidGithubURL(s string) bool {
	containsDomain := strings.HasPrefix(s, "github.com") || strings.HasPrefix(s, "https://github.com")

	return containsDomain && !IsGithubURL(s)
}

func IsNonGithubURL(s string) bool {
	return IsURL(s) && !IsGithubURL(s)
}

func RepositoryNameFromGithubURL(s string) (name string, found bool) {
	if !IsGithubURL(s) {
		return "", false
	}

	pattern := regexp.MustCompile(`github\.com/([\w\-_]+/[\w\-_]+)(\.git)?(/)?$`)
	matches := pattern.FindStringSubmatch(s)

	return matches[1], true
}

// IsValidRepositoryReference returns true if s is a valid repository reference in the form of "owner/repo".
func IsValidRepositoryReference(s string) bool {
	if strings.Count(s, "/") != 1 || len(s) < 3 {
		return false
	}

	pattern := regexp.MustCompile(`^[\w\-_]+/[\w\-_]+$`)
	return pattern.MatchString(s)
}

type RepositoryReference struct {
	Owner string
	Name  string
}

func ParseRepositoryReference(s string) (*RepositoryReference, error) {
	if !IsValidRepositoryReference(s) {
		return nil, NewInvalidGitHubProjectReferenceError(s)
	}

	// parts is guaranteed to have 2 elements because IsValidRepositoryReference checks, so no need to check for bounds
	parts := strings.Split(s, "/")

	return &RepositoryReference{Owner: parts[0], Name: parts[len(parts)-1]}, nil
}

func (rr *RepositoryReference) String() string {
	return fmt.Sprintf("%s/%s", rr.Owner, rr.Name)
}

// IsLocalFile returns true if the file at 's' exists.
func IsLocalFile(s string) bool {
	if s == "" {
		return false
	}

	_, err := os.Stat(s)
	return err == nil
}

// IsDirectory returns true if the file at 'path' is a directory.
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// searches for an asset thaat has the same name as the requested one but
// ending with .sha256 or .sha256sum
func FindChecksumAsset(asset Asset, assets []Asset) Asset {
	for _, a := range assets {
		if a.Name == asset.Name+".sha256sum" || a.Name == asset.Name+".sha256" {
			return a
		}
	}
	return Asset{}
}

// IsDefinitelyNotExec returns true if the file is definitely not an executable.
func IsDefinitelyNotExec(file string) bool {
	// file is definitely not executable if it is .deb, .1, or .txt
	return strings.HasSuffix(file, ".deb") ||
		strings.HasSuffix(file, ".1") ||
		strings.HasSuffix(file, ".txt")
}

// IsExec returns true if the file is an executable based on the file name or the file mode (executable bit).
func IsExec(file string, mode os.FileMode) bool {
	if IsDefinitelyNotExec(file) {
		return false
	}

	// file is executable if it is one of the following:
	// *.exe, *.appimage, no extension, executable file permissions
	return strings.HasSuffix(file, ".exe") ||
		strings.HasSuffix(file, ".appimage") ||
		!strings.Contains(file, ".") ||
		mode&0111 != 0
}

// ModeFrom returns the mode with the executable bit set if the file is an executable.
func ModeFrom(fname string, mode fs.FileMode) fs.FileMode {
	if IsDefinitelyNotExec(fname) {
		return mode
	}
	if IsExec(fname, mode) {
		return mode | 0111
	}
	return mode
}

// GetRename attempts to guess what to rename 'file' to for an appropriate executable name.
func GetRename(file string, nameguess string) string {
	if IsDefinitelyNotExec(file) {
		return file
	}

	if strings.HasSuffix(file, ".appimage") {
		// remove the .appimage extension
		return file[:len(file)-len(".appimage")]
	}

	if strings.HasSuffix(file, ".exe") {
		// directly use xxx.exe
		return file
	}

	// otherwise use the rename guess
	return nameguess
}

// SetWhen returns the newValue if condition is true, otherwise it returns the original value.
// use generics like <T interface{}> to make this function more flexible
func SetIf[T interface{}](condition bool, original T, newValue T) T {
	if condition {
		return newValue
	}
	return original
}

func FatalHandler(err error) error {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
	return err
}

func GetCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	return dir
}

// check if two interfaces implement the same type
func SameImplementedInterface(a, b any) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}

// IsErrorOf returns true if err is not nil and is of the type errorCompare.
func IsErrorOf(err error, errorCompare error) bool {
	return err != nil && errors.Is(err, errorCompare)
}

// IsErrorNotOf returns true if err is not nil and is not of the type errorCompare.
func IsErrorNotOf(err error, errorCompare error) bool {
	return err != nil && !errors.Is(err, errorCompare)
}

func ExtractToolNameFromURL(url string) string {
	// Regular expression to find the tool name pattern.
	// This regex assumes that the tool name is followed by a version number and/or OS/architecture information.
	re := regexp.MustCompile(`([^\/]+)-?v?[\d\.]*-?([\w]*-?[\w]*\.\w+)?`)

	url = strings.TrimSuffix(url, "/")
	if strings.Contains(url, "://") {
		url = url[strings.LastIndex(url, "://")+3:]
	}

	// Extracting the basename of the URL
	base := strings.TrimSuffix(url, "/")
	base = base[strings.LastIndex(base, "/")+1:]

	// Using regex to find matches
	matches := re.FindStringSubmatch(base)

	if !strings.ContainsRune(base, '.') && len(matches) > 1 {
		re2 := regexp.MustCompile(`^[a-zA-Z0-9-]+`)
		m2 := re2.FindStringSubmatch(matches[1])
		foundName := strings.Trim(m2[0], "-")
		return foundName
	}

	return "Unknown"
}
