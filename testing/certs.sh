#!/bin/bash

set -eux

ROOT=$(git rev-parse --show-toplevel)

cd $ROOT

CONF_DIR=$(pwd)/testing

TMP=$(pwd)/tmp
echo $TMP
cd $TMP

openssl genrsa -out rootCA.key 2048 &> /dev/null

openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1460 -out rootCA.pem -config $CONF_DIR/ca.cnf &> /dev/null

openssl req -new -nodes -sha256 -out server.csr -newkey rsa:2048 -keyout server.key -config $CONF_DIR/server.csr.cnf &> /dev/null

openssl x509 -req -in server.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out server.crt -days 500 -sha256 -extensions v3_req -extfile $CONF_DIR/server.csr.cnf &> /dev/null
