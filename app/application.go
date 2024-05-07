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
	. "github.com/permafrost-dev/eget/lib/appflags"
	. "github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/data"
	"github.com/permafrost-dev/eget/lib/detectors"
	"github.com/permafrost-dev/eget/lib/download"
	. "github.com/permafrost-dev/eget/lib/extraction"
	"github.com/permafrost-dev/eget/lib/finders"
	"github.com/permafrost-dev/eget/lib/github"
	. "github.com/permafrost-dev/eget/lib/globals"
	"github.com/permafrost-dev/eget/lib/utilities"
	. "github.com/permafrost-dev/eget/lib/utilities"
	"github.com/permafrost-dev/eget/lib/verifiers"
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

	result := &Application{
		Opts:       Flags{},
		Output:     nil,
		Cache:      *data.NewCache("./eget.db.json"),
		Outputs:    outputs,
		Filesystem: vfs.OSFS,
	}

	result.initOutputs()

	return result
}

func (app *Application) DownloadClient() *download.Client {
	token, _ := getGithubToken()
	return download.NewClient(token)
}

func (app *Application) Run() *ReturnStatus {
	app.Cache.LoadFromFile()

	target, _ := app.ProcessFlags(FatalHandler)
	target = app.ProcessCommands(target)

	if err := app.targetToProject(target); err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	finder := app.getFinder()
	findResult := app.getFindResult(finder)
	// app.cacheTarget(&finder, findResult)

	if shouldReturn, returnStatus := app.shouldReturn(findResult.Error); shouldReturn {
		return returnStatus
	}

	assetWrapper := NewAssetWrapper(findResult.Assets)
	detector, err := detectors.DetermineCorrectDetector(&app.Opts, nil)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	// get the url and candidates from the detector
	asset, candidates, err := detector.Detect(assetWrapper.Assets)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	if len(candidates) != 0 {
		asset = app.selectFromMultipleAssets(candidates, err) // manually select which asset to download
	}

	assetWrapper.Asset = &asset

	app.writeLine(assetWrapper.Asset.DownloadURL) // print the URL

	body := app.downloadAsset(assetWrapper.Asset, &findResult) // download with progress bar and get the response body
	app.VerifyChecksums(assetWrapper, body)

	extractor, err := app.getExtractor(assetWrapper.Asset, finder.Tool)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	bin, bins, err := extractor.Extract(body, app.Opts.All) // get extraction candidates
	// if err != nil && len(bins) == 0 {
	// 	return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	// }
	if err != nil && len(bins) != 0 && !app.Opts.All {
		bin = app.selectFromMultipleCandidates(bin, bins, err)
	}

	extractedCount := app.ExtractBins(bin, app.wrapBins(bins, bin), app.Opts.All)

	return NewReturnStatus(Success, nil, fmt.Sprintf("extracted files: %d", extractedCount))
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
		return true, NewReturnStatus(Success, nil, fmt.Sprintf("%s: %v", app.Target, err))
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
		time.Now().Add(time.Hour*1),
	)
}

func (app *Application) targetToProject(target string) error {
	app.Target = target
	app.TargetFound = false

	if !IsValidRepositoryReference(app.Target) {
		return fmt.Errorf("invalid GitHub repository URL %s", app.Target)
	}

	var err error

	//app.Target, app.TargetFound = RepositoryNameFromGithubURL(app.Target)
	app.Reference, err = ParseRepositoryReference(app.Target)

	if err != nil {
		return fmt.Errorf("invalid GitHub repository reference  %s: %w", app.Target, err)
	}

	// if !app.TargetFound {
	// 	return fmt.Errorf("GitHub repository not found: '%s'", app.Target)
	// }

	return nil
}

