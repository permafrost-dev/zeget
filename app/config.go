package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/permafrost-dev/eget/lib/globals"
	"github.com/permafrost-dev/eget/lib/home"
	"github.com/permafrost-dev/eget/lib/utilities"
)

type ConfigGlobal struct {
	All          bool   `toml:"all"`
	DownloadOnly bool   `toml:"download_only"`
	File         string `toml:"file"`
	GithubToken  string `toml:"github_token"`
	Quiet        bool   `toml:"quiet"`
	ShowHash     bool   `toml:"show_hash"`
	Source       bool   `toml:"download_source"`
	System       string `toml:"system"`
	Target       string `toml:"target"`
	UpgradeOnly  bool   `toml:"upgrade_only"`
}

type ConfigRepository struct {
	All          bool     `toml:"all"`
	AssetFilters []string `toml:"asset_filters"`
	DownloadOnly bool     `toml:"download_only"`
	File         string   `toml:"file"`
	Name         string   `toml:"name"`
	Quiet        bool     `toml:"quiet"`
	ShowHash     bool     `toml:"show_hash"`
	Source       bool     `toml:"download_source"`
	System       string   `toml:"system"`
	Tag          string   `toml:"tag"`
	Target       string   `toml:"target"`
	UpgradeOnly  bool     `toml:"upgrade_only"`
	Verify       string   `toml:"verify_sha256"`
	DisableSSL   bool     `toml:"disable_ssl"`
}

type Config struct {
	Meta struct {
		Keys     []string
		MetaData *toml.MetaData
	}
	Global       ConfigGlobal `toml:"global"`
	Repositories map[string]ConfigRepository
}

func BuildConfigurationFilename(paths ...string) string {
	var result string
	if len(paths) > 0 {
		result = filepath.Join(paths...)
	}

	return filepath.Join(result, "."+globals.ApplicationName+".toml")
}

func LoadConfigurationFile(path string) (*Config, error) {
	var conf Config
	meta, err := toml.DecodeFile(path, &conf)

	if err != nil {
		return &conf, err
	}

	meta, err = toml.DecodeFile(path, &conf.Repositories)

	conf.Meta.Keys = make([]string, len(meta.Keys()))

	for i, key := range meta.Keys() {
		conf.Meta.Keys[i] = key.String()
	}

	conf.Meta.MetaData = &meta

	return &conf, err
}

func GetOSConfigPath(homePath string) string {
	var configDir string

	defaultConfig := map[string]string{
		"windows": "LocalAppData",
		"default": ".config",
	}

	var goos string
	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("LOCALAPPDATA")
		goos = "windows"
	default:
		configDir = os.Getenv("XDG_CONFIG_HOME")
		goos = "default"
	}

	if configDir == "" {
		configDir = filepath.Join(homePath, defaultConfig[goos])
	}

	return BuildConfigurationFilename(configDir)
}

func (app *Application) tryLoadingConfigFiles(config *Config, homePath string) (*Config, error) {
	var err error
	var cfg = config
	var filenames = []string{}

	if configFilePath, ok := os.LookupEnv("EGET_CONFIG"); ok && configFilePath != "" {
		filenames = append(filenames, configFilePath)
	}

	filenames = append(filenames,
		BuildConfigurationFilename(homePath),
		BuildConfigurationFilename(),
		GetOSConfigPath(homePath),
	)

	for _, filename := range filenames {
		if !utilities.IsLocalFile(filename) {
			continue
		}
		if cfg, err = LoadConfigurationFile(filename); err == nil {
			return cfg, nil
		}
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	if err == nil {
		err = fmt.Errorf("no configuration file found")
	}

	return &Config{}, err
}

func (app *Application) initializeConfig() {
	var err error
	var config *Config

	homePath, _ := os.UserHomeDir()
	config, err = app.tryLoadingConfigFiles(config, homePath)

	if err != nil {
		app.Config = &Config{
			Global: ConfigGlobal{
				All:          false,
				DownloadOnly: false,
				GithubToken:  "",
				Quiet:        false,
				ShowHash:     false,
				Source:       false,
				UpgradeOnly:  false,
			},
			Repositories: make(map[string]ConfigRepository, 0),
		}
		app.writeErrorLine("error loading configuration file: %s", err.Error())
		return
	}

	delete(config.Repositories, "global")

	// set default global values
	config.Global.All = utilities.SetIf(!config.Meta.MetaData.IsDefined("global", "system"), config.Global.All, false)
	config.Global.GithubToken = utilities.SetIf(!config.Meta.MetaData.IsDefined("global", "github_token"), config.Global.GithubToken, "")
	config.Global.Quiet = utilities.SetIf(!config.Meta.MetaData.IsDefined("global", "quiet"), config.Global.Quiet, false)
	config.Global.DownloadOnly = utilities.SetIf(!config.Meta.MetaData.IsDefined("global", "download_only"), config.Global.DownloadOnly, false)
	config.Global.ShowHash = utilities.SetIf(!config.Meta.MetaData.IsDefined("global", "show_hash"), config.Global.ShowHash, false)
	config.Global.UpgradeOnly = utilities.SetIf(!config.Meta.MetaData.IsDefined("global", "upgrade_only"), config.Global.UpgradeOnly, false)
	config.Global.Target = utilities.SetIf(!config.Meta.MetaData.IsDefined("global", "target"), config.Global.Target, utilities.GetCurrentDirectory())

	// set default repository values
	for name, repo := range config.Repositories {
		repo.All = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "all"), repo.All, config.Global.All)
		repo.AssetFilters = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "asset_filters"), repo.AssetFilters, []string{})
		repo.DownloadOnly = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "download_only"), repo.DownloadOnly, config.Global.DownloadOnly)
		repo.Quiet = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "quiet"), repo.Quiet, config.Global.Quiet)
		repo.ShowHash = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "show_hash"), repo.ShowHash, config.Global.ShowHash)
		repo.Target = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "target"), repo.Target, config.Global.Target)
		repo.UpgradeOnly = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "upgrade_only"), repo.UpgradeOnly, config.Global.UpgradeOnly)
		repo.Source = utilities.SetIf(!config.Meta.MetaData.IsDefined(name, "download_source"), repo.Source, config.Global.Source)

		config.Repositories[name] = repo
	}

	app.Config = config
}

