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

if [[ "$(uname -m)" == 'x86_64' ]]
then
    # Run current server with those local certs on port 9090
    # With race detection on x86_64
    echo "Also doing race detection"
    go run -race . --cert-filename=./tmp/server.crt --cert-key-filename=./tmp/server.key --server-port=9090 --logger-filename=./tmp/kube-audit-rest.log &
else
    # Run current server with those local certs on port 9090
    go run . --cert-filename=./tmp/server.crt --cert-key-filename=./tmp/server.key --server-port=9090 --logger-filename=./tmp/kube-audit-rest.log &
fi

# Wait for server to run
while ! nc -z localhost 9090; do   
  sleep 1 # wait for 1/10 of the second before check again
done

go run testing/locally/main.go

export TEST_EXIT="$?"

sleep 2 # Scientific way of waiting for the file to be written as async...

# Removing backgrounded process
pkill kube-audit-rest

# Ensure every line has a requestReceivedTimestamp 
if [ "$(cat tmp/kube-audit-rest.log | grep -c "requestReceivedTimestamp")" -ne "$(wc -l tmp/kube-audit-rest.log | cut -d ' '  -f 1)" ]; then
    echo "output not as expected, not all lines contain requestReceivedTimestamp"
    exit 1
fi


# Sort audit log by uid as it's the only guaranteed field, and kube-audit-rest doesn't guarantee request ordering
# Removing the requestReceivedTimestamp timestamp as it's not deterministic
cat tmp/kube-audit-rest.log | jq -s -c '. | sort_by(.request.uid)| del(.[].requestReceivedTimestamp)| .[]' > tmp/kube-audit-rest-sorted.log

diff testing/locally/data/kube-audit-rest-sorted.log tmp/kube-audit-rest-sorted.log && [ "$TEST_EXIT" -eq "0" ] && echo "Test passed" || bash -c 'echo "output not as expected" && exit 255'
