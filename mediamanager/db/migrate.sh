#!/bin/bash
go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate \
    -database postgres:/// \
    -path "$(dirname "$0")/migrations" \
    $@