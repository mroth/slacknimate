.PHONY: install_deps build package clobber
.DEFAULT_GOAL := build

build:
	go build -o bin/slacknimate

install: 
	go install .

uninstall:
	rm $(GOPATH)/bin/slacknimate

package_deps:
	go get github.com/laher/goxc

package: package_deps build
	goxc -pv=`./bin/slacknimate -v | cut -d' ' -f3` \
	     --resources-include="README*,LICENSE*,examples" \
	     -d="builds" -bc="linux,!arm darwin windows" xc archive rmbin

clobber:
	rm -rf builds bin
