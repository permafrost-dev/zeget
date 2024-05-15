# zeget: easy pre-built binary installation

<p align="center">
  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/permafrost-dev/zeget/build-and-test.yml?branch=main&style=flat-square&logo=github&logoColor=white&label=test%20suite&nocache=1">
  <!--<img alt="Release" src="https://img.shields.io/github/release/permafrost-dev/zeget.svg?label=Release&style=flat-square" />-->
  <img alt="Github last commit (main branch)" src="https://img.shields.io/github/last-commit/permafrost-dev/zeget/main?display_timestamp=committer&style=flat-square&logo=github&logoColor=white" />
  <img alt="MIT License" src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" />
  <br>
  <img alt="codecov" src="https://img.shields.io/codecov/c/gh/permafrost-dev/zeget?style=flat-square&logo=codecov&logoColor=white" />
  <img alt="Code Climate maintainability" src="https://img.shields.io/codeclimate/maintainability/permafrost-dev/zeget?style=flat-square&logo=codeclimate&logoColor=white" />
  <img alt="Code Climate tech debt" src="https://img.shields.io/codeclimate/tech-debt/permafrost-dev/zeget?style=flat-square&logo=codeclimate&logoColor=white&nocache=1" />
  <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/permafrost-dev/zeget?style=flat-square" />
</p>

**zeget** is the best way to easily get pre-built binaries for your favorite
tools; it downloads and extracts pre-built binaries from releases on GitHub.

To use it, provide a repository and zeget will search through the assets from the
latest release to find a suitable binary for your system. If successful, the asset is
downloaded and extracted to the current directory.

