#! /bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

export GOPATH="$DIR/../"
echo "Successfully change path to $GOPATH"

export SLCHOME="$DIR/../testenv/"
echo "Successfully change SLCHOME to $SLCHOME"
