#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/waffles/ man | gzip -c -9 >manpages/waffles.1.gz
