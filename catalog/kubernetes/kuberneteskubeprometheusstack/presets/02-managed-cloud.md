# Managed cloud preset

The production-shaped install for EKS, GKE, AKS and peers: a 50Gi
Prometheus volume with 15-day retention, a durable Alertmanager
volume, persistence on the bundled Grafana, and every control-plane
scraper that a managed provider hides turned OFF — with the matching
default-rule groups disabled so the alert set stays honest.

The API server and kubelets stay scraped (they are reachable on every
managed cloud). kube-proxy is off by default in this preset because
its metrics port commonly binds to localhost on managed platforms and
because Cilium's kube-proxy-replacement leaves nothing to scrape;
re-enable it only after you have verified the target is up on your
distribution.

The two overcommit alerts are off for the same reason: managed node
pools usually autoscale, and those alerts count today's nodes, so they
fire on every burst the autoscaler is about to absorb. On fixed-size
pools they are true; remove `disabled_alerts` there.

Set `prometheus.external_labels.cluster` to the cluster's real name
so multi-cluster federation and remote-write destinations can tell
series apart. Compose exposure for Grafana or Prometheus over the
exported service handles — every UI stays ClusterIP.

See [02-managed-cloud.yaml](./02-managed-cloud.yaml) for the manifest.
