.PHONY: build_deps build package_deps package clobber
.DEFAULT_GOAL := build

build_deps:
	@type go >/dev/null 2>&1 || \
		{ echo >&2 "I require go but it is not installed.  Aborting."; exit 1; }

build: build_deps
	go build -o bin/slacknimate ./cmd/slacknimate

package_deps:
	@type goreleaser >/dev/null 2>&1 || \
		{ echo >&2 "I require goreleaser but it is not installed.  Aborting."; exit 1; }

package: package_deps
	goreleaser release --rm-dist --skip-publish

clobber:
	rm -rf dist bin
