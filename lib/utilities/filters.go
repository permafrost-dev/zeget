package utilities

import (
	"regexp"
)

// FilenameToAssetFilters takes a filename string and returns a list of inclusive asset filters based on the filename
// it's used to avoid the need for prompting the user to select an asset on subsequent runs after having selected an asset
// on the first download.
func FilenameToAssetFilters(filename string) []string {
	//split the string based on the regex pattern `\b`:
	re := regexp.MustCompile(`(arm64|amd64|x86|i386|mips64|[a-zA-Z]+)`)
	parts := re.FindAllStringSubmatch(filename, -1)

	result := []string{}
	for idx, part := range parts {
		if idx == 0 {
			continue
		}
		if len(part) > 0 {
			result = append(result, part[0])
		}
	}

	return result
}
