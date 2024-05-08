#!/bin/sh

# This script installs Eget.
#
# Quick install: `curl https://raw.githubusercontent.com/permafrost-dev/eget/scripts/install.sh | bash`
#
# Acknowledgments:
#   - eget: https://github.com/zyedidia/zyedidia.github.io/blob/master/eget.sh
#   - getmic.ro: https://github.com/benweissmann/getmic.ro

set -e -u

REPONAME="permafrost-dev/eget"

githubLatestTag() {
  finalUrl=$(curl "https://github.com/$1/releases/latest" -s -L -I -o /dev/null -w '%{url_effective}')
  printf "%s\n" "${finalUrl##*v}"
}

platform=''
machine=$(uname -m)

if [ "${GETEGET_PLATFORM:-x}" != "x" ]; then
  platform="$GETEGET_PLATFORM"
else
  case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    "linux")
      case "$machine" in
        "arm64"* | "aarch64"* ) platform='linux_arm64' ;;
        "arm"* | "aarch"*) platform='linux_arm' ;;
        *"86") platform='linux_386' ;;
        *"64") platform='linux_amd64' ;;
      esac
      ;;
    "darwin")
      case "$machine" in
        "arm64"* | "aarch64"* ) platform='darwin_arm64' ;;
        *"64") platform='darwin_amd64' ;;
      esac
      ;;
    "msys"*|"cygwin"*|"mingw"*|*"_nt"*|"win"*)
      case "$machine" in
        *"86") platform='windows_386' ;;
        *"64") platform='windows_amd64' ;;
      esac
      ;;
  esac
fi

if [ "$platform" = "" ]; then
  cat << 'EOM'
/=====================================\\
|      COULD NOT DETECT PLATFORM      |
\\=====================================/
Uh oh! We couldn't automatically detect your operating system.
To continue with installation, please choose from one of the following values:
- linux_arm
- linux_arm64
- linux_386
- linux_amd64
- darwin_amd64
- darwin_arm64
- windows_386
- windows_amd64
Export your selection as the GETEGET_PLATFORM environment variable, and then
re-run this script.
For example:
  $ export GETEGET_PLATFORM=linux_amd64
EOM
  printf "  $ curl https://raw.githubusercontent.com/%s/eget | bash\n" "$REPONAME"
  exit 1
fi

TAG=$(githubLatestTag $REPONAME)
extension='tar.gz'

if [ "$platform" = "windows_amd64" ] || [ "$platform" = "windows_386" ]; then
  extension='zip'
fi

printf "Detected platform: %s\n" "$platform"
printf "Latest Version: %s\n" "$TAG"
printf "Downloading https://github.com/%s/releases/download/v%s/eget-%s-%s.%s\n" "$REPONAME" "$TAG" "$TAG" "$platform" "$extension"

curl -L "https://github.com/$REPONAME/releases/download/v$TAG/eget-$TAG-$platform.$extension" > "eget.$extension"

case "$extension" in
  "zip") unzip -j "eget.$extension" -d "eget-$TAG-$platform" ;;
  "tar.gz") tar -xvzf "eget.$extension" "eget-$TAG-$platform/eget" ;;
esac

mv "eget-$TAG-$platform/eget" ./eget
chmod +x ./eget

rm "eget.$extension"
rm -rf "eget-$TAG-$platform"

cat <<-'EOM'
Eget has been downloaded to the current directory.
You can run it with:
./eget
EOM