// if multiple candidates are returned, the user must select manually which one to download
func (app *Application) selectFromMultipleAssets(candidates []Asset, err error) Asset {
	// if app.cli.NoInteraction || app.Opts.NoInteraction {
	// 	//Fatal("error: multiple candidates found, cannot select automatically (user interaction disabled)")
	// }

	app.writeErrorLine("%v: please select manually", err)
	choices := make([]interface{}, len(candidates))

	for i := range candidates {
		choices[i] = path.Base(candidates[i].Name)
	}

	choice := app.userSelect(choices)

	return candidates[choice-1]
}

// if there are multiple candidates, have the user select manually
func (app *Application) selectFromMultipleCandidates(bin ExtractedFile, bins []ExtractedFile, err error) ExtractedFile {
	// if app.cli.NoInteraction || app.Opts.NoInteraction {
	// 	//Fatal("error: multiple assets found, cannot prompt user for selection (user interaction disabled)")
	// }

	app.writeErrorLine("%v: please select manually", err)

	choices := make([]interface{}, len(bins)+1)
	for i := range bins {
		choices[i] = bins[i]
	}

	choices[len(bins)] = "all"
	choice := app.userSelect(choices)

	if choice == len(bins)+1 {
		app.Opts.All = true
	} else {
		return bins[choice-1]
	}

	return bin
}

type ProcessFlagsErrorHandlerFunc = func(err error) error

