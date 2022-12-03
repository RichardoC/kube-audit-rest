#!/bin/bash

set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)

cd $ROOT

# Ensure tmp folder exists
mkdir -p tmp/

# Ensure local certs exist
testing/certs.sh


echo "Removing old test file"
rm -f tmp/kube-audit-rest.log;

echo "Removing old servers if still running"
pkill kube-audit-rest || echo "No old server running"

# Run current server with those local certs on port 9090
go run . --cert-filename=./tmp/server.crt --cert-key-filename=./tmp/server.key --server-port=9090 --logger-filename=./tmp/kube-audit-rest.log &

sleep 5 # Scientific way of waiting for the server to be ready...

go run testing/locally/main.go

export TEST_EXIT="$?"

sleep 2 # Scientific way of waiting for the file to be written as async...

# Removing backgrounded process
pkill kube-audit-rest

# Sort audit log by uid as it's the only guaranteed field, and kube-audit-rest doesn't guarantee request ordering
cat tmp/kube-audit-rest.log | jq -s -c '. | sort_by(.request.uid) | .[]' > tmp/kube-audit-rest-sorted.log

diff testing/locally/data/kube-audit-rest-sorted.log tmp/kube-audit-rest-sorted.log && [ "$TEST_EXIT" -eq "0" ] && echo "Test passed" || bash -c 'echo "output not as expected" && exit 255'
