.PHONY: install_deps build package clobber
.DEFAULT: build

build:
	go build -o bin/slacknimate

install_deps:
	go get github.com/laher/goxc

package: install_deps build
	goxc -pv=`./bin/slacknimate -v | cut -d' ' -f3` \
	 		 -d="builds" -bc="linux,!arm darwin windows" xc archive rmbin

clobber:
	rm -rf builds
