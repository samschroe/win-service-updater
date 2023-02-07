#!/bin/bash

UNFORMATTED=$(gofmt -l -e .)

if [[ -n "$UNFORMATTED" ]]; then
    echo "The following files need formatted:"
    for f in $UNFORMATTED; do
        echo "  $f"
    done
    exit 1
fi
