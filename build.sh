#!/bin/bash
set -euxo pipefail

GITCOMMIT=`git describe --always`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
LDFLAGS="-X main.GitCommit=${GITCOMMIT} -X main.GitDate=${GITDATE}"
BINARY=bin/eth2-client-metrics-exporter
BUILDPLATFORM=${BUILDPLATFORM:-"linux/amd64"}
TARGETPLATFORM=${TARGETPLATFORM:-"linux/amd64"}

if [ "$BUILDPLATFORM" != "$TARGETPLATFORM" ]; then
    echo "Cross-compiling to $TARGETPLATFORM"
    target_platform=(${TARGETPLATFORM//\// })
    export GOOS=${target_platform[0]}
    export GOARCH=${target_platform[1]}
    if [ "${#target_platform[@]}" -gt 2 ]; then
        export GOARM=${target_platform[2]//v}
    fi
else
    echo "Compiling to $TARGETPLATFORM"
fi

go build --ldflags="${LDFLAGS}" -o "${BINARY}"
