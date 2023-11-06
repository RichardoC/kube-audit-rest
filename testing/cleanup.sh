#!/bin/bash


# Default to root if no .git missing
ROOT=$(git rev-parse --show-toplevel || echo '.' )

cd $ROOT

rm -rf tmp

kubectl delete -f k8s/webhook.yaml
kubectl delete -f k8s/namespace.yaml

rm kube-audit-rest
