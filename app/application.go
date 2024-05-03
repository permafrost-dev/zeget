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
)

type Application struct {
	Output     io.Writer
	Opts       Flags
	cli        CliFlags
	Args       []string
	flagparser *flags.Parser
	Config     *Config
}

func NewApplication() *Application {
	result := &Application{
		Opts:   Flags{},
		Output: nil,
	}

	result.initOutputWriter()

	return result
}

func (app *Application) Run() error {
	target, err := app.processFlags()

	if err != nil {
		Fatal(err)
	}

	finder, tool := app.getFinder(target)
	assets, err := finder.Find()
	FatalIf(err)

	if err != nil && errors.Is(err, ErrNoUpgrade) {
		app.writeLine("%s: %v", target, err)
		SuccessExit()
	}

	detector, err := DetermineCorrectDetector(&app.Opts)
	FatalIf(err)

	// get the url and candidates from the detector
	asset, candidates, err := detector.Detect(assets)
	FatalIf(err)

	if len(candidates) != 0 && err != nil {
		// if multiple candidates are returned, the user must select manually which one to download
		app.writeErrorLine("%v: please select manually", err)
		choices := make([]interface{}, len(candidates))
		for i := range candidates {
			choices[i] = path.Base(candidates[i].Name)
		}
		choice := app.userSelect(choices)
		asset = candidates[choice-1]
	}

	app.writeLine(asset.DownloadURL) // print the URL
	// download with progress bar and get the response body
	body := app.downloadAsset(&asset)
	app.verifyChecksums(asset, assets, body)

	extractor, err := app.getExtractor(asset.DownloadURL, tool)
	FatalIf(err)

	// get extraction candidates
	bin, bins, err := extractor.Extract(body, app.Opts.All)

	if err != nil && len(bins) == 0 {
		Fatal(err)
	}

	if len(bins) != 0 && err != nil && !app.Opts.All {
		// if there are multiple candidates, have the user select manually
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
			bin = bins[choice-1]
		}
	}

	if app.Opts.All {
		if len(bins) == 0 {
			bins = []ExtractedFile{bin}
		}

		for _, bin := range bins {
			app.extract(bin)
		}
	}

	if !app.Opts.All {
		app.extract(bin)
	}

	return nil
}

