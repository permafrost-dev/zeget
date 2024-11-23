package registry

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/permafrost-dev/zeget/lib/errors"
	"github.com/permafrost-dev/zeget/lib/utilities"
)

// LockFile contains all the data for the lockfile
type LockFile struct {
	Os       string        `json:"os"`
	Arch     string        `json:"arch"`
	Packages []PackageData `json:"packages"`
	Filename string
}

// PackageData contains the information for an installed binary
type PackageData struct {
	Source       string   `json:"source"`
	Owner        string   `json:"owner"`
	Repo         string   `json:"repo"`
	Tag          string   `json:"tag"`
	InstalledAt  string   `json:"date_installed"`
	AssetFilters []string `json:"asset_filters"`
	Asset        string   `json:"asset"`
	Binary       string   `json:"binary"`
	URL          string   `json:"url"`
	BinaryHash   string   `json:"binaryHash"`
}

func (lf *LockFile) Save() {
	WriteLockFileJSON(*lf, lf.Filename)
}

func (lf *LockFile) AddPackage(pkg PackageData) {
	lf.Packages = append(lf.Packages, pkg)
}

func (lf *LockFile) RemovePackageByIndex(index int) error {
	var err error

	lf.Packages, err = RemovePackage(lf.Packages, index)

	return err
}

func (lf *LockFile) RemovePackage(repoName string) error {
	for i, pkg := range lf.Packages {
		if pkg.Owner+"/"+pkg.Repo == repoName {
			lf.Packages, _ = RemovePackage(lf.Packages, i)
			return nil
		}
	}

	return errors.AssetsNotFoundError{Tag: repoName}
}

func (lf *LockFile) AddOrUpdatePackage(pkg PackageData) {
	for i, p := range lf.Packages {
		if p.Owner+"/"+p.Repo == pkg.Owner+"/"+pkg.Repo {
			lf.Packages[i] = pkg
			return
		}
	}

	lf.AddPackage(pkg)
}

func (lf *LockFile) GetPackage(repoName string) (PackageData, error) {
	for _, pkg := range lf.Packages {
		if pkg.Owner+"/"+pkg.Repo == repoName {
			return pkg, nil
		}
	}

	return PackageData{}, errors.AssetsNotFoundError{Tag: repoName}
}

func readLockFileJSON(lockFilePath string) (LockFile, error) {

	lockFileBytes, err := os.ReadFile(lockFilePath)
	if err != nil {
		return LockFile{}, err
	}

	var lockFile LockFile
	err = json.Unmarshal(lockFileBytes, &lockFile)
	if err != nil {
		return LockFile{}, err
	}

	return lockFile, nil
}

// WriteLockFileJSON will write the lockfile JSON file
func WriteLockFileJSON(lockFileJSON LockFile, outputPath string) error {

	lockFileBytes, err := json.MarshalIndent(lockFileJSON, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, lockFileBytes, 0644)
	if err != nil {
		return err
	}

	// fmt.Printf("📄 Updated %v\n", constants.GreenColor(outputPath))

	return nil
}

// RemovePackage will remove a package from a LockFile.Packages slice
func RemovePackage(pkgs []PackageData, index int) ([]PackageData, error) {
	if len(pkgs) == 0 {
		return []PackageData{}, errors.NoPackagesInLockfileError{}
	}

	if index < 0 || index >= len(pkgs) {
		return []PackageData{}, errors.IndexOutOfBoundsInLockfileError{}
	}

	return append(pkgs[:index], pkgs[index+1:]...), nil
}

// ReadRegistryFileContents will read the contents of the Stewfile
// func ReadRegistryFileContents(registryFilePath string) ([]PackageData, error) {
// 	file, err := os.Open(registryFilePath)
// 	if err != nil {
// 		return []PackageData{}, err
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)

// 	var packages []PackageData
// 	for scanner.Scan() {
// 		packageText := scanner.Text()
// 		pkg, err := ParseCLIInput(packageText)
// 		if err != nil {
// 			return []PackageData{}, err
// 		}
// 		packages = append(packages, pkg)
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return []PackageData{}, err
// 	}

// 	return packages, nil
// }

// ReadRegistryLockFileContents will read the contents of the Stewfile.lock.json
func ReadRegistryLockFileContents(lockFilePath string) ([]PackageData, error) {
	lockFile, err := readLockFileJSON(lockFilePath)
	if err != nil {
		return []PackageData{}, err
	}
	return lockFile.Packages, nil
}

// NewLockFile creates a new instance of the LockFile struct
func NewLockFile(registryLockFilePath, userOS, userArch string) (LockFile, error) {
	var lockFile LockFile
	lockFileExists, err := utilities.PathExists(registryLockFilePath)
	if err != nil {
		return LockFile{}, err
	}
	if !lockFileExists {
		lockFile = LockFile{Os: userOS, Arch: userArch, Packages: []PackageData{}}
	} else {
		lockFile, err = readLockFileJSON(registryLockFilePath)
		if err != nil {
			return LockFile{}, err
		}
	}

	lockFile.Filename = registryLockFilePath

	return lockFile, nil
}

// DeleteAssetAndBinary will delete the asset from the ~/.stew/pkg path and delete the binary from the ~/.stew/bin path
func DeleteAssetAndBinary(stewPkgPath, stewBinPath, asset, binary string) error {
	assetPath := filepath.Join(stewPkgPath, asset)
	binPath := filepath.Join(stewBinPath, binary)
	err := os.RemoveAll(assetPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(binPath)
	if err != nil {
		return err
	}
	return nil
}