func update[T any](config T, cli *T) T {
	if cli == nil {
		return config
	}
	return *cli
}

// Move the loaded configuration file global options into the opts variable
func (app *Application) SetGlobalOptionsFromConfig() error {

	if app.Config.Global.GithubToken != "" && os.Getenv("EGET_GITHUB_TOKEN") == "" {
		os.Setenv("EGET_GITHUB_TOKEN", app.Config.Global.GithubToken)
	}

	app.Opts.Tag = update("", app.cli.Tag)
	app.Opts.Prerelease = update(false, app.cli.Prerelease)
	app.Opts.Source = update(app.Config.Global.Source, app.cli.Source)
	targ, err := home.Expand(app.Config.Global.Target)
	if err != nil {
		return err
	}

	app.Opts.Output = update(targ, app.cli.Output)
	app.Opts.System = update(app.Config.Global.System, app.cli.System)
	app.Opts.ExtractFile = update("", app.cli.ExtractFile)
	app.Opts.All = update(app.Config.Global.All, app.cli.All)
	app.Opts.Quiet = update(app.Config.Global.Quiet, app.cli.Quiet)
	app.Opts.DLOnly = update(app.Config.Global.DownloadOnly, app.cli.DLOnly)
	app.Opts.UpgradeOnly = update(app.Config.Global.UpgradeOnly, app.cli.UpgradeOnly)
	app.Opts.Asset = update([]string{}, app.cli.Asset)
	app.Opts.Hash = update(app.Config.Global.ShowHash, app.cli.Hash)
	app.Opts.Verify = update("", app.cli.Verify)
	app.Opts.Remove = update(false, app.cli.Remove)
	app.Opts.DisableSSL = update(false, app.cli.DisableSSL)

	return nil
}

// Move the loaded configuration file project options into the opts variable
func (app *Application) SetProjectOptionsFromConfig(projectName string) error {
	for name, repo := range app.Config.Repositories {
		if name != projectName {
			continue
		}
		app.Opts.All = update(repo.All, app.cli.All)
		app.Opts.Asset = update(repo.AssetFilters, app.cli.Asset)
		app.Opts.DLOnly = update(repo.DownloadOnly, app.cli.DLOnly)
		app.Opts.ExtractFile = update(repo.File, app.cli.ExtractFile)
		app.Opts.Hash = update(repo.ShowHash, app.cli.Hash)
		targ, err := home.Expand(repo.Target)
		if err != nil {
			return err
		}
		app.Opts.Output = update(targ, app.cli.Output)
		app.Opts.Quiet = update(repo.Quiet, app.cli.Quiet)
		app.Opts.Source = update(repo.Source, app.cli.Source)
		app.Opts.System = update(repo.System, app.cli.System)
		app.Opts.Tag = update(repo.Tag, app.cli.Tag)
		app.Opts.UpgradeOnly = update(repo.UpgradeOnly, app.cli.UpgradeOnly)
		app.Opts.Verify = update(repo.Verify, app.cli.Verify)
		app.Opts.DisableSSL = update(repo.DisableSSL, app.cli.DisableSSL)

		break
	}

	return nil
}
