# kube-rest-audit
Want to get a kubernetes audit log without having the ability to configure the kube-api-server such as with EKS?
Use kube-rest-audit

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
kubectl -n kube-rest-audit logs -l app=kube-rest-audit 
```

Confirm that the k8s API is happy with this webhook (log location may vary, check Rancher Desktop docs)
If it's working there should be no mention of this webook.

```bash
vim $HOME/.local/share/rancher-desktop/logs/k3s.log
```

Example failures

```
W1127 13:26:10.911971    3402 dispatcher.go:142] Failed calling webhook, failing open kube-rest-audit.kube-rest-audit.svc.cluster.local: failed calling webhook "kube-rest-audit.kube-rest-audit.svc.cluster.local": failed to call webhook: Post "https://kube-rest-audit.kube-rest-audit.svc:443/log-request?timeout=1s": x509: certificate signed by unknown authority (possibly because of "crypto/rsa: verification error" while trying to verify candidate authority certificate "ca.local")
W1127 13:35:04.936121    3402 dispatcher.go:142] Failed calling webhook, failing open kube-rest-audit.kube-rest-audit.svc.cluster.local: failed calling webhook "kube-rest-audit.kube-rest-audit.svc.cluster.local": failed to call webhook: Post "https://kube-rest-audit.kube-rest-audit.svc:443/log-request?timeout=1s": no endpoints available for service "kube-rest-audit"
E1127 13:35:04.936459    3402 dispatcher.go:149] failed calling webhook "kube-rest-audit.kube-rest-audit.svc.cluster.local": failed to call webhook: Post "https://kube-rest-audit.kube-rest-audit.svc:443/log-request?timeout=1s": no endpoints available for service "kube-rest-audit"

```

## Known limitations
From the k8s documentation

```text
Rules describes what operations on what resources/subresources the webhook cares about. The webhook cares about an operation if it matches _any_ Rule. However, in order to prevent ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks from putting the cluster in a state which cannot be recovered from without completely disabling the plugin, ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks are never called on admission requests for ValidatingWebhookConfiguration and MutatingWebhookConfiguration objects.
```

This webhook also cannot know that all other validating webhooks passed so may log requests that were failed by other validating webhooks afterwarss.

Due to the failure:ignore there may be missing requests that were not logged in the interests of availability.

WARNING: This will log all details of the request! This namespace should be very locked down to prevent priviledge escalation!



## Next steps
* Use flags for certs locations
* Use structured logging
* Add local automated testing
* upload images on git commit
* make a distroless version
* add option to write to a file rather than STDOUT with rotation or a max size
