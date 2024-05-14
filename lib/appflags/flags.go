package appflags

import "github.com/permafrost-dev/eget/lib/filters"

type Flags struct {
	Tag           string
	Prerelease    bool
	Source        bool
	Output        string
	System        string
	ExtractFile   string
	All           bool
	Quiet         bool
	DLOnly        bool
	UpgradeOnly   bool
	Asset         []string
	Sha256        bool
	Hash          bool
	Verify        string
	Remove        bool
	DisableSSL    bool
	NoInteraction bool
	Verbose       bool
	NoProgress    bool
	Filters       []*filters.Filter
}

type CliFlags struct {
	Tag           *string   `short:"t" long:"tag" description:"tagged release to use instead of latest"`
	Prerelease    *bool     `long:"pre-release" description:"include pre-releases when fetching the latest version"`
	Source        *bool     `long:"source" description:"download the source code for the target repo instead of a release"`
	Output        *string   `long:"to" description:"move to given location after extracting"`
	System        *string   `short:"s" long:"system" description:"target system to download for (use \"all\" for all choices)"`
	ExtractFile   *string   `short:"f" long:"file" description:"glob to select files for extraction"`
	All           *bool     `long:"all" description:"extract all candidate files"`
	Quiet         *bool     `short:"q" long:"quiet" description:"only print essential output"`
	DLOnly        *bool     `short:"d" long:"download-only" description:"stop after downloading the asset (no extraction)"`
	UpgradeOnly   *bool     `long:"upgrade-only" description:"only download if release is more recent than current version"`
	Asset         *[]string `short:"a" long:"asset" description:"download a specific asset containing the given string; can be specified multiple times for additional filtering; use ^ for anti-match"`
	Hash          *bool     `short:"H" long:"hash" description:"show the SHA-256 hash of the downloaded asset"`
	Sha256        *bool     `long:"sha256" description:"show the SHA-256 hash of the downloaded asset"`
	Verify        *string   `long:"verify-sha256" description:"verify the downloaded asset checksum against the one provided"`
	Remove        *bool     `short:"r" long:"remove" description:"remove the given file from $EGET_BIN or the current directory"`
	Version       bool      `short:"V" long:"version" description:"show version information"`
	Help          bool      `short:"h" long:"help" description:"show this help message"`
	DownloadAll   bool      `short:"D" long:"download-all" description:"download all projects defined in the config file"`
	DisableSSL    *bool     `short:"k" long:"disable-ssl" description:"disable SSL verification for download requests"`
	NoInteraction bool      `long:"no-interaction" description:"do not prompt for user input"`
	Verbose       *bool     `short:"v" long:"verbose" description:"show verbose output"`
	NoProgress    *bool     `long:"no-progress" description:"do not show download progress"`
	Filters       *string   `short:"F" long:"filter" description:"filter assets using functions like 'all', 'any', 'none', 'has', 'ext'"`
}
