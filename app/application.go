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
	. "github.com/permafrost-dev/eget/lib/assets"
	"github.com/permafrost-dev/eget/lib/data"
	"github.com/permafrost-dev/eget/lib/download"
	"github.com/permafrost-dev/eget/lib/finders"
	"github.com/permafrost-dev/eget/lib/verifiers"
)

type Application struct {
	Output     io.Writer
	Opts       Flags
	cli        CliFlags
	Args       []string
	flagparser *flags.Parser
	Config     *Config
	Cache      data.Cache
}

func NewApplication() *Application {
	result := &Application{
		Opts:   Flags{},
		Output: nil,
		Cache:  *data.NewCache("./eget.db.json"),
	}

	result.initOutputWriter()

	return result
}

func (app *Application) DownloadClient() *download.Client {
	token, _ := getGithubToken()
	return download.NewClient(token)
}

func (app *Application) Run() {
	target, _ := app.ProcessFlags(FatalHandler)
	target = app.ProcessCommands(target)

	app.Cache.LoadFromFile()
	app.Cache.AddRepository(target, "", []string{}, time.Now().Add(time.Hour*1))

	finder, tool := app.getFinder(target)
	assets, err := finder.Find(app.DownloadClient())

	if err != nil && errors.Is(err, finders.ErrNoUpgrade) {
		app.writeLine("%s: %v", target, err)
		SuccessExit()
	}

	FatalIf(err)

	assetWrapper := NewAssetWrapper(assets)
	detector, err := DetermineCorrectDetector(&app.Opts, nil)
	FatalIf(err)

	// get the url and candidates from the detector
	asset, candidates, err := detector.Detect(assetWrapper.Assets)
	FatalIf(err)

	if len(candidates) != 0 {
		asset = app.selectFromMultipleAssets(candidates, err) // manually select which asset to download
	}

	assetWrapper.Asset = &asset

	app.writeLine(assetWrapper.Asset.DownloadURL) // print the URL
	body := app.downloadAsset(assetWrapper.Asset) // download with progress bar and get the response body
	app.VerifyChecksums(assetWrapper, body)

	fmt.Printf("cache %v\n", app.Cache.Data)

	extractor, err := app.getExtractor(assetWrapper.Asset, tool)
	FatalIf(err)

	bin, bins, err := extractor.Extract(body, app.Opts.All) // get extraction candidates
	if err != nil && len(bins) == 0 {
		Fatal(err)
	}

	if err != nil && len(bins) != 0 && !app.Opts.All {
		bin = app.selectFromMultipleCandidates(bin, bins, err)
	}

	if app.Opts.All {
		if len(bins) == 0 {
			bins = []ExtractedFile{bin}
		}

		for _, bin := range bins {
			app.extract(bin)
		}

		return
	}

	app.extract(bin)
}

// if multiple candidates are returned, the user must select manually which one to download
func (app *Application) selectFromMultipleAssets(candidates []Asset, err error) Asset {
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
		rdat, err := app.GetRateLimit()
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

		if err := os.Remove(filepath.Join(ebin, target)); err != nil {
			app.writeErrorLine("%s", err.Error())
			os.Exit(1)
		}

		app.writeLine("Removed `%s`", filepath.Join(ebin, target))
		SuccessExit()
	}

	return target, nil
}

func (app *Application) downloadAsset(asset *Asset) []byte {
	buf := &bytes.Buffer{}

	repo, _ := app.Cache.AddRepository(asset.Name, "", []string{}, time.Now().Add(time.Hour*1))
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

	if err != nil {
		app.writeLine("Checksum verification failed, could not create a verifier.")
		return verifiers.VerifyChecksumFailedNoVerifier
	}

	if err = verifier.Verify(body); err != nil {
		app.writeLine("Checksum verification failed, %v", err)
		return verifiers.VerifyChecksumVerificationFailed
	}

	if app.Opts.Verify == "" && sumAsset != (Asset{}) {
		app.writeLine("Checksum verified with %s", path.Base(sumAsset.Name))
		return verifiers.VerifyChecksumSuccess
	}

	if app.Opts.Verify != "" {
		app.writeLine("Checksum verified")
		return verifiers.VerifyChecksumSuccess
	}

	return verifiers.VerifyChecksumNone
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
func (app *Application) getFinder(project string) (finder finders.Finder, tool string) {
	if IsLocalFile(project) || IsNonGithubURL(project) {
		app.Opts.System = "all"
		return &finders.DirectAssetFinder{URL: project}, tool
	}

	if IsInvalidGithubURL(project) {
		Fatal(fmt.Sprintf("invalid GitHub repository URL %s", project))
	}

	if IsGithubURL(project) {
		project, _ = RepositoryNameFromGithubURL(project)
	}

	if !IsValidRepositoryReference(project) {
		Fatal("invalid argument (must be of the form `user/repo`)")
	}

	tool = ParseRepositoryReference(project).Name

	if app.Opts.Source {
		tag := SetIf(app.Opts.Tag != "", "main", app.Opts.Tag)
		return &finders.GithubSourceFinder{Repo: project, Tag: tag, Tool: tool}, tool
	}

	tag := SetIf(app.Opts.Tag != "", "latest", fmt.Sprintf("tags/%s", app.Opts.Tag))

	var mint time.Time
	if app.Opts.UpgradeOnly {
		parts := strings.Split(project, "/")
		last := parts[len(parts)-1]
		mint = Bintime(last, app.Opts.Output)
	}

	return &finders.GithubAssetFinder{
		Repo:       project,
		Tag:        tag,
		Prerelease: app.Opts.Prerelease,
		MinTime:    mint,
	}, tool
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
			app.writeLine("verify against %s", item)

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
			app.writeLine("verify against %s", item)
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
		return NewExtractor(path.Base(asset.DownloadURL), tool, gc), nil
	}

	return NewExtractor(path.Base(asset.DownloadURL), tool, &BinaryChooser{Tool: tool}), nil
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
