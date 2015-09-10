#!/bin/sh

GIT_TAG=`git describe --always --dirty`
BTIME=`date -u +%s`

rm -fr ./build
# -w and -s diasables debugging stuff leading to a
# reduction of binaries sizes
#godep go build -ldflags "-w -X main.Version=${GIT_TAG} -X main.BuildTime=${BTIME}" -o ./build/bin/twiph
go build -ldflags "-w -X main.Version=${GIT_TAG} -X main.BuildTime=${BTIME}" -o ./build/bin/twiph
tar -czvf ./build/linux_X86-64.tar.gz ./build/bin
# Build for Mac OSX
env GOOS=darwin GOARCH=amd64 go build -ldflags "-w -X main.Version=${GIT_TAG} -X main.BuildTime=${BTIME}" -o ./build/bin/twiph
tar -czvf ./build/darwin_X86-64.tar.gz ./build/bin
