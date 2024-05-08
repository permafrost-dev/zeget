---
title: zeget
section: 1
header: zeget Manual
---

## NAME

  zeget - easily install prebuilt binaries from GitHub

## SYNOPSIS

  zeget `[--version] [--help] [OPTIONS] TARGET`

## DESCRIPTION

  zeget is a tool for downloading and extracting prebuilt binaries from releases
  on GitHub. To use it, provide a repository and zeget will search through the
  assets from the latest release in an attempt to find a suitable prebuilt
  binary for your system. If one is found, the asset will be downloaded and
  zeget will extract the binary to the current directory. zeget should only be
  used for installing simple, static prebuilt binaries, where the extracted
  binary is all that is needed for installation. For more complex installation,
  you may use the `--download-only` option, and perform extraction manually.

  The **`PROJECT`** argument passed to zeget should either be a GitHub
  repository, formatted as **`user/repo`**, in which case zeget will search the
  release assets, a direct URL, in which case zeget will directly download and
  extract from the given URL, or a local file, in which case zeget will extract
  directly from the local file.

  If zeget downloads an asset called `xxx` and there also exists an asset called
  `xxx.sha256` or `xxx.sha256sum`, zeget will automatically verify that the
  SHA-256 checksum of the downloaded asset matches the one contained in that
  file, and abort installation if a mismatch occurs.

  When installing an executable, zeget will place it in the current directory by
  default. If the environment variable **`zeget_BIN`** is non-empty, zeget will
  place the executable in that directory. The `--to` flag may also be used to
  customize the install location.

  Directories can also be specified as files to extract, and all files within
  them will be extracted. For example:

      zeget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5

  GitHub limits API requests to 60 per hour for unauthenticated users. If you
  would like to perform more requests (up to 5,000 per hour), you can set up a
  personal access token and assign it to an environment variable named either
  **`GITHUB_TOKEN`** or **`ZEGET_GITHUB_TOKEN`** when running zeget. If both are set,
  **`ZEGET_GITHUB_TOKEN`** will take precedence. zeget will read this variable and
  send the token as authorization with requests to GitHub. It is also possible to
  read the token from a file by using `@/path/to/file` as the token value.

  The behavior of zeget is configurable in a number of ways via options.
  Documentation for these options is provided below.

## OPTIONS

  `-t, --tag=`

:    Use the given tagged release instead of the latest release. If the project does not have a tag that matches exactly, zeget will look for a tag that contains the   given string, and use the latest one. Example: **`zeget -t nightly zyedidia/micro`**.

  `--pre-release`

:    Include pre-releases when fetching the latest version. This will get the latest overall release, even if it is a pre-release.

  `--source`

:    Download the source code for the repository (only works for GitHub repositories) rather than a release. Downloads from the "master" branch by default. Use `--tag` to download a different tag or branch.

  `--to=`

:    Move the executable to the given name after extraction. If the name is `-`, it the data will be written to stdout. Example: **`zeget zyedidia/micro --to /usr/local/bin`**. Example: **`zeget --asset nvim.appimage --to nvim neovim/neovim`**.

  `-s, --system=`

:    Use the given system as the target instead of the host. Systems follow the notation 'OS/Arch', where OS is a valid OS (darwin, windows, linux, netbsd, openbsd, freebsd, android, illumos, solaris, plan9), and Arch is a valid architecture (amd64, 386, arm, arm64, riscv64). If the special value **all** is used, all possibilities are given and the user must select manually. Example: **`zeget -s darwin/amd64 zyedidia/micro`**.

  `-f, --file=`

:    Extract the file that matches the given glob. You may want use this option to extract non-binary files. Example: **`zeget -f LICENSE zyedidia/micro`**.

  `--all`

:    Extract all candidate files.

  `-q, --quiet`

:    Only print essential output.

  `--download-only`

:    Stop after downloading the asset. This prevents zeget from performing extraction, allowing you to perform manual installation after the asset is downloaded.

  `--download-all`

:   Download all projects defined in the configuration file.

   --upgrade-only

:    Only download the asset if the release is more recent than an existing asset with the same name in `$zeget_BIN`, or the current directory if `$zeget_BIN` is not defined.

  `-a, --asset=`

