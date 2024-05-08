#!/bin/sh

INSTALLER_FN="./install.sh"

if [ -d ".git" ]; then
    INSTALLER_FN="./scripts/install.sh"
fi

NEWHASH=$(sha256sum $INSTALLER_FN | awk '{print $1}')
cat README.md | sed "s|<code data-id=\"installer-checksum\">.*</code><br>|<code data-id=\"installer-checksum\">$NEWHASH</code><br>|g" > README.md.tmp
mv -f README.md.tmp README.md