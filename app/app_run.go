package app

import (
	"fmt"
	"time"

	"github.com/permafrost-dev/zeget/lib/assets"
	. "github.com/permafrost-dev/zeget/lib/assets"
	"github.com/permafrost-dev/zeget/lib/detectors"
	"github.com/permafrost-dev/zeget/lib/reporters"
	"github.com/permafrost-dev/zeget/lib/utilities"
	. "github.com/permafrost-dev/zeget/lib/utilities"
)

func (app *Application) Run() *ReturnStatus {
	app.Cache.LoadFromFile()

	target, returnStatus := app.RunSetup(FatalHandler)
	if returnStatus != nil {
		return returnStatus
	}

	if err := app.targetToProject(target); err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	cacheItem := app.Cache.Data.GetRepositoryEntryByKey(app.Target, &app.Cache)
	if len(app.Opts.Asset) == 0 && len(cacheItem.Filters) > 0 {
		app.Opts.Asset = cacheItem.Filters
	}

	app.RefreshRateLimit()
	if err := app.RateLimitExceeded(); err != nil {
		app.WriteErrorLine("GitHub rate limit exceeded. It resets at %s.", app.Cache.Data.RateLimit.Reset.Format(time.RFC1123))
		return NewReturnStatus(FatalError, nil, fmt.Sprintf("error: %v", err))
	}

	finder := app.getFinder()
	findResult := app.getFindResult(finder)

	if len(app.Opts.Filters) > 0 {
		var temp []assets.Asset = []assets.Asset{}

		for _, filter := range app.Opts.Filters {
			for _, a := range findResult.Assets {
				if filter.Apply(a) {
					temp = append(temp, a)
				}
			}
		}
		findResult.Assets = temp

		if len(findResult.Assets) == 0 {
			findResult.Error = fmt.Errorf("no assets found matching filters")
			return NewReturnStatus(FatalError, findResult.Error, fmt.Sprintf("error: %v", findResult.Error))
		}
	}

	app.cacheTarget(&finder, &findResult)

	if shouldReturn, returnStatus := app.shouldReturn(findResult.Error); shouldReturn {
		return returnStatus
	}

	assetWrapper := NewAssetWrapper(findResult.Assets)
	detector, err := detectors.DetermineCorrectDetector(&app.Opts, nil)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	// get the url and candidates from the detector
	detected, err := detector.Detect(assetWrapper.Assets)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	asset := detected.Asset

	if len(detected.Candidates) != 0 {
		asset, err = app.selectFromMultipleAssets(detected.Candidates, err) // manually select which asset to download
		if err != nil {
			return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
		}

		// convert the selected asset to an array of filters, then save them to file for future use
		app.Cache.Data.GetRepositoryEntryByKey(app.Target, &app.Cache).Filters = asset.Filters
		app.Cache.SaveToFile()
	}

	assetWrapper.Asset = &asset

	app.WriteLine("› downloading %s...", assetWrapper.Asset.DownloadURL) // print the URL

	body, err := app.downloadAsset(assetWrapper.Asset, &findResult) // download with progress bar and get the response body
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}
	app.VerifyChecksums(assetWrapper, body)

	if app.Opts.Sha256 || app.Opts.Hash {
		reporters.NewAssetSha256HashReporter(assetWrapper.Asset, app.Output).Report(string(body))
	}

	tagDownloaded := utilities.SetIf(app.Opts.Tag != "", "latest", app.Opts.Tag)
	app.Cache.Data.GetRepositoryEntryByKey(app.Target, &app.Cache).UpdateDownloadedAt(tagDownloaded)

	extractor, err := app.getExtractor(assetWrapper.Asset, finder.Tool)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	bin, bins, err := extractor.Extract(body, app.Opts.All) // get extraction candidates
	// if err != nil && len(bins) == 0 {
	// 	return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	// }
	if err != nil && len(bins) != 0 && !app.Opts.All {
		var e error
		bin, e = app.selectFromMultipleCandidates(bin, bins, err)
		if e != nil {
			return NewReturnStatus(FatalError, e, fmt.Sprintf("error: %v", e))
		}
	}

	extractedCount := app.ExtractBins(bin, app.wrapBins(bins, bin), app.Opts.All)

	if app.Opts.Verbose {
		reporters.NewMessageReporter(app.Output, "number of extracted files: %d\n", extractedCount).Report()
	}

	return NewReturnStatus(Success, nil, fmt.Sprintf("extracted files: %d", extractedCount))
}
