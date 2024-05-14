package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	. "github.com/permafrost-dev/zeget/lib/appflags"
	. "github.com/permafrost-dev/zeget/lib/assets"
	"github.com/permafrost-dev/zeget/lib/data"
	"github.com/permafrost-dev/zeget/lib/download"
	. "github.com/permafrost-dev/zeget/lib/extraction"
	"github.com/permafrost-dev/zeget/lib/filters"
	"github.com/permafrost-dev/zeget/lib/finders"
	"github.com/permafrost-dev/zeget/lib/github"
	. "github.com/permafrost-dev/zeget/lib/globals"
	"github.com/permafrost-dev/zeget/lib/home"
	"github.com/permafrost-dev/zeget/lib/utilities"
	. "github.com/permafrost-dev/zeget/lib/utilities"
	"github.com/permafrost-dev/zeget/lib/verifiers"
	"github.com/twpayne/go-vfs/v5"
)

type ApplicationOutputs struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Discard io.Writer
}

type Application struct {
	Output      io.Writer
	Outputs     *ApplicationOutputs
	Opts        Flags
	cli         CliFlags
	Args        []string
	flagparser  *flags.Parser
	Config      *Config
	Cache       data.Cache
	Filesystem  vfs.FS
	Reference   *RepositoryReference
	Target      string
	TargetFound bool
}

var ErrNoTargetGiven = errors.New("no target given")
var ErrSuccess = errors.New("success")

func NewApplicationOutputs(stdout io.Writer, stderr io.Writer) *ApplicationOutputs {
	if stdout == nil {
		stdout = os.Stdout
	}

	if stderr == nil {
		stderr = os.Stderr
	}

	return &ApplicationOutputs{
		Stdout:  stdout,
		Stderr:  stderr,
		Discard: io.Discard,
	}
}

func NewApplication(outputs *ApplicationOutputs) *Application {
	if outputs == nil {
		outputs = NewApplicationOutputs(nil, nil)
	}

	vf := vfs.OSFS

	result := &Application{
		Opts:       Flags{},
		Output:     nil,
		Cache:      *data.NewCache(GetCacheFilename(vf)),
		Outputs:    outputs,
		Filesystem: vf,
	}

	result.initOutputs()

	return result
}

func (app *Application) ToolName() string {
	return app.Reference.Name
}

func (app *Application) DownloadClient() *download.Client {
	token, _ := getGithubToken()
	return download.NewClient(token)
}

func (app *Application) RunSetup(_ ProcessFlagsErrorHandlerFunc) (string, *ReturnStatus) {
	var err error
	var target string

	if target, err = app.ProcessFlags(FatalHandler); err != nil {
		if errors.Is(err, ErrNoTargetGiven) || errors.Is(err, ErrSuccess) {
			return "", NewReturnStatus(Success, nil, "")
		}

		return "", NewReturnStatus(FatalError, err, fmt.Sprintf("run setup error: %v", err))
	}

	return app.ProcessCommands(target), nil
}

func (app *Application) wrapBins(bins []ExtractedFile, bin ExtractedFile) []ExtractedFile {
	if len(bins) == 0 {
		return []ExtractedFile{bin}
	}
	return bins
}

