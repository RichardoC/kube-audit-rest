#!/bin/bash

set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)

cd $ROOT

export COMMIT="$(git rev-parse HEAD)"

# For storing temporary files that version control will ignore, such as certs
mkdir -p tmp

# Create required certs
testing/certs.sh

nerdctl build -f Dockerfile-distroless . --namespace k8s.io -t "ghcr.io/richardoc/kube-audit-rest:${COMMIT}-distroless"
# nerdctl build -f Dockerfile-alpine . --namespace k8s.io -t "richardoc/kube-audit-rest:${COMMIT}-alpine"

kubectl -n kube-audit-rest apply -f k8s/namespace.yaml

# Upload the TLS cert and replace if exists
kubectl -n kube-audit-rest create secret tls kube-audit-rest --cert=tmp/server.crt --key=tmp/server.key --dry-run=client -oyaml | kubectl -n kube-audit-rest apply -f -

# Substitute in the correct image tag
cat k8s/deployment.yaml | envsubst | kubectl -n kube-audit-rest apply -f -

kubectl -n kube-audit-rest apply -f k8s/service.yaml

# Substitute in the correct CA into the webhook
export CABUNDLEB64="$(cat tmp/rootCA.pem | base64 | tr -d '\n')"
cat k8s/webhook.yaml | envsubst | kubectl -n kube-audit-rest apply -f -
unset CABUNDLEB64

kubectl -n kube-audit-rest rollout restart deployment/kube-audit-rest
