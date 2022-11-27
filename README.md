# kube-rest-audit
Want to get a kubernetes audit log without having the ability to configure the kube-api-server such as with EKS?
Use kube-rest-audit

## Deploying

## Building
Requires nerdctl and rancher desktop as a way of building/testing locally with k8s.

```bash
./testing/setup.sh
```

### Testing

## Known limitations
From the k8s documentation

```text
Rules describes what operations on what resources/subresources the webhook cares about. The webhook cares about an operation if it matches _any_ Rule. However, in order to prevent ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks from putting the cluster in a state which cannot be recovered from without completely disabling the plugin, ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks are never called on admission requests for ValidatingWebhookConfiguration and MutatingWebhookConfiguration objects.
```

This webhook also cannot know that all other validating webhooks passed so may log requests that were failde by other validating webhooks.

## Next steps
* Add local automated testing
* upload images on git commit
