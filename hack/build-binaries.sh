#!/bin/bash
set -e
set -o pipefail

HACKDIR=$( dirname $( readlink -f "${BASH_SOURCE[0]}" ) )
BASEDIR="${HACKDIR}/.."

export GO111MODULE=on
export GOPROXY=off
export GOFLAGS=-mod=vendor
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

for name in $( ls "${BASEDIR}/cmd/" ); do
	path="${BASEDIR}/cmd/${name}"
	if [ ! -d "${path}" ]; then
		continue
	fi
	echo "go build -o \"${BASEDIR}/_output/${name}\" \"${path}\""
	go build -o "${BASEDIR}/_output/${name}${suffix}" "${path}"
done
