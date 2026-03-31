#!/usr/bin/env bash
go test -coverpkg=./... -coverprofile=coverage.out ./tests/... -count=1 && go tool cover -html=coverage.out -o coverage.html
