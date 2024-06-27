# Example - Running kube-audit-rest and ingesting events into elastic search

After following this example, you will have an elastic search cluster running, with all creation/mutation events (except the limitations listed in the readme of this repo, and some spammy ones) being ingested into that cluster using vector. You'll also have configuration that drops the data field of secrets so that they aren't logged and filter out a variety of low signal to noise objects. This can be seen in the `vector-config` configmap in `examples/full-elastic-stack/k8s/kube-audit-rest.yaml`

You'll be able to navigate around in kibana and get a feel for the schema used, and what is available form this tool.

## Prerequisites

* Internet access

* A kubernetes cluster
    * which you have admin level privileges to
    * that you don't mind having to recreate
    * that doesn't already have elastic search operator running

* openssl
* kubectl
* bash
* envsubst
* base64
* echo

A good example would be Rancher Desktop, or minikube.

## How to follow the guide

Run all commands in the ```bash ``` blocks, and run them from a terminal at the root of this repo.

Warning, this is designed to be run on a local cluster which can be destroyed afterwards.

## Setting up elastic search

Largely following <https://www.elastic.co/downloads/elastic-cloud-kubernetes>

Set up the custom resources eck requires, then run the operator and lastly start an elastic search and kibana.

```bash
kubectl create -f https://download.elastic.co/downloads/eck/2.9.0/crds.yaml

kubectl apply -f https://download.elastic.co/downloads/eck/2.9.0/operator.yaml

kubectl apply -f examples/full-elastic-stack/k8s/elastic-cluster.yaml

```

Check that an operator pod is running in elastic-system

```bash
kubectl -n elastic-system get po
```

Then check that the elastic cluster, and kibana is running in example-kube-audit-rest

```bash
kubectl -n example-kube-audit-rest get po
NAME                                         READY   STATUS    RESTARTS   AGE
elasticsearch-kube-audit-rest-es-default-0         1/1     Running   0          23m
kibana-kube-audit-rest-kb-868975c597-4r9nj   1/1     Running   0          23m
```

## Accessing Kibana
Port forward kibana in a terminal, this will keep running until you terminate it with `ctrl+c`

```bash
kubectl -n example-kube-audit-rest port-forward svc/kibana-kube-audit-rest-kb-http   60443:https
```

To see that kibana is working navigate to <https://localhost:60443/app/discover#/> from your browser, and click to ignore the invalid certificate

Use another terminal to get the password to access

```bash
echo "username is elastic"
echo "password is $(kubectl -n example-kube-audit-rest get secret elasticsearch-kube-audit-rest-es-elastic-user -o=jsonpath='{.data.elastic}' | base64 --decode; echo)"
```

## Set up kube-audit-rest

Using a locked version from 2023-10-05

This configuration is deliberately non-HA, and will allow API calls to keep running if the wehook isn't running (failurePolicy:Ignore rather than Fail)

It will record all create/mutation/deletion API calls, which can leak service account tokens via secrets. This is to show to maximum capabilities.

In production limit this only to resources you want to capture.

## Create required certificates and upload them

Webhooks are required to serve TLS, so creating a the certificate authority and tls certificates

```bash
./examples/full-elastic-stack/certs/certs.sh
```

Upload the TLS certificate for use by the kube-audit-rest workload.

```bash
kubectl -n example-kube-audit-rest create secret tls kube-audit-rest --cert=./tmp/full-elastic-stack/server.crt --key=tmp/full-elastic-stack/server.key --dry-run=client -oyaml | kubectl -n example-kube-audit-rest apply -f -
```

## Deploy kube-audit-rest

```bash
kubectl -n example-kube-audit-rest apply -f examples/full-elastic-stack/k8s/kube-audit-rest.yaml
```

## Deploy the validation webhook

Warning, this is set to apply to every API call, and block the call if the webhook doesn't respond with success.

Webhooks are required to serve TLS, so the templating is including the certificate authority so kubernetes trusts our certificate

```bash
export CABUNDLEB64="$(cat tmp/full-elastic-stack/rootCA.pem | base64 | tr -d '\n')"
cat examples/full-elastic-stack/k8s/ValidatingWebhookConfiguration.yaml | envsubst | kubectl apply -f -
unset CABUNDLEB64
```

If you have any issues, delete the webhook with the following command, and change the failurePolicy to Ignore rather than Fail

## Do some api calls so you have something to look at
```bash
kubectl create ns test-namespace
kubectl -n test-namespace create serviceaccount abc
kubectl -n test-namespace create secret generic example-secret  --from-literal=VerySecret=topsecret
kubectl -n test-namespace delete secret example-secret
kubectl delete namespace test-namespace
```

## View the data in elastic search via kibana
Navigate to <https://127.0.0.1:60443/app/discover#/> provided the port forward from earlier is still running, or restart it if required.

Create a data view
```
Name: example-kube-audit-rest-audit-events
Index pattern: example-kube-audit-rest-audit-events
Timestamp field: timestamp
```

Then click "Save data view to Kibana"

You'll be able to see that the API calls regarding the secret were redacted

## Install in another namespace
If you would like to deploy to another namespace, recursively replace all occurences of `example-kube-audit-rest`
with your desired namespace name. Then just follow the guide from the local README.md - the commands will have
been adjusted. If you don't feel secure about just replacing everything, go through the matches
manually. Use at your own risk!

## Attach to an already deployed Elasticsearch-instance.
After testing it in a separate namespace, you can also use this example to connect kube-audit-rest with an existing
Elasticsearch instance. Adjust the namespace as indicated above, then recursively replace `elasticsearch-kube-audit-rest` 
the name of your Elasticsearch instance. If you don't feel secure about just replacing everything, go through the matches
manually.

Then follow this guide, but skip deploying the Elastic CRD's, Operator and 
`examples/full-elastic-stack/k8s/elastic-cluster.yaml` as you already have them. On your own risk :)

## Tidyup

WARNING this *will* delete the elastic operator, if it's already running in this cluster

```bash

export CABUNDLEB64="$(cat tmp/full-elastic-stack/rootCA.pem | base64 | tr -d '\n')"
cat examples/full-elastic-stack/k8s/ValidatingWebhookConfiguration.yaml | envsubst | kubectl delete -f -
unset CABUNDLEB64

kubectl delete namespace example-kube-audit-rest

kubectl delete -f https://download.elastic.co/downloads/eck/2.9.0/crds.yaml

kubectl delete -f https://download.elastic.co/downloads/eck/2.9.0/operator.yaml

kubectl delete -f examples/full-elastic-stack/k8s/elastic-cluster.yaml

```