![zeget Demo](https://github.com/zyedidia/blobs/blob/master/eget-demo.gif?raw=true)

> **zeget origins**
>
> zeget was forked from the original [eget](https://github.com/zyedidia/eget) to implement bug fixes,
> improvements, new features, user requests, and implement major refactoring to a package-based
> architecture.
> zeget also endeavors to have a robust test suite with an acceptable code coverage percentage.
> It is mostly backward-compatible with eget, and should work as a drop-in replacement for most cases.
>

There are a few notable changes from the original utility:

- Verify checksums from `checksums.txt` assets
- Download release assets from private repositories
- Remembers user asset selections to avoid re-prompting
- Prettier output
- Numerous bug fixes
- New flags, such as `--no-interaction`
- Improved test coverage using `ginkgo` and a CI/CD pipeline
- `golangci-lint` integration
- `goreleaser` integration

zeget has a number of detection mechanisms and should work out-of-the-box with
most software that is distributed via single binaries on GitHub releases. First
try using zeget on your software, it is likely that is will "just work".
Otherwise, see the FAQ for a clear set of rules to make your software compatible with zeget.

For more in-depth documentation, see [DOCS.md](DOCS.md).

## Examples

```sh
zeget zyedidia/micro -t nightly
zeget jgm/pandoc --to /usr/local/bin
zeget junegunn/fzf
zeget neovim/neovim --no-interaction
zeget ogham/exa --asset ^musl
zeget --system darwin/amd64 sharkdp/fd
zeget BurntSushi/ripgrep
zeget -f eget.1 permafrost-dev/zeget
zeget zachjs/sv2v
zeget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz --file go --to ~/go1.21.1
zeget --all --file '*' ActivityWatch/activitywatch
```

## How to get zeget

Before you can get anything, you have to get zeget. If you already have zeget and want to upgrade, use `zeget permafrost-dev/zeget`.

### Quick-install script

```sh
curl -o install.sh https://raw.githubusercontent.com/permafrost-dev/zeget/main/scripts/install.sh
sha256sum install.sh # verify with hash below
bash install.sh
```

Or alternatively (less secure):

```sh
curl https://raw.githubusercontent.com/permafrost-dev/zeget/main/scripts/install.sh | sh
```

You can then place the downloaded binary in a location on your `$PATH` such as `/usr/local/bin`.

To verify the script, use `sha256sum install.sh` after downloading; the sha256 checksum is:

> <code data-id="installer-checksum">077d2431344eae1765039dfb8b3b4dff2efdd36677e5d49673dd0fe650d73e87</code><br>

### Pre-built binaries

Pre-built binaries are available on the [releases](https://github.com/permafrost-dev/zeget/releases) page.

### From source

Install the latest released version:

```sh
go install github.com/permafrost-dev/zeget@latest
```

or install from HEAD:

```sh
git clone https://github.com/permafrost-dev/zeget
cd eget
task build # or go build (produces incomplete version information)
```

zeget uses Task, a modern replacement for `make`, for performing various tasks. Install Task from the [official repository](https://github.com/go-task/task).

A man page can be generated by cloning the repository and running `task build-docs` (requires pandoc).
You can also use `zeget` to download the man page: `zeget -f zeget.1 permafrost-dev/zeget`.

## Usage

The `TARGET` argument passed to zeget should either be a GitHub repository,
formatted as `user/repo`, in which case Eget will search the release assets, a
direct URL, in which case Eget will directly download and extract from the
given URL, or a local file, in which case Eget will extract directly from the
local file.

If zeget downloads an asset called `xxx` and there also exists an asset called
`xxx.sha256` or `xxx.sha256sum`, or zeget will automatically verify that the
SHA-256 checksum of the downloaded asset matches the one contained in that
file, and abort installation if a mismatch occurs.

Likewise, if there is a `checksums.txt` asset in the release, zeget will attempt
to use this file to verify the SHA-256 checksum of the downloaded asset. It is
assumed that the file was generated using `sha256sum` or a similar tool, and that
the file contains individual lines in the format `hash file.ext`, with one
line per filename/hash.

When installing an executable, zeget will place it in the current directory by
default. If the environment variable `ZEGET_BIN` is non-empty, zeget will
place the executable in that directory.

Directories can also be specified as files to extract, and all files within
them will be extracted. For example:

```sh
zeget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz --file go --to ~/go1.21.1
```

GitHub limits API requests to 60 per hour for unauthenticated users. If you
would like to perform more requests (up to 5,000 per hour), you can set up a
personal access token and assign it to an environment variable named either
`GITHUB_TOKEN` or `ZEGET_GITHUB_TOKEN` when running zeget. If both are set,
`ZEGET_GITHUB_TOKEN` will take precedence. zeget will read this variable and
send the token as authorization with requests to GitHub. It is also possible
to read the token from a file by using `@/path/to/file` as the token value.

Zeget uses a cache to store information about repositories, releases, and user-selected
downloads when multiple assets are available. The cache is stored in the user's home
directory by default as `~/.zeget.cache.json`. The cache allows zeget to remember and
reuse selections made by the user (avoiding user input), and to avoid making unnecessary
requests to GitHub.

Note that some flags have changed from the [original utility](https://github.com/zyedidia/eget). The following is the output of `zeget --help`:

```sh
Usage:
  zeget [OPTIONS] TARGET

Application Options:
  -t, --tag=            tagged release to use instead of latest
      --pre-release     include pre-releases when fetching the latest version
      --source          download the source code for the target repo instead of a release
      --to=             move to given location after extracting
  -s, --system=         target system to download for (use "all" for all choices)
  -f, --file=           glob to select files for extraction
      --all             extract all candidate files
  -q, --quiet           only print essential output
  -d, --download-only   stop after downloading the asset (no extraction)
      --upgrade-only    only download if release is more recent than current version
  -a, --asset=          download a specific asset containing the given string; can be 
                        specified multiple times  for additional filtering; 
                        use '^' or '!' prefix for anti-match
  -H, --hash            show the SHA-256 hash of the downloaded asset
      --sha256          show the SHA-256 hash of the downloaded asset
      --verify-sha256=  verify the downloaded asset checksum against the one provided
  -r, --remove          remove the given file from $EGET_BIN or the current directory
  -V, --version         show version information
  -h, --help            show this help message
  -D, --download-all    download all projects defined in the config file
  -k, --disable-ssl     disable SSL verification for download requests
      --no-interaction  do not prompt for user input
  -v, --verbose         show verbose output
      --no-progress     do not show download progress
```

## Configuration

zeget can be configured using a TOML file located at `~/.zeget.toml` or it will fallback to the expected `XDG_CONFIG_HOME` directory of your os. Alternatively,
the configuration file can be located in the same directory as the zeget binary or the path specified with the environment variable `ZEGET_CONFIG`.

Both global settings can be configured, as well as setting on a per-repository basis.

Sections can be named either `global` or `"owner/repo"`, where `owner` and `repo`
are the owner and repository name of the target repository (not that the `owner/repo`
format is quoted).

For example, the following configuration file will set the `--to` flag to `~/bin` for
all repositories, and will set the `--to` flag to `~/.local/bin` for the `zendydia/micro`
repository.

```toml
[global]
target = "~/bin"

["zyedidia/micro"]
target = "~/.local/bin"
```

## Available settings - global section

| Setting | Related Flag | Description | Default |
| --- | --- | --- | --- |
| `github_token` | `N/A` | GitHub API token to use for requests | `""` |
| `all` | `--all` | Whether to extract all candidate files. | `false` |
| `download_only` | `--download-only` | Whether to stop after downloading the asset (no extraction). | `false` |
| `download_source` | `--source` | Whether to download the source code for the target repo instead of a release. | `false` |
| `file` | `--file` | The glob to select files for extraction. | `*` |
| `quiet` | `--quiet` | Whether to only print essential output. | `false` |
| `show_hash` | `--sha256` | Whether to show the SHA-256 hash of the downloaded asset. | `false` |
| `system` | `--system` | The target system to download for. | `all` |
| `target` | `--to` | The directory to move the downloaded file to after extraction. | `.` |
| `upgrade_only` | `--upgrade-only` | Whether to only download if release is more recent than current version. | `false` |
| `ignore_patterns` | `N/A` | An array of regular expressions to always ignore when detecting candidates for selection or extraction. | `[]` |

## Available settings - repository sections

| Setting | Related Flag | Description | Default |
| --- | --- | --- | --- |
| `all` | `--all` | Whether to extract all candidate files. | `false` |
| `asset_filters` | `--asset` |  An array of partial asset names to filter the available assets for download. | `[]` |
| `download_only` | `--download-only` | Whether to stop after downloading the asset (no extraction). | `false` |
| `download_source` | `--source` | Whether to download the source code for the target repo instead of a release. | `false` |
| `file` | `--file` | The glob to select files for extraction. | `*` |
| `quiet` | `--quiet` | Whether to only print essential output. | `false` |
| `show_hash` | `--sha256` | Whether to show the SHA-256 hash of the downloaded asset. | `false` |
| `system` | `--system` | The target system to download for. | `all` |
| `target` | `--to` | The directory to move the downloaded file to after extraction. | `.` |
| `upgrade_only` | `--upgrade-only` | Whether to only download if release is more recent than current version. | `false` |
| `verify_sha256` | `--verify-sha256` | Verify the sha256 hash of the asset against a provided hash. | `""` |

## Example configuration

```toml
[global]
    github_token = "ghp_1234567890"
    quiet = false
    show_hash = false
    upgrade_only = true
    target = "/home/user1/bin"
    ignore_patterns = ["(.sig|.pem|.apk)$", "musl", "static"]

["zyedidia/micro"]
    upgrade_only = false
    show_hash = true
    asset_filters = [ "static", ".tar.gz" ]
    target = "~/.local/bin/micro"
```

By using the configuration above, you could run the following command to download the latest release of `micro`:

```bash
zeget zyedidia/micro
```

Without the configuration, you would need to run the following command instead:

```bash
export EGET_GITHUB_TOKEN=ghp_1234567890 &&\
zeget zyedidia/micro --to ~/.local/bin/micro --sha256 --asset static --asset .tar.gz
```

## FAQ

### How is this different from a package manager?

zeget only downloads pre-built binaries uploaded to GitHub by the developers of
the repository. zeget does not "install" executables by placing them
in system-wide directories (such as `/usr/local/bin`) unless instructed.
zeget works best for installing software that comes as a single binary with no
additional files needed (CLI tools made in Go, Rust, or Haskell tend to fit
this description).

### Does zeget keep track of installed binaries?

zeget does maintain a cache containing information about installed binaries.
In general, however the cache items expire after a certain amount of time and
are automatically removed.

### Is this secure?

Eget does not run any downloaded code -- it just finds executables from GitHub
releases and downloads/extracts them. If you trust the code you are downloading
(i.e. if you trust downloading pre-built binaries from GitHub) then using Eget
is perfectly safe. If Eget finds a matching asset ending in `.sha256` or
`.sha256sum`, the SHA-256 checksum of your download will be automatically
verified. You can also use the `--sha256` or `--verify-sha256` options to
manually verify the SHA-256 checksums of your downloads (checksums are provided
in an alternative manner by your download source).

### Does this work only for GitHub repositories?

At the moment Eget supports searching GitHub releases, direct URLs, and local
files. If you provide a direct URL instead of a GitHub repository, Eget will
skip the detection phase and download directly from the given URL. If you
provide a local file, Eget will skip detection and download and just perform
extraction from the local file.

### How can I make my software compatible with Eget?

Eget should work out-of-the-box with many methods for releasing software, and
does not require that you build your release process for Eget in particular.
However, here are some rules that will guarantee compatibility with Eget.

- Provide your pre-built binaries as GitHub release assets.
- Format the system name as `OS_Arch` and include it in every pre-built binary
  name. Supported OSes are `darwin`/`macos`, `windows`, `linux`, `netbsd`,
  `openbsd`, `freebsd`, `android`, `illumos`, `solaris`, `plan9`. Supported
  architectures are `amd64`, `i386`, `arm`, `arm64`, `riscv64`.
- If desired, include either `*.sha256` files for each asset that contains the SHA-256
  checksum, or a `checksums.txt` that contains the SHA-256 checksums for all files in
  the asset archive. These checksums will be automatically verified by zeget.
- Include only a single executable or appimage per system in each release archive.
- Use `.tar.gz`, `.tgz`, `.tar.bz2`, `.tar.xz`, `.tar`, or `.zip` for archives. You may
  also directly upload the executable without an archive, or a compressed executable
  ending in `.gz`, `.bz2`, or `.xz`.

### Does this work with monorepos?

Yes, you can pass a tag or tag identifier with the `--tag TAG` option. If no
tag exactly matches, Eget will look for the latest release with a tag that
contains `TAG`. So if your repository contains releases for multiple different
projects, just pass the appropriate tag (for the project you want) to Eget, and
it will find the latest release for that particular project (as long as
releases for that project are given tags that contain the project name).

## Contributing

Please see [CONTRIBUTING](.github/CONTRIBUTING.md) for details.

## Security Vulnerabilities

Please review [our security policy](../../security/policy) on how to report security vulnerabilities.

## Credits

- [Patrick Organ](https://github.com/patinthehat)
- [Zachary Yedidia](https://github.com/zyedidia)
- [All Contributors](../../contributors)

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
