package detectors_test

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"testing"

	"github.com/permafrost-dev/zeget/lib/appflags"
	. "github.com/permafrost-dev/zeget/lib/assets"
	. "github.com/permafrost-dev/zeget/lib/detectors"
)

func TestAllDetector_Detect(t *testing.T) {
	detector := AllDetector{}

	tests := []struct {
		name           string
		assets         []Asset
		wantMatch      string
		wantCandidates []Asset
		wantErr        bool
	}{
		{
			name: "Single asset",
			assets: []Asset{
				{Name: "asset1", DownloadURL: "http://example.com/asset1"},
			},
			wantMatch:      "asset1",
			wantCandidates: nil,
			wantErr:        false,
		},
		{
			name: "Multiple assets",
			assets: []Asset{
				{Name: "asset1", DownloadURL: "http://example.com/asset1"},
				{Name: "asset2", DownloadURL: "http://example.com/asset2"},
			},
			wantMatch: "",
			wantCandidates: []Asset{
				{Name: "asset1", DownloadURL: "http://example.com/asset1"},
				{Name: "asset2", DownloadURL: "http://example.com/asset2"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDetected, err := detector.Detect(tt.assets)
			gotMatch := gotDetected.Asset
			gotCandidates := gotDetected.Candidates
			if (err != nil) != tt.wantErr {
				t.Errorf("AllDetector.Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatch.Name != tt.wantMatch {
				t.Errorf("AllDetector.Detect() gotMatch = %v, want %v", gotMatch, tt.wantMatch)
			}
			if !reflect.DeepEqual(gotCandidates, tt.wantCandidates) {
				t.Errorf("AllDetector.Detect() gotCandidates = %v, want %v", gotCandidates, tt.wantCandidates)
			}
		})
	}
}

func TestSingleAssetDetector_Detect(t *testing.T) {
	tests := []struct {
		name           string
		detector       SingleAssetDetector
		assets         []Asset
		wantMatch      string
		wantCandidates []string
		wantErr        bool
	}{
		{
			name:     "Exact match",
			detector: SingleAssetDetector{Asset: "asset1", Anti: false},
			assets: []Asset{
				{Name: "asset1", DownloadURL: "http://example.com/asset1"},
				{Name: "asset2", DownloadURL: "http://example.com/asset2"},
			},
			wantMatch:      "asset1",
			wantCandidates: nil,
			wantErr:        false,
		},
		{
			name:     "No match",
			detector: SingleAssetDetector{Asset: "asset3", Anti: false},
			assets: []Asset{
				{Name: "asset1", DownloadURL: "http://example.com/asset1"},
				{Name: "asset2", DownloadURL: "http://example.com/asset2"},
			},
			wantMatch:      "",
			wantCandidates: nil,
			wantErr:        true,
		},
		{
			name:     "Anti match",
			detector: SingleAssetDetector{Asset: "asset1", Anti: true},
			assets: []Asset{
				{Name: "asset1", DownloadURL: "http://example.com/asset1"},
				{Name: "asset2", DownloadURL: "http://example.com/asset2"},
			},
			wantMatch:      "asset2",
			wantCandidates: nil,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDetected, err := tt.detector.Detect(tt.assets)
			gotMatch := gotDetected.Asset
			gotCandidates := gotDetected.Candidates
			if (err != nil) != tt.wantErr {
				t.Errorf("SingleAssetDetector.Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatch.Name != tt.wantMatch {
				t.Errorf("SingleAssetDetector.Detect() gotMatch = %v, want %v", gotMatch, tt.wantMatch)
			}
			if len(gotCandidates) != len(tt.wantCandidates) {
				t.Errorf("SingleAssetDetector.Detect() gotCandidates = %v, want %v", gotCandidates, tt.wantCandidates)
			}
		})
	}
}

func TestSystemDetector_Detect(t *testing.T) {
	linuxAMD64, _ := NewSystemDetector("linux", "amd64")
	linuxARM, _ := NewSystemDetector("linux", "arm")

	tests := []struct {
		name           string
		detector       *SystemDetector
		assets         []Asset
		wantMatch      string
		wantCandidates []Asset
		wantErr        bool
	}{
		{
			name:     "Match OS and Arch",
			detector: linuxAMD64,
			assets: []Asset{
				{Name: "program-linux-amd64.tar.gz", DownloadURL: "http://example.com/program-linux-amd64.tar.gz"},
				{Name: "program-linux-arm.tar.gz", DownloadURL: "http://example.com/program-linux-arm.tar.gz"},
			},
			wantMatch:      "program-linux-amd64.tar.gz",
			wantCandidates: nil,
			wantErr:        false,
		},
		{
			name:     "Match only OS",
			detector: linuxARM,
			assets: []Asset{
				{Name: "program-linux-amd64.tar.gz", DownloadURL: "http://example.com/program-linux-amd64.tar.gz"},
				{Name: "program-linux-arm.tar.gz", DownloadURL: "http://example.com/program-linux-arm.tar.gz"},
			},
			wantMatch:      "program-linux-arm.tar.gz",
			wantCandidates: nil,
			wantErr:        false,
		},
		{
			name:     "No matches",
			detector: linuxAMD64,
			assets: []Asset{
				{Name: "program-windows-amd64.zip", DownloadURL: "http://example.com/program-windows-amd64.zip"},
				{Name: "program-macos.dmg", DownloadURL: "http://example.com/program-macos.dmg"},
			},
			wantMatch: "",
			wantCandidates: []Asset{
				{Name: "program-windows-amd64.zip", DownloadURL: "http://example.com/program-windows-amd64.zip"},
				{Name: "program-macos.dmg", DownloadURL: "http://example.com/program-macos.dmg"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDetected, err := tt.detector.Detect(tt.assets)
			gotMatch := gotDetected.Asset
			gotCandidates := gotDetected.Candidates

			if (err != nil) != tt.wantErr {
				t.Errorf("SystemDetector.Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatch.Name != tt.wantMatch {
				t.Errorf("SystemDetector.Detect() gotMatch = %v, want %v", gotMatch, tt.wantMatch)
			}
			if !reflect.DeepEqual(gotCandidates, tt.wantCandidates) {
				t.Errorf("SystemDetector.Detect() gotCandidates = %v, want %v", gotCandidates, tt.wantCandidates)
			}
		})
	}
}

func TestDetermineCorrectDetector(t *testing.T) {
	tests := []struct {
		name       string
		flags      appflags.Flags
		want       Detector
		wantErr    bool
		wantAssets []string
	}{
		{
			name: "all detector",
			flags: appflags.Flags{
				System: "all",
			},
			want:       &AllDetector{},
			wantErr:    false,
			wantAssets: []string{},
		},
		{
			name: "specific system detector",
			flags: appflags.Flags{
				System: "linux/amd64",
			},
			want: &SystemDetector{
				Os:   OSLinux,
				Arch: ArchAMD64,
			},
			wantErr:    false,
			wantAssets: []string{},
		},
		{
			name:  "default system detector",
			flags: appflags.Flags{
				// System is empty, should default to runtime.GOOS/runtime.GOARCH
			},
			want: &SystemDetector{
				Os:   OSLinux,
				Arch: ArchAMD64,
			},
			wantErr:    false,
			wantAssets: []string{},
		},
		{
			name: "invalid system format",
			flags: appflags.Flags{
				System: "invalidformat",
			},
			want: &SystemDetector{
				Os:   OSLinux,
				Arch: ArchAMD64,
			},
			wantErr:    false,
			wantAssets: []string{},
		},
		{
			name: "test 3",
			flags: appflags.Flags{
				System: "linux/amd64",
				Asset:  []string{"asset1", "^asset2"},
			},
			want: &DetectorChain{
				Detectors: []Detector{
					&SingleAssetDetector{
						Asset:     "asset1",
						Anti:      false,
						IsPattern: false,
						Compiled:  nil,
					},
					&SingleAssetDetector{
						Asset:     "asset2",
						Anti:      true,
						IsPattern: true,
						Compiled:  regexp.MustCompile("^asset2"),
					},
				},
			},
			wantErr:    false,
			wantAssets: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system, _ := NewSystemDetector(runtime.GOOS, runtime.GOARCH)
			got, err := DetermineCorrectDetector(&tt.flags, []string{}, system)
			fmt.Printf("got: %v\n", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineCorrectDetector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("DetermineCorrectDetector() got = %T, want %T", got, tt.want)
			}
		})
	}
}

func TestGetPatternDetectors(t *testing.T) {
	ignoredPatterns := []string{
		"^foo$",
		"bar.*",
		"baz[0-9]+",
	}

	systemDetector := &SystemDetector{
		Os:   OSLinux,
		Arch: ArchAMD64,
	}

	detectorChain, err := GetPatternDetectors(ignoredPatterns, systemDetector)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if detectorChain == nil {
		t.Fatal("Expected detectorChain to be non-nil")
	}

	expectedLength := len(ignoredPatterns)
	if len(detectorChain.Detectors) != expectedLength {
		t.Fatalf("Expected %d detectors, got %d", expectedLength, len(detectorChain.Detectors))
	}

	for i, detector := range detectorChain.Detectors {
		singleAssetDetector, ok := detector.(*SingleAssetDetector)
		if !ok {
			t.Fatalf("Detector at index %d is not a *SingleAssetDetector", i)
		}

		expectedPattern := ignoredPatterns[i]

		if singleAssetDetector.Asset != expectedPattern {
			t.Errorf("Detector at index %d has Asset %q, expected %q", i, singleAssetDetector.Asset, expectedPattern)
		}

		if !singleAssetDetector.Anti {
			t.Errorf("Detector at index %d has Anti %v, expected true", i, singleAssetDetector.Anti)
		}

		if !singleAssetDetector.IsPattern {
			t.Errorf("Detector at index %d has IsPattern %v, expected true", i, singleAssetDetector.IsPattern)
		}

		if singleAssetDetector.Compiled == nil {
			t.Errorf("Detector at index %d has nil Compiled regexp", i)
		} else {
			if singleAssetDetector.Compiled.String() != expectedPattern {
				t.Errorf("Detector at index %d has Compiled regexp %q, expected %q", i, singleAssetDetector.Compiled.String(), expectedPattern)
			}
		}
	}

	if detectorChain.System != systemDetector {
		t.Errorf("Expected detectorChain.System to be the same as input systemDetector")
	}
}
