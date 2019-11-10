#!/bin/bash

command cd $(dirname "$1") || (command echo "could not change dir" && command exit 1)

testedCount=0
mCount=0

for file in $(find -type f -name '*.go'); do
  if [[ ! $file =~ [._]test\.go ]] && [[ ! $file =~ [._]bench ]]; then
    testFile="${file/.go/}_test.go"
    package="$(grep -E 'package .*' < $file | grep -E -o '\w+$')"
    cat <<EOF

$file [$package]
---------------------------------
EOF
    for method in $(grep -E -o '^\s*func\s+(\([^)(]+\)\s+)?[A-Z]\w*' < $file | grep -E -o '\w+$'); do
      mCount=$((mCount+1))
      testPattern="Test.*${method}"
      if [[ ! -f "$testFile" ]] || [[ $(grep -E "$testPattern" < "$testFile") == "" ]]; then
        echo -e "[TEST NOT FOUND] func $package.$method"
      else
        testedCount=$((testedCount+1))
      fi
    done
  fi
done

echo "COVERAGE $testedCount/$mCount"
