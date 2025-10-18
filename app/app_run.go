package app

import (
	"fmt"
	"time"

	. "github.com/permafrost-dev/zeget/lib/assets"
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

	finder, findResult := app.Find()

	if result := app.ProcessFilters(finder, findResult); result != nil {
		return result
	}

	cacheItem = app.cacheTarget(finder, findResult)

	if shouldReturn, returnStatus := app.shouldReturn(findResult.Error); shouldReturn {
		return returnStatus
	}

	assetWrapper := NewAssetWrapper(findResult.Assets)
	detected, err := app.DetectAssets(assetWrapper)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	if err := app.FilterDetectedAssets(detected, findResult, assetWrapper); err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	if err := app.HandleMultipleCandidates(detected, assetWrapper); err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	body, err := app.DownloadAndVerify(assetWrapper, findResult)
	if err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	if err := app.ExtractDownloadedAsset(assetWrapper, body, finder); err != nil {
		return NewReturnStatus(FatalError, err, fmt.Sprintf("error: %v", err))
	}

	cacheItem.Filters = assetWrapper.Asset.Filters
	cacheItem.LastDownloadAt = time.Now().Local()
	cacheItem.LastDownloadTag = utilities.ParseVersionTagFromURL(assetWrapper.Asset.DownloadURL, app.Opts.Tag)
	cacheItem.LastDownloadHash = utilities.CalculateStringHash(string(body))
	cacheItem.Save()

	return NewReturnStatus(Success, nil, fmt.Sprintf("successfully downloaded and extracted files"))
}
