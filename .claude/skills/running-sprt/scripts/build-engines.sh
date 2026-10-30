#!/usr/bin/env bash
# Builds base and dev UCI engines for an SPRT and prints their bench signatures.
# Usage: build-engines.sh <scratch-dir> [base-ref] [dev-ref]
# base-ref defaults to HEAD. Without dev-ref, dev is the working tree.
set -euo pipefail

if [ $# -lt 1 ]; then
	echo "usage: build-engines.sh <scratch-dir> [base-ref] [dev-ref]" >&2
	exit 2
fi

scratch=$1
base_ref=${2:-HEAD}
dev_ref=${3:-}
repo=$(git rev-parse --show-toplevel)

build_tree() {
	go -C "$1" build -trimpath -buildvcs=false -o "$scratch/$2-uci.exe" ./cmd/uci
}

build_ref() {
	local name=$1 ref=$2
	local tree="$scratch/$name-tree"
	rm -rf "$tree"
	mkdir -p "$tree"
	git -C "$repo" archive "$ref" | tar -x -C "$tree"
	build_tree "$tree" "$name"
}

mkdir -p "$scratch"
build_ref base "$base_ref"

if [ -n "$dev_ref" ]; then
	build_ref dev "$dev_ref"
else
	build_tree "$repo" dev
fi

if cmp -s "$scratch/base-uci.exe" "$scratch/dev-uci.exe"; then
	echo "base and dev are identical builds, so the test would be meaningless" >&2
	exit 1
fi

echo "base: $("$scratch/base-uci.exe" bench | tail -n 1)"
echo "dev:  $("$scratch/dev-uci.exe" bench | tail -n 1)"
