package lockfile

import (
	"fmt"
	"os"
	"time"

	//"github.com/bytedance/sonic"
	"encoding/json"
)

type PackageInformation struct {
	Name    string
	Version string
}

type Lockfile struct {
	Filename  string
	UpdatedAt time.Time
	Packages  []PackageInformation
}

func NewLockfile(filename string) *Lockfile {
	return &Lockfile{
		Filename: filename,
		Packages: []PackageInformation{},
	}
}

// Save to file:
func (l *Lockfile) Save() error {
	if l.Filename == "" {
		return fmt.Errorf("Lockfile filename is empty")
	}

	return os.WriteFile(l.Filename, []byte(l.Json()), 0644)
}

func (l *Lockfile) Load() error {
	if l.Filename == "" {
		return fmt.Errorf("Lockfile filename is empty")
	}

	data, err := os.ReadFile(l.Filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, l)
}

func (l *Lockfile) Json() string {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return ""
	}

	return string(data)
}

func (l *Lockfile) AddPackage(packageInfo *PackageInformation) {
	l.Packages = append(l.Packages, *packageInfo)
}

func (l *Lockfile) GetPackage(name string) *PackageInformation {
	for _, packageInfo := range l.Packages {
		if packageInfo.Name == name {
			return &packageInfo
		}
	}
	return nil
}

func (l *Lockfile) RemovePackage(name string) {
	for i, packageInfo := range l.Packages {
		if packageInfo.Name == name {
			l.Packages = append(l.Packages[:i], l.Packages[i+1:]...)
			break
		}
	}
}
