#!/bin/bash

GOOSES=("windows")
TAGS=("debug")

FAILED=0

function handle_exit {
    if [[ $1 -eq 0 ]]; then
        echo -e '\033[0;32mPASSED\033[0m'
    else
        FAILED=$1
        echo -e '\033[0;31mFAILED\033[0m'
        echo "${@:2}" 
        echo
    fi
}

for OS in "${GOOSES[@]}"; do
    echo -n "Vetting \"$OS\" without tags: "
    RESULT=$(GOOS=$OS go vet -tags "full" ./... 2>&1)
    handle_exit $? $RESULT

    for TAG in "${TAGS[@]}"; do
        echo -n "Vetting \"$OS\" with \"$TAG\" tag: "
        RESULT=$(GOOS=$OS go vet -tags "full $TAG" ./... 2>&1)
        handle_exit $? $RESULT
    done

done

exit "$FAILED"
