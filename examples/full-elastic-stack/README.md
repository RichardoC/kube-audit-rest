
All commands should be run from the root of the repo

## Setting up elastic search

Largely following <https://www.elastic.co/downloads/elastic-cloud-kubernetes>


kubectl create -f https://download.elastic.co/downloads/eck/2.9.0/crds.yaml

kubectl apply -f https://download.elastic.co/downloads/eck/2.9.0/operator.yaml

kubectl apply -f examples/full-elastic-stack/k8s/elastic-cluster.yaml



Wait for workloads to start working

Port forward kibana in a terminal

kubectl -n example-kube-audit-rest port-forward svc/kibana-kube-audit-rest-kb-http   60443:https


See that kibana is working

Navigate to <https://localhost:60443/app/discover#/> from your browser, and click to ignore the invalid certificate

Use another terminal to get the password to access

echo "username is elastic"
echo "password is $(kubectl -n example-kube-audit-rest get secret example-kube-audit-rest-es-elastic-user -o=jsonpath='{.data.elastic}' | base64 --decode; echo)"

## Set up kube-audit-rest

Using a locked version from 2023-10-05

Deliberately configuring this in a way that *will* block API calls, to show that you can audit all creation/mutations if required, subject to the limitations in the root README.md

## set up certs

./testing/certs.sh

### Upload the TLS cert and replace if exists
kubectl -n example-kube-audit-rest create secret tls kube-audit-rest --cert=tmp/server.crt --key=tmp/server.key --dry-run=client -oyaml | kubectl -n example-kube-audit-rest apply -f -

## Deploy kube-audit-rest
Webhooks are required to serve TLS, so the templating is including the certificate authority so kubernetes trusts our certificate

export CABUNDLEB64="$(cat tmp/rootCA.pem | base64 | tr -d '\n')"
cat examples/full-elastic-stack/k8s/kube-audit-rest.yaml | envsubst | kubectl -n example-kube-audit-rest apply -f -
unset CABUNDLEB64

