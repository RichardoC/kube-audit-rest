# kube-audit-rest
Want to get a kubernetes audit log without having the ability to configure the kube-api-server such as with EKS?
Use kube-audit-rest

## What this is
A simple logger of requests to the k8s api.

## What this isn't
A filtering/redaction/forwarder system. That's for the user to do.

## Deploying

## Building
Requires nerdctl and rancher desktop as a way of building/testing locally with k8s.

```bash
./testing/setup.sh

# To cleanup
./testing/cleanup.sh
```

### Testing
Run via the Building commands, then the following should contain various admission requests

```bash
kubectl -n kube-audit-rest logs -l app=kube-audit-rest 
```

Confirm that the k8s API is happy with this webhook (log location may vary, check Rancher Desktop docs)
If it's working there should be no mention of this webook.

```bash
vim $HOME/.local/share/rancher-desktop/logs/k3s.log
```

Example failures

```
W1127 13:26:10.911971    3402 dispatcher.go:142] Failed calling webhook, failing open kube-audit-rest.kube-audit-rest.svc.cluster.local: failed calling webhook "kube-audit-rest.kube-audit-rest.svc.cluster.local": failed to call webhook: Post "https://kube-audit-rest.kube-audit-rest.svc:443/log-request?timeout=1s": x509: certificate signed by unknown authority (possibly because of "crypto/rsa: verification error" while trying to verify candidate authority certificate "ca.local")
W1127 13:35:04.936121    3402 dispatcher.go:142] Failed calling webhook, failing open kube-audit-rest.kube-audit-rest.svc.cluster.local: failed calling webhook "kube-audit-rest.kube-audit-rest.svc.cluster.local": failed to call webhook: Post "https://kube-audit-rest.kube-audit-rest.svc:443/log-request?timeout=1s": no endpoints available for service "kube-audit-rest"
E1127 13:35:04.936459    3402 dispatcher.go:149] failed calling webhook "kube-audit-rest.kube-audit-rest.svc.cluster.local": failed to call webhook: Post "https://kube-audit-rest.kube-audit-rest.svc:443/log-request?timeout=1s": no endpoints available for service "kube-audit-rest"

```

### Local testing

```bash
testing/locally/local-testing.sh
...
Test passed
{"level":"info","msg":"Server is shutting down...","time":"2022-12-01T19:43:01Z"}
{"level":"info","msg":"Server stopped","time":"2022-12-01T19:43:01Z"}
Terminated
```

If this failed, you will see `output not as expected`

## Known limitations
From the k8s documentation

```text
Rules describes what operations on what resources/subresources the webhook cares about. The webhook cares about an operation if it matches _any_ Rule. However, in order to prevent ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks from putting the cluster in a state which cannot be recovered from without completely disabling the plugin, ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks are never called on admission requests for ValidatingWebhookConfiguration and MutatingWebhookConfiguration objects.
```

This webhook also cannot know that all other validating webhooks passed so may log requests that were failed by other validating webhooks afterwarss.

Due to the failure:ignore there may be missing requests that were not logged in the interests of availability.

WARNING: This will log all details of the request! This namespace should be very locked down to prevent priviledge escalation!

### Certificate expires/invalid
The application logs will be full of the following error, and you will *not* get any more audit logs until this is fixed.
```2022/11/27 15:36:42 http: TLS handshake error from 10.42.0.1:46380: EOF```



## Next steps
* upload images on git commit
* make a distroless version
* explain zero stability guarantees until above completed
* clarify logs are not guaranteed to be ordered because there aren't guarantees from k8s that the requests would arrive in order.
* explain how to limit resources it's logging via the webhook resource (just a link to the k8s docs)
* follow GH best practises for workflows/etc
* add prometheus metrics, particularly for mem/cpu/total requests dealt with/invalid certificate refusal from client as this probably needs an alert as the cert needs replaced...
* make it clear just how bad an idea stdout is, preferably with a PoC exploit of using that to take over a cluster via logs...
* make it clear log file only exists if requests are sent
* clarify log file format is the raw response with no newlines in the json, with one response per line.
* clarify that kubernetes may not loadbalance between replicas as expected.
* test properly rather than use sleeps to manage async things...
* have the testing main.go spin up/shut down the binaries rather than using bash and make it clearer that diff is required.

## Completed next steps
* Use flags for certs locations
* write to a file rather than STDOUT with rotation and/or a max size
* Use structured logging
* rename to kube-audit-rest from kube-rest-audit
* Add examples folder
* Add local testing
* use zap for logging rather than logrus
* despite the issues, make it possible to log to stdout/stderr, as useful for capturing less sensitive info directly without infra