func (app *Application) shouldReturn(err error) (bool, *ReturnStatus) {

	// remote asset is not a newer version than the current version
	if IsErrorOf(err, finders.ErrNoUpgrade) {
		return true, NewReturnStatus(Success, finders.ErrNoUpgrade, fmt.Sprintf("%s: %v", app.Target, err))
	}

	// some other error occurred
	if IsErrorNotOf(err, finders.ErrNoUpgrade) {
		return true, NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	return false, nil
}

func (app *Application) getFindResult(finder finders.ValidFinder) finders.FindResult {
	if finder.Finder == nil {
		return finders.FindResult{Error: fmt.Errorf("finder is nil")}
	}

	return *finder.Finder.Find(app.DownloadClient())
}

func (app *Application) cacheTarget(finding *finders.ValidFinder, findResult *finders.FindResult) {
	app.Cache.AddRepository(
		app.Target,
		finding.Tool,
		app.Opts.Asset,
		findResult,
		time.Now().Add(time.Hour*24*7),
	)
}

func (app *Application) targetToProject(target string) error {
	var err error

	app.Target = target
	app.TargetFound = false

	if app.Reference, err = ParseRepositoryReference(app.Target); err != nil {
		return err
	}

	return nil
}

// if multiple candidates are returned, the user must select manually which one to download
func (app *Application) selectFromMultipleAssets(candidates []Asset, err error) (Asset, error) {
	if app.cli.NoInteraction || app.Opts.NoInteraction {
		return Asset{}, fmt.Errorf("error: multiple candidates found, cannot select automatically (user interaction disabled)")
	}

	app.WriteErrorLine("%v: please select manually", err)
	choices := make([]interface{}, len(candidates))

	for i := range candidates {
		choices[i] = path.Base(candidates[i].Name)
	}

	choice, err := app.userSelect(choices)
	if err != nil {
		return Asset{}, fmt.Errorf("error: %v", err)
	}

	choiceStr := fmt.Sprintf("%s", choices[choice-1])

	result := candidates[choice-1]
	result.Filters = utilities.FilenameToAssetFilters(choiceStr)

	return result, nil

}

// if there are multiple candidates, have the user select manually
func (app *Application) selectFromMultipleCandidates(bin ExtractedFile, bins []ExtractedFile, err error) (ExtractedFile, error) {
	if app.cli.NoInteraction || app.Opts.NoInteraction {
		return ExtractedFile{}, fmt.Errorf("error: multiple assets found, cannot prompt user for selection (user interaction disabled)")
	}

	app.WriteErrorLine("%v: please select manually", err)

	choices := make([]interface{}, len(bins)+1)
	for i := range bins {
		choices[i] = bins[i]
	}

	choices[len(bins)] = "all"
	choice, err := app.userSelect(choices)

	if err != nil {
		return ExtractedFile{}, err
	}

	if choice == len(bins)+1 {
		app.Opts.All = true
	} else {
		return bins[choice-1], nil
	}

	return bin, nil
}

func (app *Application) RateLimitExceeded() error {
	if app.Cache.Data.RateLimit.Remaining < 10 {
		return errors.New("GitHub rate limit exceeded")
	}

	return nil
}

func (app *Application) RefreshRateLimit() error {
	//diff in minutes between the current time and the rate limit reset time:
	diff := app.Cache.Data.RateLimit.Reset.Sub(time.Now().Local()).Round(time.Second).Seconds() / 60

	// not a percentage per se, but the rate limit remaining divided by the number of minutes until the rate limit resets
	// provides a rough estimate of the rate of requests that can be made per minute until the rate limit resets.
	rate := int64(app.Cache.Data.RateLimit.Remaining) / int64(diff)

	// fmt.Printf("Rate limit reset in %v ,minutes\n", diff)
	// fmt.Printf("Remaining rate limit: %d\n", app.Cache.Data.RateLimit.Remaining)
	// fmt.Printf("Total rate limit: %d\n", rate)

	// if rate is <50, always refresh the rate limit
	if app.Cache.Data.RateLimit.Reset != nil && rate > 50 {
		if app.Cache.Data.RateLimit.Reset.After(time.Now().Local()) {
			return errors.New("rate limit has not reset")
		}
	}

	rateLimitDate, err := github.FetchRateLimit(app.DownloadClient())

	if err == nil {
		app.Cache.SetRateLimit(
			rateLimitDate.Limit,
			rateLimitDate.Remaining,
			rateLimitDate.ResetsAt.Local(),
		)
	}

	return err
}

type ProcessFlagsErrorHandlerFunc = func(err error) error

func (app *Application) ProcessCommands(target string) string {
	switch target {
	case "upgrade":
		app.WriteLine("upgrading to the latest version of " + ApplicationName + "...")
		return ApplicationRepository
	default:
		return target
	}
}

func (app *Application) ProcessFlags(errorHandler ProcessFlagsErrorHandlerFunc) (string, error) {
	app.flagparser = flags.NewParser(&app.cli, flags.PassDoubleDash|flags.PrintErrors)
	app.flagparser.Usage = "[OPTIONS] TARGET"

	args, err := app.flagparser.Parse()
	if err != nil {
		errorHandler(err)
		return "", err
	}
	app.Args = args

	if app.cli.Version {
		fmt.Println("eget version", Version)
		return "", ErrSuccess
	}

	if app.cli.Help {
		app.flagparser.WriteHelp(os.Stdout)
		return "", ErrSuccess
	}

	app.initializeConfig()

	if err := app.SetGlobalOptionsFromConfig(); err != nil {
		errorHandler(err)
		return "", err
	}

	target := ""

	if len(app.Args) > 0 {
		target = app.Args[0]
	}

	if err := app.SetProjectOptionsFromConfig(target); err != nil {
		errorHandler(err)
		return "", err
	}

	if app.cli.DownloadAll {
		if err := app.downloadConfigRepositories(); err != nil {
			errorHandler(err)
			return "", err
		}
		return "", ErrSuccess
	}

	if len(app.Args) <= 0 {
		app.WriteLine("no target given")
		app.flagparser.WriteHelp(os.Stdout)
		return "", ErrNoTargetGiven
	}

	if app.Opts.DisableSSL {
		app.WriteErrorLine("warning: SSL verification is disabled")
	}

	if app.Opts.Remove {
		ebin := os.Getenv("EGET_BIN")
		searchPath := SetIf(ebin == "", ebin, app.Config.Global.Target)

		fn := filepath.Join(searchPath, filepath.Base(target))
		if err := os.Remove(fn); err != nil {
			app.WriteErrorLine("%s", err.Error())
		} else {
			app.WriteLine("Removed `%s`", home.NewPathCompactor().Compact(fn))
		}
	}

	if app.cli.NoInteraction {
		app.Opts.NoInteraction = true
	} else {
		app.Opts.NoInteraction = false
	}

	app.Opts.Verbose = app.cli.Verbose != nil && *app.cli.Verbose
	app.Opts.NoProgress = app.cli.NoProgress != nil && *app.cli.NoProgress

	app.Opts.Filters = []*filters.Filter{}
	if app.cli.Filters != nil {
		app.Opts.Filters = filters.NewParser().ParseDefinitions(*app.cli.Filters)
	}

	return target, nil
}

func (app *Application) downloadAsset(asset *Asset, findResult *finders.FindResult) ([]byte, error) {
	buf := &bytes.Buffer{}

	repo, _ := app.Cache.AddRepository(asset.Name, "", []string{}, findResult, time.Now().Add(time.Hour*1))
	repo.UpdateCheckedAt()

	if err := app.Download(asset.DownloadURL, buf); err != nil {
		return []byte{}, fmt.Errorf("%s (URL: %s)", err, asset.DownloadURL)
	}

	repo.UpdateDownloadedAt(asset.DownloadURL)

	return buf.Bytes(), nil
}

func (app *Application) VerifyChecksums(wrapper *AssetWrapper, body []byte) verifiers.VerifyChecksumResult {
	verifier, sumAsset, err := app.getVerifier(*wrapper.Asset, wrapper.Assets)
	needsNewLine := false

	if verifier != nil && verifier.String() != (verifiers.NoVerifier{}).String() {
		app.Write("› performing verification for %s...", wrapper.Asset.Name)
		needsNewLine = true
	}

	if err != nil {
		app.WriteLine("failed, could not create a verifier.")
		return verifiers.VerifyChecksumFailedNoVerifier
	}

	if err = verifier.Verify(body); err != nil {
		app.WriteLine("failed, %v", err)
		return verifiers.VerifyChecksumVerificationFailed
	}

	if app.Opts.Verify == "" && sumAsset.Name != "" {
		app.WriteLine("verified ✔")
		return verifiers.VerifyChecksumSuccess
	}

	if app.Opts.Verify != "" {
		app.WriteLine("verified ✔")
		return verifiers.VerifyChecksumSuccess
	}

	if needsNewLine {
		app.WriteLine("skipped")
	}

	return verifiers.VerifyChecksumNone
}

func (app *Application) ExtractBins(bin ExtractedFile, bins []ExtractedFile, extractAll bool) int {
	if extractAll {
		for _, bin := range bins {
			app.extract(bin)
		}
		return len(bins)
	}

	if err := app.extract(bin); err != nil {
		app.WriteErrorLine("error: %v", err)
		return 0
	}

	return 1

}

func (app *Application) extract(bin ExtractedFile) error {
	mode := bin.Mode()

	// write the extracted file to a file on disk, in the --to directory if requested
	out := filepath.Base(bin.Name)
	if app.Opts.Output == "-" {
		out = "-"
	}

	if app.Opts.Output != "" && IsDirectory(app.Opts.Output) {
		out = filepath.Join(app.Opts.Output, out)
	}

	if app.Opts.Output != "" && !IsDirectory(app.Opts.Output) && app.Opts.All {
		os.MkdirAll(app.Opts.Output, 0755)
		out = filepath.Join(app.Opts.Output, out)
	}

	out = SetIf(app.Opts.Output != "", app.Opts.Output, out)

	// only use $EGET_BIN if all of the following are true
	// 1. $EGET_BIN is non-empty
	// 2. --to is not a path (not a path if no path separator is found)
	// 3. The extracted file is executable
	if os.Getenv("EGET_BIN") != "" && !strings.ContainsRune(out, os.PathSeparator) && mode&0111 != 0 && !bin.Dir {
		out = filepath.Join(os.Getenv("EGET_BIN"), out)
	}

	if err := bin.Extract(out); err != nil {
		return err
	}

	app.WriteLine("› extracted `%s` to `%s` ✔", bin.ArchiveName, home.NewPathCompactor().Compact(out))

	return nil
}

// Determine the appropriate Finder to use. If a.Opts.URL is provided, we use
// a DirectAssetFinder. Otherwise we use a GithubAssetFinder. When a Github
// repo is provided, we assume the repo name is the 'tool' name (for direct
// URLs, the tool name is unknown and remains empty).
func (app *Application) getFinder() finders.ValidFinder {
	if IsLocalFile(app.Target) || IsNonGithubURL(app.Target) {
		app.Opts.System = "all"
		found := finders.DirectAssetFinder{URL: app.Target}

		toolName := utilities.ExtractToolNameFromURL(found.URL)
		if toolName == "Unknown" {
			toolName = ""
		}

		return finders.NewValidFinder(found, toolName)
	}

	if app.Opts.Source {
		tag := SetIf(app.Opts.Tag != "", "main", app.Opts.Tag)
		result := finders.GithubSourceFinder{Repo: app.Reference.String(), Tag: tag, Tool: app.ToolName()}

		return finders.NewValidFinder(result, app.ToolName())
	}

	tag := SetIf(app.Opts.Tag != "", "latest", fmt.Sprintf("tags/%s", app.Opts.Tag))

	var mint time.Time
	if app.Opts.UpgradeOnly {
		mint = Bintime(app.ToolName(), app.Opts.Output)
	}

	result := finders.NewGithubAssetFinder(app.Reference, tag, app.Opts.Prerelease, mint)

	return finders.NewValidFinder(result, app.ToolName())
}

func (app *Application) getVerifier(asset Asset, assets []Asset) (verifier verifiers.Verifier, _ Asset, err error) {
	// sumAsset = Asset{
	// 	Filters: []string{},
	// }

	if app.Opts.Verify != "" {
		verifier, err = verifiers.NewSha256Verifier(app.DownloadClient(), app.Opts.Verify)
		if err != nil {
			return nil, Asset{}, fmt.Errorf("create Sha256Verifier: %w", err)
		}
		return verifier, Asset{}, nil
	}

	for _, item := range assets {
		if item.Name == asset.Name+".sha256sum" || item.Name == asset.Name+".sha256" {
			app.WriteVerboseLine("verification against %s (%s)", item.Name, item.DownloadURL)

			verifier := verifiers.Sha256AssetVerifier{AssetURL: item.DownloadURL}
			verifier.WithClient(app.DownloadClient())

			return &verifier, item, nil
		}
		if strings.Contains(item.Name, "checksum") {
			binaryURL, err := url.Parse(asset.DownloadURL)
			if err != nil {
				return nil, item, fmt.Errorf("extract binary name from asset url: %s: %w", asset, err)
			}
			binaryName := path.Base(binaryURL.Path)
			app.WriteVerboseLine("› performing checksum verifications against %s (%s)", item.Name, item.DownloadURL)
			return &verifiers.Sha256SumFileAssetVerifier{Sha256SumAssetURL: item.DownloadURL, BinaryName: binaryName, Client: download.NewClient("")}, item, nil
		}
	}

	return &verifiers.NoVerifier{}, Asset{}, nil
}

// Determine which extractor to use. If --download-only is provided, we
// just "extract" the downloaded archive to itself. Otherwise we try to
// extract the literal file provided by --file, or by default we just
// extract a binary with the tool name that was possibly auto-detected
// above.
func (app *Application) getExtractor(asset *Asset, tool string) (extractor Extractor, err error) {
	if app.Opts.DLOnly {
		return &SingleFileExtractor{
			Name:   path.Base(asset.DownloadURL),
			Rename: path.Base(asset.DownloadURL),
			Decompress: func(r io.Reader) (io.Reader, error) {
				return r, nil
			},
		}, nil
	}

	if app.Opts.ExtractFile != "" {
		gc, err := NewGlobChooser(app.Opts.ExtractFile)
		if err != nil {
			return nil, err
		}
		return NewExtractor(app.Filesystem, path.Base(asset.DownloadURL), tool, gc), nil
	}

	return NewExtractor(app.Filesystem, path.Base(asset.DownloadURL), tool, &BinaryChooser{Tool: tool}), nil
}

// Would really like generics to implement this...
// Make the user select one of the choices and return the index of the
// selection.
func (app *Application) userSelect(choices []interface{}) (int, error) {
	for i, c := range choices {
		app.WriteErrorLine("(%d) %v", i+1, c)
	}

	var choice int

	for {
		app.WriteError("Enter selection number: ")
		_, err := fmt.Scanf("%d", &choice)
		if err == nil && (choice <= 0 || choice > len(choices)) {
			err = fmt.Errorf("%d is out of bounds", choice)
		}
		if err == nil {
			break
		}
		if errors.Is(err, io.EOF) {
			return -1, fmt.Errorf("Error reading selection")
		}

		app.WriteErrorLine("Invalid selection: %v", err)
	}

	return choice, nil
}

func (app *Application) downloadConfigRepositories() error {
	hasError := false
	errorList := []error{}

	binary, err := os.Executable()

	if err != nil {
		binary = os.Args[0]
	}

	for name := range app.Config.Repositories {
		cmd := exec.Command(binary, name)
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			hasError = true
			errorList = append(errorList, err)
		}
	}

	if hasError {
		return fmt.Errorf("one or more errors occurred while downloading: %v", errorList)
	}

	return nil
}
