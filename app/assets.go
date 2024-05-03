package app

type AssetWrapper struct {
	Assets []Asset
	Asset  *Asset
}

func NewAssetWrapper(assets []Asset) *AssetWrapper {
	return &AssetWrapper{
		Assets: assets,
		Asset:  nil,
	}
}
