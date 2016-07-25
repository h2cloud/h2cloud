#! /bin/bash

CTN="dr-1M"
PTH="testenv/maindir/"
PFX="obj"

./batchPutter -container $CTN -path $PTH -prefix $PFX -from 0 -to 10000
