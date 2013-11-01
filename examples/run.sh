#!/bin/sh

ROOT=`readlink -f $0|xargs dirname`
set -e
for i in $ROOT/*.go; do
	go run $i
done

echo "all pass"
