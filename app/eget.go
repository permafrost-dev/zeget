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

// Determine the appropriate Finder to use. If opts.URL is provided, we use
// a DirectAssetFinder. Otherwise we use a GithubAssetFinder. When a Github
// repo is provided, we assume the repo name is the 'tool' name (for direct
// URLs, the tool name is unknown and remains empty).
func getFinder(project string, opts *Flags) (finder Finder, tool string) {
	if IsLocalFile(project) || IsNonGithubUrl(project) {
		opts.System = "all"
		return &DirectAssetFinder{ URL: project }, tool
	}

	if IsInvalidGithubUrl(project) {
		Fatal(fmt.Sprintf("invalid GitHub repository URL %s", project))
	}

	if IsGithubUrl(project) {
		project, _ = RepositoryNameFromGithubUrl(project)
	}

	repo := project
	if !IsValidRepositoryReference(repo) {
		Fatal("invalid argument (must be of the form `user/repo`)")
	}

	tool = ParseRepositoryReference(repo).Name

	if opts.Source {
		tag := SetIf(opts.Tag != "", "main", opts.Tag)

		finder = &GithubSourceFinder{
			Repo: repo,
			Tag:  tag,
			Tool: tool,
		}

		return finder, tool
	}

	tag := SetIf(opts.Tag != "", "latest", fmt.Sprintf("tags/%s", opts.Tag))

	var mint time.Time
	if opts.UpgradeOnly {
		parts := strings.Split(project, "/")
		last := parts[len(parts)-1]
		mint = Bintime(last, opts.Output)
	}

	finder = &GithubAssetFinder{
		Repo:       repo,
		Tag:        tag,
		Prerelease: opts.Prerelease,
		MinTime:    mint,
	}

	return finder, tool
}

func getVerifier(asset Asset, assets []Asset, opts *Flags) (verifier Verifier, sumAsset Asset, err error) {

	sumAsset = Asset{}

	if opts.Verify != "" {
		verifier, err = NewSha256Verifier(opts.Verify)
		if err != nil {
			return nil, Asset{}, fmt.Errorf("create Sha256Verifier: %w", err)
		}
		return verifier, Asset{}, nil
	}

	for _, a := range assets {
		if a.Name == asset.Name+".sha256sum" || a.Name == asset.Name+".sha256" {
			fmt.Printf("verify against %s\n", a)

			return &Sha256AssetVerifier{AssetURL: a.DownloadURL}, a, nil
		}
		if strings.Contains(a.Name, "checksum") {
			binaryUrl, err := url.Parse(asset.DownloadURL)
			if err != nil {
				return nil, a, fmt.Errorf("extract binary name from asset url: %s: %w", asset, err)
			}
			binaryName := path.Base(binaryUrl.Path)
			fmt.Printf("verify against %s\n", a)
			return &Sha256SumFileAssetVerifier{Sha256SumAssetURL: a.DownloadURL, BinaryName: binaryName}, a, nil
		}
	}

	if opts.Hash {
		return &Sha256Printer{}, Asset{}, nil
	}

	return &NoVerifier{}, Asset{}, nil
}

// func getVerifier(sumAsset Asset, opts *Flags) (verifier Verifier, err error) {
// 	if opts.Verify != "" {
// 		verifier, err = NewSha256Verifier(opts.Verify)
// 	} else if sumAsset != (Asset{}) {
// 		verifier = &Sha256AssetVerifier{
// 			AssetURL: sumAsset.DownloadURL,
// 		}
// 	} else if opts.Hash {
// 		verifier = &Sha256Printer{}
// 	} else {
// 		verifier = &NoVerifier{}
// 	}
// 	return verifier, err
// }

