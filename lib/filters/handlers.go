package filters

import (
	"path"
	"strings"

	"github.com/permafrost-dev/eget/lib/assets"
)

var FilterMap = map[string]*Filter{
	"all":  NewFilter("all", allHandler, FilterActionInclude),
	"any":  NewFilter("any", anyHandler, FilterActionInclude),
	"ext":  NewFilter("ext", extensionHandler, FilterActionInclude),
	"none": NewFilter("none", noneHandler, FilterActionExclude),
	"has":  NewFilter("has", hasHandler, FilterActionInclude),
}

var anyHandler FilterHandler = func(item assets.Asset, args []string) bool {
	for _, arg := range args {
		if strings.EqualFold(item.Name, arg) {
			return true
		}
	}

	return false
}

var allHandler FilterHandler = func(item assets.Asset, args []string) bool {
	for _, arg := range args {
		if !strings.EqualFold(item.Name, arg) {
			return false
		}
	}

	return true
}

var hasHandler FilterHandler = func(item assets.Asset, args []string) bool {
	for _, arg := range args {
		if !strings.EqualFold(item.Name, arg) {
			return false
		}
	}

	return true
}

var noneHandler FilterHandler = func(item assets.Asset, args []string) bool {
	for _, arg := range args {
		if strings.EqualFold(item.Name, arg) {
			return false
		}
	}

	return true
}

var extensionHandler FilterHandler = func(item assets.Asset, args []string) bool {
	extension := path.Ext(item.Name)

	for _, arg := range args {
		if strings.EqualFold(extension, arg) {
			return true
		}
	}

	return false
}
