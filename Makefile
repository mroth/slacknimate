.PHONY: install_deps build package clobber
.DEFAULT: build

build:
	go build -o bin/slacknimate

package_deps:
	go get github.com/laher/goxc

package: package_deps build
	goxc -pv=`./bin/slacknimate -v | cut -d' ' -f3` \
	 		 -d="builds" -bc="linux,!arm darwin windows" xc archive rmbin

clobber:
	rm -rf builds