func (app *Application) ProcessCommands(target string) string {
	switch target {
	case "upgrade":
		app.writeLine("upgrading to the latest version of " + ApplicationName + "...")
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
		os.Exit(0)
	}

	if app.cli.Help {
		app.flagparser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	app.initializeConfig()

	if err := app.SetGlobalOptionsFromConfig(); err != nil {
		errorHandler(err)
		return "", err
	}

	if app.cli.Rate {
		rdat, err := github.FetchRateLimit(app.DownloadClient())
		app.Cache.SetRateLimit(
			rdat.Limit,
			rdat.Remaining,
			time.Unix(rdat.Reset, 0).Local(),
		)
		FatalIf(err)
		fmt.Println(rdat)
		os.Exit(0)
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
		ConditionalExit(err)
	}

	if len(app.Args) <= 0 {
		app.writeLine("no target given")
		app.flagparser.WriteHelp(os.Stdout)
		SuccessExit()
	}

	if app.Opts.DisableSSL {
		app.writeErrorLine("warning: SSL verification is disabled")
	}

	if app.Opts.Remove {
		ebin := os.Getenv("EGET_BIN")
		searchPath := SetIf(ebin == "", ebin, app.Config.Global.Target)

		fn := filepath.Join(searchPath, filepath.Base(target))
		if err := os.Remove(fn); err != nil {
			app.writeErrorLine("%s", err.Error())
		} else {
			app.writeLine("Removed `%s`", fn)
		}
	}

	if app.cli.NoInteraction {
		app.Opts.NoInteraction = true
	} else {
		app.Opts.NoInteraction = false
	}

	return target, nil
}

func (app *Application) downloadAsset(asset *Asset, findResult *finders.FindResult) []byte {
	buf := &bytes.Buffer{}

	repo, _ := app.Cache.AddRepository(asset.Name, "", []string{}, findResult, time.Now().Add(time.Hour*1))
	repo.UpdateCheckedAt()
	app.Cache.SaveToFile()

	if err := app.Download(asset.DownloadURL, buf); err != nil {
		Fatal(fmt.Sprintf("%s (URL: %s)", err, asset.DownloadURL))
	}

	repo.UpdateDownloadedAt(asset.DownloadURL)
	app.Cache.SaveToFile()

	return buf.Bytes()
}

func (app *Application) VerifyChecksums(wrapper *AssetWrapper, body []byte) verifiers.VerifyChecksumResult {
	verifier, sumAsset, err := app.getVerifier(*wrapper.Asset, wrapper.Assets)

	if verifier != nil && !utilities.SameImplementedInterface(verifier, verifiers.NoVerifier{}) {
		app.write("› performing verification for %s...", wrapper.Asset.Name)
	}

	if err != nil {
		app.writeLine("Checksum verification failed, could not create a verifier.")
		return verifiers.VerifyChecksumFailedNoVerifier
	}

	if err = verifier.Verify(body); err != nil {
		app.writeLine("failed, %v", err)
		return verifiers.VerifyChecksumVerificationFailed
	}

	if app.Opts.Verify == "" && sumAsset != (Asset{}) {
		app.writeLine("verified ✔")
		return verifiers.VerifyChecksumSuccess
	}

	if app.Opts.Verify != "" {
		app.writeLine("verified ✔")
		return verifiers.VerifyChecksumSuccess
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

	app.extract(bin)

	return 1

}

func (app *Application) extract(bin ExtractedFile) {
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
		Fatal(err)
	}

	app.writeLine("Extracted `%s` to `%s`", bin.ArchiveName, out)
}

// Determine the appropriate Finder to use. If a.Opts.URL is provided, we use
// a DirectAssetFinder. Otherwise we use a GithubAssetFinder. When a Github
// repo is provided, we assume the repo name is the 'tool' name (for direct
// URLs, the tool name is unknown and remains empty).
func (app *Application) getFinder() finders.ValidFinder {
	project := app.Target

	if IsLocalFile(project) || IsNonGithubURL(project) {
		app.Opts.System = "all"
		found := finders.DirectAssetFinder{URL: project}

		return finders.NewValidFinder(found, "")
	}

	project = app.Reference.Name

	tool := app.Reference.Name

	if app.Opts.Source {
		tag := SetIf(app.Opts.Tag != "", "main", app.Opts.Tag)
		result := finders.GithubSourceFinder{Repo: project, Tag: tag, Tool: tool}

		return finders.NewValidFinder(result, tool)
	}

	tag := SetIf(app.Opts.Tag != "", "latest", fmt.Sprintf("tags/%s", app.Opts.Tag))

	var mint time.Time
	if app.Opts.UpgradeOnly {
		parts := strings.Split(project, "/")
		last := parts[len(parts)-1]
		mint = Bintime(last, app.Opts.Output)
	}

	result := finders.GithubAssetFinder{
		Repo:       app.Reference.Owner + "/" + app.Reference.Name,
		Tag:        tag,
		Prerelease: app.Opts.Prerelease,
		MinTime:    mint,
	}

	return finders.NewValidFinder(result, tool)
}

func (app *Application) getVerifier(asset Asset, assets []Asset) (verifier verifiers.Verifier, sumAsset Asset, err error) {
	sumAsset = Asset{}

	if app.Opts.Verify != "" {
		verifier, err = verifiers.NewSha256Verifier(app.DownloadClient(), app.Opts.Verify)
		if err != nil {
			return nil, Asset{}, fmt.Errorf("create Sha256Verifier: %w", err)
		}
		return verifier, Asset{}, nil
	}

	for _, item := range assets {
		if item.Name == asset.Name+".sha256sum" || item.Name == asset.Name+".sha256" {
			app.writeLine("verification against %s (%s)", item.Name, item.DownloadURL)

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
			app.writeLine("› performing checksum verifications against %s (%s)", item.Name, item.DownloadURL)
			return &verifiers.Sha256SumFileAssetVerifier{Sha256SumAssetURL: item.DownloadURL, BinaryName: binaryName, Client: download.NewClient("")}, item, nil
		}
	}

	if app.Opts.Hash {
		return &verifiers.Sha256Printer{}, Asset{}, nil
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
func (app *Application) userSelect(choices []interface{}) int {
	for i, c := range choices {
		app.writeErrorLine("(%d) %v", i+1, c)
	}

	var choice int

	for {
		app.writeError("Enter selection number: ")
		_, err := fmt.Scanf("%d", &choice)
		if err == nil && (choice <= 0 || choice > len(choices)) {
			err = fmt.Errorf("%d is out of bounds", choice)
		}
		if err == nil {
			break
		}
		if errors.Is(err, io.EOF) {
			Fatal("Error reading selection")
		}

		app.writeErrorLine("Invalid selection: %v", err)
	}

	return choice
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