// Determine which extractor to use. If --download-only is provided, we
// just "extract" the downloaded archive to itself. Otherwise we try to
// extract the literal file provided by --file, or by default we just
// extract a binary with the tool name that was possibly auto-detected
// above.
func getExtractor(url, tool string, opts *Flags) (extractor Extractor, err error) {
	if opts.DLOnly {
		extractor = &SingleFileExtractor{
			Name:   path.Base(url),
			Rename: path.Base(url),
			Decompress: func(r io.Reader) (io.Reader, error) {
				return r, nil
			},
		}
	} else if opts.ExtractFile != "" {
		gc, err := NewGlobChooser(opts.ExtractFile)
		if err != nil {
			return nil, err
		}
		extractor = NewExtractor(path.Base(url), tool, gc)
	} else {
		extractor = NewExtractor(path.Base(url), tool, &BinaryChooser{
			Tool: tool,
		})
	}
	return extractor, nil
}

// Would really like generics to implement this...
// Make the user select one of the choices and return the index of the
// selection.
func userSelect(choices []interface{}) int {
	for i, c := range choices {
		fmt.Fprintf(os.Stderr, "(%d) %v\n", i+1, c)
	}
	var choice int
	for {
		fmt.Fprint(os.Stderr, "Enter selection number: ")
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

		fmt.Fprintf(os.Stderr, "Invalid selection: %v\n", err)
	}
	return choice
}

