#!/usr/bin/env bash
gofumpt -w . && golangci-lint run