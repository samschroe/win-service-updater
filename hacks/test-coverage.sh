#!/bin/bash

go test -coverprofile=/tmp/profile.out ./updater/...
go tool cover -html=/tmp/profile.out
