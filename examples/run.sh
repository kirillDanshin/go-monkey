#!/bin/sh

ROOT=`dirname $0`
set -e
for i in $ROOT/*.go; do
	echo $i
	go run $i
done

echo "all pass"
