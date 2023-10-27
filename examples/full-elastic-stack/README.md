
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

./examples/full-elastic-stack/certs/certs.sh


### Upload the TLS cert and replace if exists
kubectl -n example-kube-audit-rest create secret tls kube-audit-rest --cert=./tmp/full-elastic-stack/server.crt --key=tmp/full-elastic-stack/server.key --dry-run=client -oyaml | kubectl -n example-kube-audit-rest apply -f -

## Deploy kube-audit-rest
Webhooks are required to serve TLS, so the templating is including the certificate authority so kubernetes trusts our certificate

export CABUNDLEB64="$(cat tmp/full-elastic-stack/rootCA.pem | base64 | tr -d '\n')"
cat examples/full-elastic-stack/k8s/kube-audit-rest.yaml | envsubst | kubectl -n example-kube-audit-rest apply -f -
unset CABUNDLEB64

## Do some api calls so you have something to look at
kubectl create ns test-namespace
kubectl -n test-namespace create service account abc
kubectl delete namespace test-namespace

## View the data in elastic search via kibana
Navigate to <https://127.0.0.1:60443/app/discover#/> provided the port forward from earlier is still running, or restart it if required.

Create a data view 
Name: example-kube-audit-rest-audit-events
Index pattern: example-kube-audit-rest-audit-events
Timestamp field: timestamp

Then click "Save data view to Kibana"

TODO: Make parsing prettier, so that it shows up properly in Elastic