func (app *Application) processFlags() (string, error) {
	app.flagparser = flags.NewParser(&app.cli, flags.PassDoubleDash|flags.PrintErrors)
	app.flagparser.Usage = "[OPTIONS] TARGET"

	args, err := app.flagparser.Parse()
	if err != nil {
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

	if err := SetGlobalOptionsFromConfig(app.Config, app.flagparser, &app.Opts, app.cli); err != nil {
		return "", err
	}

	if app.cli.Rate {
		rdat, err := GetRateLimit()
		if err != nil {
			Fatal(err)
		}
		fmt.Println(rdat)
		os.Exit(0)
	}

	target := ""

	if len(app.Args) > 0 {
		target = app.Args[0]
	}

	if err := SetProjectOptionsFromConfig(app.Config, app.flagparser, &app.Opts, app.cli, target); err != nil {
		Fatal(err)
	}

	if app.cli.DownloadAll {
		err := app.downloadConfigRepositories()
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
	if err := app.Download(asset.DownloadURL, buf); err != nil {
		Fatal(fmt.Sprintf("%s (URL: %s)", err, asset.DownloadURL))
	}

	return buf.Bytes()
}

func (app *Application) verifyChecksums(asset Asset, assets []Asset, body []byte) {
	verifier, sumAsset, err := app.getVerifier(asset, assets)

	if err != nil {
		app.writeLine("Checksum verification failed, could not create a verifier.")
		return
	}

	if err = verifier.Verify(body); err != nil {
		app.writeLine("Checksum verification failed, %v", err)
	}

	if app.Opts.Verify == "" && sumAsset != (Asset{}) {
		app.writeLine("Checksum verified with %s", path.Base(sumAsset.Name))
	}

	if app.Opts.Verify != "" {
		app.writeLine("Checksum verified")
	}
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

	fmt.Fprintf(app.Output, "Extracted `%s` to `%s`\n", bin.ArchiveName, out)
}

// Determine the appropriate Finder to use. If a.Opts.URL is provided, we use
// a DirectAssetFinder. Otherwise we use a GithubAssetFinder. When a Github
// repo is provided, we assume the repo name is the 'tool' name (for direct
// URLs, the tool name is unknown and remains empty).
func (app *Application) getFinder(project string) (finder Finder, tool string) {
	if IsLocalFile(project) || IsNonGithubUrl(project) {
		app.Opts.System = "all"
		return &DirectAssetFinder{URL: project}, tool
	}

	if IsInvalidGithubUrl(project) {
		Fatal(fmt.Sprintf("invalid GitHub repository URL %s", project))
	}

	if IsGithubUrl(project) {
		project, _ = RepositoryNameFromGithubUrl(project)
	}

	if !IsValidRepositoryReference(project) {
		Fatal("invalid argument (must be of the form `user/repo`)")
	}

	tool = ParseRepositoryReference(project).Name

	if app.Opts.Source {
		tag := SetIf(app.Opts.Tag != "", "main", app.Opts.Tag)
		return &GithubSourceFinder{Repo: project, Tag: tag, Tool: tool}, tool
	}

	tag := SetIf(app.Opts.Tag != "", "latest", fmt.Sprintf("tags/%s", app.Opts.Tag))

	var mint time.Time
	if app.Opts.UpgradeOnly {
		parts := strings.Split(project, "/")
		last := parts[len(parts)-1]
		mint = Bintime(last, app.Opts.Output)
	}

	return &GithubAssetFinder{
		Repo:       project,
		Tag:        tag,
		Prerelease: app.Opts.Prerelease,
		MinTime:    mint,
	}, tool
}

func (app *Application) getVerifier(asset Asset, assets []Asset) (verifier Verifier, sumAsset Asset, err error) {
	sumAsset = Asset{}

	if app.Opts.Verify != "" {
		verifier, err = NewSha256Verifier(app.Opts.Verify)
		if err != nil {
			return nil, Asset{}, fmt.Errorf("create Sha256Verifier: %w", err)
		}
		return verifier, Asset{}, nil
	}

	for _, item := range assets {
		if item.Name == asset.Name+".sha256sum" || item.Name == asset.Name+".sha256" {
			app.writeLine("verify against %s", item)
			return &Sha256AssetVerifier{AssetURL: item.DownloadURL}, item, nil
		}
		if strings.Contains(item.Name, "checksum") {
			binaryUrl, err := url.Parse(asset.DownloadURL)
			if err != nil {
				return nil, item, fmt.Errorf("extract binary name from asset url: %s: %w", asset, err)
			}
			binaryName := path.Base(binaryUrl.Path)
			app.writeLine("verify against %s", item)
			return &Sha256SumFileAssetVerifier{Sha256SumAssetURL: item.DownloadURL, BinaryName: binaryName}, item, nil
		}
	}

	if app.Opts.Hash {
		return &Sha256Printer{}, Asset{}, nil
	}

	return &NoVerifier{}, Asset{}, nil
}

// Determine which extractor to use. If --download-only is provided, we
// just "extract" the downloaded archive to itself. Otherwise we try to
// extract the literal file provided by --file, or by default we just
// extract a binary with the tool name that was possibly auto-detected
// above.
func (a *Application) getExtractor(url, tool string) (extractor Extractor, err error) {
	if a.Opts.DLOnly {
		return &SingleFileExtractor{
			Name:   path.Base(url),
			Rename: path.Base(url),
			Decompress: func(r io.Reader) (io.Reader, error) {
				return r, nil
			},
		}, nil
	}

	if a.Opts.ExtractFile != "" {
		gc, err := NewGlobChooser(a.Opts.ExtractFile)
		if err != nil {
			return nil, err
		}
		return NewExtractor(path.Base(url), tool, gc), nil
	}

	return NewExtractor(path.Base(url), tool, &BinaryChooser{Tool: tool}), nil
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

	for name, _ := range app.Config.Repositories {
		cmd := exec.Command(binary, name)
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			hasError = true
			errorList = append(errorList, err)
		}
	}

	if hasError {
		return fmt.Errorf("one or more errors occurred while downloading: %v", errorList)
	}

	return nil
}
