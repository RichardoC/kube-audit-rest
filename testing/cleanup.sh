#!/bin/bash

rm -rf tmp

kubectl delete -f k8s/webhook.yaml
kubectl delete -f k8s/namespace.yaml

rm kube-rest-audit