func downloadConfigRepositories(config *Config) error {
	hasError := false
	errorList := []error{}

	binary, err := os.Executable()

	if err != nil {
		binary = os.Args[0]
	}

	for name, _ := range config.Repositories {
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

var opts Flags

func Run() {
	var cli CliFlags

	flagparser := flags.NewParser(&cli, flags.PassDoubleDash|flags.PrintErrors)
	flagparser.Usage = "[OPTIONS] TARGET"
	args, err := flagparser.Parse()

	if err != nil {
		os.Exit(1)
	}

	if cli.Version {
		fmt.Println("eget version", Version)
		os.Exit(0)
	}

	if cli.Help {
		flagparser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	config, err := InitializeConfig()
	if err != nil {
		Fatal(err)
	}

	if err := SetGlobalOptionsFromConfig(config, flagparser, &opts, cli); err != nil {
		Fatal(err)
	}

	if cli.Rate {
		rdat, err := GetRateLimit()
		if err != nil {
			Fatal(err)
		}
		fmt.Println(rdat)
		os.Exit(0)
	}

	target := ""

	if len(args) > 0 {
		target = args[0]
	}

	if err := SetProjectOptionsFromConfig(config, flagparser, &opts, cli, target); err != nil {
		Fatal(err)
	}

	if cli.DownloadAll {
		err = downloadConfigRepositories(config)
		ConditionalExit(err)
	}

	if len(args) <= 0 {
		fmt.Println("no target given")
		flagparser.WriteHelp(os.Stdout)
		SuccessExit()
	}

	if opts.DisableSSL {
		fmt.Fprintln(os.Stderr, "warning: SSL verification is disabled")
	}

	if opts.Remove {
		ebin := os.Getenv("EGET_BIN")
		err := os.Remove(filepath.Join(ebin, target))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Removed `%s`\n", filepath.Join(ebin, target))
		SuccessExit()
	}

	output := createOutputWriter()

	finder, tool := getFinder(target, &opts)
	assets, err := finder.Find()
	FatalIf(err)

	if err != nil && errors.Is(err, ErrNoUpgrade) {
		fmt.Fprintf(output, "%s: %v\n", target, err)
		SuccessExit()
	}

	detector, err := DetermineCorrectDetector(&opts)
	FatalIf(err)

	// get the url and candidates from the detector
	asset, candidates, err := detector.Detect(assets)
	FatalIf(err)

	if len(candidates) != 0 && err != nil {
		// if multiple candidates are returned, the user must select manually which one to download
		fmt.Fprintf(os.Stderr, "%v: please select manually\n", err)
		choices := make([]interface{}, len(candidates))
		for i := range candidates {
			choices[i] = path.Base(candidates[i].Name)
		}
		choice := userSelect(choices)
		asset = candidates[choice-1]
	}

	// download with progress bar and get the response body
	body := downloadAsset(&asset, output)
	verifyChecksums(asset, assets, body, output)

	extractor, err := getExtractor(asset.DownloadURL, tool, &opts)
	FatalIf(err)

	// get extraction candidates
	bin, bins, err := extractor.Extract(body, opts.All)

	if err != nil && len(bins) == 0 {
		Fatal(err)
	}

	if len(bins) != 0 && err != nil && !opts.All {
		// if there are multiple candidates, have the user select manually
		fmt.Fprintf(os.Stderr, "%v: please select manually\n", err)
		choices := make([]interface{}, len(bins)+1)
		for i := range bins {
			choices[i] = bins[i]
		}
		choices[len(bins)] = "all"
		choice := userSelect(choices)
		if choice == len(bins)+1 {
			opts.All = true
		} else {
			bin = bins[choice-1]
		}
	}

	if opts.All {
		if len(bins) == 0 {
			bins = []ExtractedFile{bin}
		}

		for _, bin := range bins {
			extract(bin, output)
		}
	}

	if !opts.All {
		extract(bin, output)
	}
}

func createOutputWriter() io.Writer {
	// when --quiet is passed, send non-essential output to io.Discard
	if opts.Quiet {
		return io.Discard
	}

	return os.Stderr
}

func downloadAsset(asset *Asset, output io.Writer) []byte {
	fmt.Fprintf(output, "%s\n", asset.DownloadURL) // print the URL

	buf := &bytes.Buffer{}
	if err := Download(asset.DownloadURL, buf); err != nil {
		Fatal(fmt.Sprintf("%s (URL: %s)", err, asset.DownloadURL))
	}

	return buf.Bytes()
}

func verifyChecksums(asset Asset, assets []Asset, body []byte, output io.Writer) {
	verifier, sumAsset, err := getVerifier(asset, assets, &opts)

	var verifiedStr string = ""

	if err != nil {
		verifiedStr = "Checksum verification failed, could not create a verifier.\n"
		fmt.Fprintf(output, verifiedStr)
		return
	}

	if err == nil {
		err = verifier.Verify(body)
	}

	if err != nil && verifiedStr == "" {
		verifiedStr = fmt.Sprintf("Checksum verification failed, %v\n", err)
	} else if opts.Verify == "" && sumAsset != (Asset{}) {
		verifiedStr = fmt.Sprintf("Checksum verified with %s\n", path.Base(sumAsset.Name))
	} else if opts.Verify != "" {
		verifiedStr = "Checksum verified\n"
	}

	fmt.Fprintf(output, verifiedStr)
}

func extract(bin ExtractedFile, output io.Writer) {
	mode := bin.Mode()

	// write the extracted file to a file on disk, in the --to directory if requested
	out := filepath.Base(bin.Name)
	if opts.Output == "-" {
		out = "-"
	} else if opts.Output != "" && IsDirectory(opts.Output) {
		out = filepath.Join(opts.Output, out)
	} else if opts.Output != "" && opts.All {
		os.MkdirAll(opts.Output, 0755)
		out = filepath.Join(opts.Output, out)
	} else {
		out = SetIf(opts.Output != "", opts.Output, out)

		// only use $EGET_BIN if all of the following are true
		// 1. $EGET_BIN is non-empty
		// 2. --to is not a path (not a path if no path separator is found)
		// 3. The extracted file is executable
		if os.Getenv("EGET_BIN") != "" && !strings.ContainsRune(out, os.PathSeparator) && mode&0111 != 0 && !bin.Dir {
			out = filepath.Join(os.Getenv("EGET_BIN"), out)
		}
	}

	if err := bin.Extract(out); err != nil {
		Fatal(err)
	}

	fmt.Fprintf(output, "Extracted `%s` to `%s`\n", bin.ArchiveName, out)
}
