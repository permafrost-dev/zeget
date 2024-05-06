# zeget: easy pre-built binary installation

![Codecov](https://img.shields.io/codecov/c/gh/permafrost-dev/eget?style=flat-square&logo=codecov&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/permafrost-dev/eget)](https://goreportcard.com/report/github.com/permafrost-dev/eget)
[![Release](https://img.shields.io/github/release/permafrost-dev/eget.svg?label=Release)](https://github.com/permafrost-dev/eget/releases)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/permafrost-dev/eget/blob/master/LICENSE)

**zeget** is the best way to easily get pre-built binaries for your favorite
tools; it downloads and extracts pre-built binaries from releases on GitHub.

To use it, provide a repository and zeget will search through the assets from the
latest release to find a suitable binary for your system. If successful, the asset is
downloaded and extracted to the current directory.

![Eget Demo](https://github.com/zyedidia/blobs/blob/master/eget-demo.gif?raw=true)

zeget is a fork of [eget](https://github.com/zyedidia/eget) with many bug fixes, improvements,
and new features. It is a drop-in replacement for eget and is fully compatible with the original.

zeget has a number of detection mechanisms and should work out-of-the-box with
most software that is distributed via single binaries on GitHub releases. First
try using zeget on your software, it is likely that is will "just work".
Otherwise, see the FAQ for a clear set of rules to make your software compatible with zeget.

For more in-depth documentation, see [DOCS.md](DOCS.md).

## Examples

```sh
zeget zyedidia/micro nightly
zeget jgm/pandoc --to /usr/local/bin
zeget junegunn/fzf
zeget neovim/neovim
zeget ogham/exa --asset ^musl
zeget --system darwin/amd64 sharkdp/fd
zeget BurntSushi/ripgrep
zeget -f eget.1 permafrost-dev/eget
zeget zachjs/sv2v
zeget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5
zeget --all --file '*' ActivityWatch/activitywatch
```

## How to get zeget

Before you can get anything, you have to get Eget. If you already have Eget and want to upgrade, use `eget permafrost-dev/eget`.

### Quick-install script

```sh
curl -o install.sh https://raw.githubusercontent.com/permafrost-dev/eget/main/scripts/install.sh
sha256sum install.sh # verify with hash below
bash install.sh
```

Or alternatively (less secure):

```sh
curl https://raw.githubusercontent.com/permafrost-dev/eget/main/scripts/install.sh | sh
```

You can then place the downloaded binary in a location on your `$PATH` such as `/usr/local/bin`.

To verify the script, use `sha256sum install.sh` after downloading; the sha256 checksum is:

> <code data-id="installer-checksum">7ab91ff8a3d0788d92ec96212aadbb3019bc56f514b03debda701cdb465eb970  install.sh</code><br>

### Pre-built binaries

Pre-built binaries are available on the [releases](https://github.com/permafrost-dev/eget/releases) page.

### From source

Install the latest released version:

```sh
go install github.com/permafrost-dev/eget@latest
```

or install from HEAD:

```sh
git clone https://github.com/permafrost-dev/eget
cd eget
make build # or go build (produces incomplete version information)
```

A man page can be generated by cloning the repository and running `make eget.1`
(requires pandoc). You can also use `zeget` to download the man page: `zeget -f eget.1 permafrost-dev/eget`.

## Usage

The `TARGET` argument passed to Eget should either be a GitHub repository,
formatted as `user/repo`, in which case Eget will search the release assets, a
direct URL, in which case Eget will directly download and extract from the
given URL, or a local file, in which case Eget will extract directly from the
local file.

If Eget downloads an asset called `xxx` and there also exists an asset called
`xxx.sha256` or `xxx.sha256sum`, Eget will automatically verify that the
SHA-256 checksum of the downloaded asset matches the one contained in that
file, and abort installation if a mismatch occurs.

Likewise, if there is a `checksums.txt` asset in the release, Eget will attempt
to use this file to verify the SHA-256 checksum of the downloaded asset.

When installing an executable, Eget will place it in the current directory by
default. If the environment variable `EGET_BIN` is non-empty, Eget will
place the executable in that directory.

Directories can also be specified as files to extract, and all files within
them will be extracted. For example:

```sh
eget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz --file go --to ~/go1.21.0
```

GitHub limits API requests to 60 per hour for unauthenticated users. If you
would like to perform more requests (up to 5,000 per hour), you can set up a
personal access token and assign it to an environment variable named either
`GITHUB_TOKEN` or `EGET_GITHUB_TOKEN` when running Eget. If both are set,
`EGET_GITHUB_TOKEN` will take precedence. Eget will read this variable and
send the token as authorization with requests to GitHub. It is also possible
to read the token from a file by using `@/path/to/file` as the token value.

```sh
Usage:
  eget [OPTIONS] TARGET

Application Options:
  -t, --tag=           tagged release to use instead of latest
      --pre-release    include pre-releases when fetching the latest version
      --source         download the source code for the target repo instead of a release
      --to=            move to given location after extracting
  -s, --system=        target system to download for (use "all" for all choices)
  -f, --file=          glob to select files for extraction
      --all            extract all candidate files
  -q, --quiet          only print essential output
  -d, --download-only  stop after downloading the asset (no extraction)
      --upgrade-only   only download if release is more recent than current version
  -a, --asset=         download a specific asset containing the given string; can be specified multiple times for additional filtering; use ^ for anti-match
      --sha256         show the SHA-256 hash of the downloaded asset
      --verify-sha256= verify the downloaded asset checksum against the one provided
      --rate           show GitHub API rate limiting information
  -r, --remove         remove existing target files before downloading
  -v, --version        show version information
  -h, --help           show this help message
  -D, --download-all   download all projects defined in the config file
  -k, --disable-ssl    disable SSL verification for download
```

## Configuration

Eget can be configured using a TOML file located at `~/.eget.toml` or it will fallback to the expected `XDG_CONFIG_HOME` directory of your os. Alternatively,
the configuration file can be located in the same directory as the Eget binary or the path specified with the environment variable `EGET_CONFIG`.

Both global settings can be configured, as well as setting on a per-repository basis.

Sections can be named either `global` or `"owner/repo"`, where `owner` and `repo`
are the owner and repository name of the target repository (not that the `owner/repo`
format is quoted).

For example, the following configuration file will set the `--to` flag to `~/bin` for
all repositories, and will set the `--to` flag to `~/.local/bin` for the `permafrost-dev/micro`
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
    target = "./test"

["zyedidia/micro"]
    upgrade_only = false
    show_hash = true
    asset_filters = [ "static", ".tar.gz" ]
    target = "~/.local/bin/micro"
```

By using the configuration above, you could run the following command to download the latest release of `micro`:

```bash
eget zyedidia/micro
```

Without the configuration, you would need to run the following command instead:

```bash
export EGET_GITHUB_TOKEN=ghp_1234567890 &&\
eget zyedidia/micro --to ~/.local/bin/micro --sha256 --asset static --asset .tar.gz
```

## FAQ

### How is this different from a package manager?

Eget only downloads pre-built binaries uploaded to GitHub by the developers of
the repository. It does not maintain a central list of packages, nor does it do
any dependency management. Eget does not "install" executables by placing them
in system-wide directories (such as `/usr/local/bin`) unless instructed, and it
does not maintain a registry for uninstallation. Eget works best for installing
software that comes as a single binary with no additional files needed (CLI
tools made in Go, Rust, or Haskell tend to fit this description).

### Does Eget keep track of installed binaries?

Eget does not maintain any sort of manifest containing information about
installed binaries. In general, Eget does not maintain any state across
invocations. However, Eget does support the `--upgrade-only` option, which
will first check `EGET_BIN` to determine if you have already downloaded the
tool you are trying to install -- if so it will only download a new version if
the GitHub release is newer than the binary on your file system.

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
- If desired, include `*.sha256` files for each asset, containing the SHA-256
  checksum of each asset. These checksums will be automatically verified by
  Eget.
- Include only a single executable or appimage per system in each release archive.
- Use `.tar.gz`, `.tar.bz2`, `.tar.xz`, `.tar`, or `.zip` for archives. You may
  also directly upload the executable without an archive, or a compressed
  executable ending in `.gz`, `.bz2`, or `.xz`.

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
