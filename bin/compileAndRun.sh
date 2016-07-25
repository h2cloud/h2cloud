#! /bin/bash

if go install mainpkg then
    echo "Running..."
    ./mainpkg
else
    echo "Compile Error."
fi
