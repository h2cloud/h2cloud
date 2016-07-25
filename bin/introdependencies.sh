#! /bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

export GOPATH="$DIR/../"
echo "Successfully change path to $GOPATH"

go get github.com/ncw/swift
echo "Successfully get package github.com/ncw/swift"

go get github.com/levythu/gurgling
echo "Successfully get package github.com/levythu/gurgling"