:    Download a specific asset containing the given string. If there is an exact match with an asset, that asset is used regardless (except when using `^`). If the argument begins with a `^`, then any asset that does not match the argument is a candidate. This option can be specified multiple times for additional filtering. Example: **`zeget --asset nvim.appimage neovim/neovim`**. Example **`zeget --download-only --asset amd64.deb --asset musl sharkdp/bat`**. If the assets are filterable using the `--system` detector (i.e., if applying the detector does not remove all candidates), the system detector is applied. Use `--system all` to always consider all assets.

  `--sha256`

:    Show the SHA-256 hash of the downloaded asset. This can be used to verify that the asset is not corrupted.

  `--verify-sha256=`

:    Verify the SHA-256 hash of the downloaded asset against the one provided as an argument. Similar to `--sha256`, but zeget will do the verification for you.

  `--rate`

:    Show GitHub API rate limiting information.

  `--remove`

:    Remove the target file from `$zeget_BIN` (or the current directory if unset). Note that this flag is boolean, and means zeget will treat `TARGET` as a file to be removed.

  `-k, --disable-ssl`

:    Disable SSL certificate verification for GET requests. Cannot be used in combination with a `GITHUB_TOKEN`.

  `-v, --version`

:    Show version information.

  `-h, --help`

:    Show a help message.

## CONFIGURATION

  zeget can be configured using a TOML file located at `~/.zeget.toml`. Alternatively,
  the configuration file can be located in the same directory as the zeget binary.

  Both global settings can be configured, as well as setting on a per-repository basis.

  Sections can be named either `global` or `"owner/repo"`, where `owner` and `repo`
  are the owner and repository name of the target repository (not that the `owner/repo`
  format is quoted).

  For example, the following configuration file will set the `--to` flag to `~/bin` for
  all repositories, and will set the `--to` flag to `~/.local/bin` for the `zyedidia/micro`
  repository.

    ```toml
      [global]
      to = "~/bin"

      ["zyedidia/micro"]
      to = "~/.local/bin"
    ```

  More complete example configuration:

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

  By using the configuration above, you could run the following command to download
  the latest release of `micro`:
  **`zeget zyedidia/micro`**

  Without the configuration, you would need to run the following command instead:
  **`zeget zyedidia/micro --to ~/.local/bin/micro --sha256 --asset static --asset .tar.gz`**

### Available settings

  `all`

:    Whether to extract all candidate files.

  `asset_filters`

:    An array of partial asset names to filter the available assets for download.

  `download_only`

:    Whether to stop after downloading the asset (no extraction).

  `file`

:    The glob to select files for extraction.

  `github_token`
  
:    GitHub API token to use for requests.

  `quiet`

:    Whether to only print essential output.

  `show_hash`

:    Whether to show the SHA-256 hash of the downloaded asset.

  `system`

:    The target system to download for.

  `target`

:    The directory to move the downloaded file to after extraction.

  `upgrade_only`

:    Whether to only download if release is more recent than current version.

## FOR MAINTAINERS

To guarantee compatibility of your software's pre-built binaries with zeget, you
can follow these rules.

* Provide your pre-built binaries as GitHub release assets.
* Format the system name as `OS_Arch` and include it in every pre-built binary
  name. Supported OSes are `darwin`/`macos`, `windows`, `linux`, `netbsd`, `openbsd`,
  `freebsd`, `android`, `illumos`, `solaris`, `plan9`. Supported architectures
  are `amd64`, `i386`, `arm`, `arm64`, `riscv64`.
* If desired, include `*.sha256` files for each asset, containing the SHA-256
  checksum of each asset. These checksums will be automatically verified by
  zeget.
* Include only a single executable or appimage per system in each release archive.
* Use `.tar.gz`, `.tar.bz2`, `.tar.xz`, `.tar`, or `.zip` for archives. You may
  also directly upload the executable without an archive, or a compressed
  executable ending in `.gz`, `.bz2`, or `.xz`.

If you don't follow these rules, zeget may still work well with your software.
zeget's auto-detection is much more relaxed than what is required by these
rules, but if you follow these rules your software is guaranteed to be
compatible with zeget.

## BUGS

See GitHub Issues: <https://github.com/permafrost-dev/zeget/issues>

## AUTHOR

Patrick Organ <patrick@permafrost.dev>
