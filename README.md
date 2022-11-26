# kube-rest-audit
Want to get a kubernetes audit log without having the ability to configure the kube-api-server such as with EKS?
Use kube-rest-audit

## Deploying

## Building

### Testing

## Known limitations
From the k8s documentation

```text
Rules describes what operations on what resources/subresources the webhook cares about. The webhook cares about an operation if it matches _any_ Rule. However, in order to prevent ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks from putting the cluster in a state which cannot be recovered from without completely disabling the plugin, ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks are never called on admission requests for ValidatingWebhookConfiguration and MutatingWebhookConfiguration objects.
```