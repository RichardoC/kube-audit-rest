# kube-audit-rest

Want to get a kubernetes audit log without having the ability to configure the kube-api-server such as with EKS?
Use kube-audit-rest to capture all API calls to disk, before exporting those to your logging infrastructure.
This should be much cheaper than cloudtrail which charges ~ per API call and doesn't support ingestion filtering.  

## What this is

A simple logger of requests to the k8s api.

## What this isn't

A filtering/redaction/forwarder system. That's for the user to do.

## Kubernetes distribution compatibility

Unknown but likely to work with all distributions due to how fundamental the ValidatingWebhook API is to Kubernetes operators. At worst the MutatingAdmissionWebhook API can be used instead, though that does mean that subverting this binary could lead to a cluster takeover and that it may not log the final version of the object.


## Usage

An example of how to deploy this service can be found within `./k8s` and steps to actually deploy it in `testing/setup.sh`
You could either run this centrally (though it would be difficult to tell which API calls are from which clusters) or running in each cluster.
At minimum you require
* Ability to create ValidatingWebhookConfiguration on the target k8s cluster.
* A CA, and a TLS certificate signed for the address the kubernetes control plane is connecting to for connections to kube-audit-rest
* kube-audit-rest running somewhere connectable by the kubernetes control plane.
* some disk space for kube-audit-rest to write to. Defaults to `/tmp` which is a ramfs on most linux distributions, though not on Kubernetes.
* Either to build a copy of the binary yourself, or download a copy of the docker image via the steps on the [packages page](https://github.com/RichardoC/kube-audit-rest/pkgs/container/kube-audit-rest) which is available as a distroless image (default, and `latest`) with suffix -distroless and a -alpine image based on the alpine docker image.

If you are running kube-audit-rest within the kubernetes cluster it is auditing you also require
* a deployment of kube-audit-rest running
* a service targeting the kube-audit-rest pods


### Binary options

```bash
$ kube-audit-rest --help
Usage:
  kube-audit-rest [OPTIONS]

Application Options:
      --logger-filename=    Location to log audit log to (default: /tmp/kube-audit-rest.log)
      --audit-to-std-log    Not recommended - log to stderr/stdout rather than a file
      --logger-max-size=    Maximum size for each log file in megabytes (default: 500)
      --logger-max-backups= Maximum number of rolled log files to store (default: 3)
      --cert-filename=      Location of certificate for TLS (default: /etc/tls/tls.crt)
      --cert-key-filename=  Location of certificate key for TLS (default: /etc/tls/tls.key)
      --server-port=        Port to run https server on (default: 9090)
  -v, --verbosity           Uses zap Development default verbose mode rather than production

Help Options:
  -h, --help                Show this help message
```

### Resource requirements

Unknown, if anyone performs benchmarks please open a pull request with your findings. These can be set by following the instructions [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

### Limiting which requests are logged

In your `ValidatingWebhookConfiguration` use the limited amount of resources and verbs you wish to log, rather than the `*`s in `./k8s/webhook.yaml` using the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#webhook-configuration)

## API spec for kube-audit-rest output

This is the raw [AdmissionRequest](https://github.com/kubernetes/api/blob/master/admission/v1/types.go#L39) request and can be parsed using that [schema](https://github.com/kubernetes/kubernetes/blob/master/api/openapi-spec/swagger.json)

An easier version of this to interact with can be found [here](https://github.com/yannh/kubernetes-json-schema/)

kube-audit-rest should log one request per line, in compacted json.

## Metrics
kube-audit-rest provide some metrics describing its own operations, both as an application specifically and as a go binary. .

All the specific application metrics are prefixed with `kube_audit_rest_`.

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| kube_audit_rest_valid_requests_processed_total | Counter | | Total number of valid requests processed |
| kube_audit_rest_http_requests_total | Counter | | Total number of requests to kube-audit-rest |

kube-audit-rest also exposes all default go metrics from the (Prometheus Go collector)[https://github.com/prometheus/client_golang/blob/main/prometheus/go_collector.go]

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

## Known limitations and warnings

From the k8s documentation [see rules](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#validatingwebhook-v1-admissionregistration-k8s-io)

```text
Rules describes what operations on what resources/subresources the webhook cares about. The webhook cares about an operation if it matches _any_ Rule. However, in order to prevent ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks from putting the cluster in a state which cannot be recovered from without completely disabling the plugin, ValidatingAdmissionWebhooks and MutatingAdmissionWebhooks are never called on admission requests for ValidatingWebhookConfiguration and MutatingWebhookConfiguration objects.
```

This webhook also cannot know that all other validating webhooks passed so may log requests that were failed by other validating webhooks afterwards.

Due to the `failure: ignore` in the example webhook configurations there may be missing requests that were not logged in the interests of availability of the kubernetes API..

WARNING: This will log all details of the request! This namespace should be very locked down to prevent privilege escalation!

This webhook will also record dry-run requests.

The audit log files will only exist if valid API calls are sent to the webhook binary.

API calls can be logged repeatedly due to Kubernetes repeatedly re-calling the [webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#reinvocation-policy) and thus may not be in chronological order.

### Certificate expires/invalid

The application logs will be full of the following error, and you will *not* get any more audit logs until this is fixed.
```2022/11/27 15:36:42 http: TLS handshake error from 10.42.0.1:46380: EOF```

Kubernetes may not load balance between replicas of kube-audit-rest in the way you expect as this behaviour appears to be undocumented.


## Next steps

* explain zero stability guarantees until above completed
* follow GH best practises for workflows/etc
* add prometheus metrics, particularly for mem/cpu/total requests dealt with/invalid certificate refusal from client as this probably needs an alert as the cert needs replaced...
* make it clear just how bad an idea stdout is, preferably with a PoC exploit of using that to take over a cluster via logs...
* test properly rather than use sleeps to manage async things...
* have the testing main.go spin up/shut down the binaries rather than using bash and make it clearer that diff is required.
* have workflow to test that docker image can be created once a maintainer adds a label to the PR.
* document this writes to default ephemeral storage

## Completed next steps

* Use flags for certs locations
* write to a file rather than STDOUT with rotation and/or a max size
* Use structured logging
* rename to kube-audit-rest from kube-rest-audit
* Add examples folder
* Add local testing
* use zap for logging rather than logrus for prettier http error logs
* despite the issues, make it possible to log to stdout/stderr, as useful for capturing less sensitive info directly without infra
* upload images on git commit
* make a distroless version
* clarify logs are not guaranteed to be ordered because there aren't guarantees from k8s that the requests would arrive in order.
* explain how to limit resources it's logging via the webhook resource (just a link to the k8s docs)
* make it clear log file only exists if requests are sent
* clarify log file format is the raw response with no newlines in the json, with one response per line.
* clarify that kubernetes may not loadbalance between replicas as expected.
* document that image defaults to distroless
