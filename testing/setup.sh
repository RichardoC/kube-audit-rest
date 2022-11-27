#!/bin/bash

export COMMIT="$(git rev-parse HEAD)"

# For storing temporary files that version control will ignore, such as certs
mkdir -p tmp

# Create required certs
testing/certs.sh

nerdctl build -f Dockerfile . --rm=false --namespace k8s.io -t "richardoc/kube-rest-audit:${COMMIT}"

kubectl -n kube-rest-audit apply -f k8s/namespace.yaml

# Upload the TLS cert and replace if exists
kubectl -n kube-rest-audit create secret tls kube-rest-audit --cert=tmp/server.crt --key=tmp/server.key --dry-run=client -oyaml | kubectl -n kube-rest-audit apply -f -

# Substitute in the correct image tag
cat k8s/deployment.yaml | envsubst | kubectl -n kube-rest-audit apply -f -

kubectl -n kube-rest-audit apply -f k8s/service.yaml

# Substitute in the correct CA into the webhook
export CABUNDLEB64="$(cat tmp/rootCA.pem | base64 -w0| tr -d '\n')"
cat k8s/webhook.yaml | envsubst | kubectl -n kube-rest-audit apply -f -
unset CABUNDLEB64

kubectl -n kube-rest-audit rollout restart deployment/kube-rest-audit