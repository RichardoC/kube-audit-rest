#!/bin/bash


ROOT=$(git rev-parse --show-toplevel)

cd $ROOT

rm -rf tmp

kubectl delete -f k8s/webhook.yaml
kubectl delete -f k8s/namespace.yaml

rm kube-audit-rest
