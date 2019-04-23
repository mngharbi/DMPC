#!/usr/bin/env bash

touch coverage.txt

for subsystem in $(go list ./...); do
    go test -race -coverprofile=profile.out -covermode=atomic "$subsystem"
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
