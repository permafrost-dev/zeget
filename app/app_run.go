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

	if result := app.ProcessFilters(&finder, &findResult); result != nil {
		return result
	}

	app.cacheTarget(&finder, &findResult)

	if shouldReturn, returnStatus := app.shouldReturn(findResult.Error); shouldReturn {
		return returnStatus
	}

	assetWrapper := NewAssetWrapper(findResult.Assets)
	detector, err := detectors.DetermineCorrectDetector(&app.Opts, app.Config.Global.IgnorePatterns, nil)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	// get the url and candidates from the detector
	detected, err := detector.Detect(assetWrapper.Assets)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	filterDetector, _ := detectors.GetPatternDetectors(app.Config.Global.IgnorePatterns, nil)
	filteredDetected, err := filterDetector.DetectWithoutSystem(findResult.Assets)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	if filteredDetected != nil {
		//remove filteredDetected.Candidates from detected.Candidates
		detected.Candidates = FilterArr(detected.Candidates, func(a assets.Asset) bool {
			return IsInArr(filteredDetected.Candidates, a, func(a1 assets.Asset, a2 assets.Asset) bool { return a1.Name == a2.Name })
		})

		if len(detected.Candidates) == 1 {
			detected.Asset = detected.Candidates[0]
			detected.Candidates = []assets.Asset{}
		}
	}

	assetWrapper.Asset = &detected.Asset

	if len(detected.Candidates) != 0 {
		assetWrapper.Asset, err = app.selectFromMultipleAssets(detected.Candidates, err) // manually select which asset to download
		if err != nil {
			return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
		}

		cacheItem.Filters = assetWrapper.Asset.Filters
	}

	body, result := app.DownloadAndVerify(assetWrapper, &findResult)
	if result != nil {
		return result
	}

	extractedCount, result := app.ExtractDownloadedAsset(assetWrapper, body, &finder)
	if result != nil {
		return result
	}

	cacheItem.LastDownloadAt = time.Now().Local()
	cacheItem.LastDownloadTag = utilities.ParseVersionTagFromURL(assetWrapper.Asset.DownloadURL, app.Opts.Tag)
	cacheItem.LastDownloadHash = utilities.CalculateStringHash(string(body))
	cacheItem.Save()

	if app.Opts.Verbose {
		reporters.NewMessageReporter(app.Output, "number of extracted files: %d\n", extractedCount).Report()
	}

	return NewReturnStatus(Success, nil, fmt.Sprintf("extracted files: %d", extractedCount))
}
