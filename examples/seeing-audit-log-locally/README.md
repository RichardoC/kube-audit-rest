# Seeing the audit log locally, if using rancher desktop

The following one liner will output the generated audit log to standard out, which can then be piped to a file and used for testing your parser etc.

```bash
rdctl shell sudo cat "/var/lib/kubelet/pods/$(kubectl -n kube-audit-rest get po -l app=kube-audit-rest -ojsonpath='{.items[0].metadata.uid}' )/volumes/kubernetes.io~empty-dir/tmp/kube-audit-rest.log"
```

## Warnings

This is relying on k3s putting this at specific paths, which may not be true in future versions. This was tested with Rancher Desktop 1.7.0 and Kubernetes version 1.26
