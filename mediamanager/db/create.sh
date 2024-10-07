#!/bin/bash
go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate create \
    -dir "$(dirname $0)/migrations" \
    -ext sql \
    -digits 6 \
    -seq \
    $@