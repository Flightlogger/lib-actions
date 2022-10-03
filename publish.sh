#!/bin/sh
git tag $1
git push --tags
GOPROXY=proxy.golang.org go list -m github.com/Flightlogger/lib-actions@$1
