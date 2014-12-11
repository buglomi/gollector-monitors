#!/usr/bin/env bash

BUILD_DIRS=`go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./src/... | awk '{ print substr($1, 41) }'`

subdirs=$BUILD_DIRS

cmd_prefix=""
if [ "$#" -ge 1 ] && [ "$1" == "godep" ]; then
  cmd_prefix="godep"
  shift
fi

if [ "$#" -ge 1 ] && [ "$1" == "clean" ]; then
  action="cleaning"
  cmd="go clean"
elif [ "$#" -eq 1 ] && [ "$1" == "test" ]; then
  action="testing (without race detector)"
  cmd="go test ."
  subdirs=$TEST_DIRS
elif [ "$#" -ge 2 ] && [ "$1" == "test" ] && [ "$2" == "-race" ]; then
  action="testing with race detector"
  cmd="go test -race ."
  subdirs=$TEST_DIRS
else
  action="building"
  cmd="go build"
fi

cmd="$cmd_prefix $cmd"

for dir in ${subdirs} ; do
    echo "--- $action $dir"

  pushd $dir > /dev/null; echo "- $cmd"; $cmd; popd > /dev/null
done
