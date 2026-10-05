# First Kubernetes Charts: Production Cluster Baseline + Observability Stack

**Date**: August 4, 2026
**Type**: Feature
**Provider**: Kubernetes
**Chart(s)**: `charts/kubernetes/production-cluster-baseline`, `charts/kubernetes/observability-stack`

## Summary

The rebuilt catalog's first two charts land — the two most broadly wanted
platform compositions for teams that already run a Kubernetes cluster. The
**Production Cluster Baseline** turns a bare cluster into a production
platform in one deploy: automatic TLS from Let's Encrypt, DNS records
published for every exposed service, an ingress or Gateway API entry point,
cloud secrets synced into the cluster, and the metrics API autoscaling
depends on — keyless cloud authentication throughout. The **Observability
Stack** is a complete self-hosted telemetry platform: Prometheus with
curated alerts, Loki logs, Tempo traces with service graphs, Grafana wired
to all three at deploy time, and OpenTelemetry collection from every node
and application.

These charts also establish the catalog's authoring shape: template files
grouped in subdirectories by architectural layer (now taught in the forge
rule), one-owner/many-joiners namespace composition, keyless-by-default
credential posture with relaxation as the explicit parameter, and
relationship edges wherever ordering matters but no output flows.

## Highlights

### Production Cluster Baseline

- One `dns_provider` choice (Route 53, Cloud DNS, Azure DNS, Cloudflare)
  drives both external-dns record publication and the issuers' DNS-01
  challenge solvers — the two halves of the DNS story can never diverge.
- Production AND staging Let's Encrypt issuers with identical solvers, so
  certificate rollouts are proven against staging before touching
  production rate limits.
- Cloud credentials are keyless by default: solver and store auth blocks
  stay empty and inherit each controller's workload-identity binding
  (IRSA / GKE Workload Identity / Azure Workload Identity), wired from
  per-controller identity references.
- Exposure is an explicit either/or: ingress-nginx as the cluster's default
  IngressClass, or the Gateway API arm (standard-channel CRDs, a
  GatewayClass, a shared HTTP Gateway) with the implementation deliberately
  bring-your-own.
- Secret sync spans four backends (AWS Secrets Manager, GCP Secret Manager,
  Azure Key Vault, Vault/OpenBao) behind one enum on a cluster-wide store.

### Observability Stack

- A first-class namespace resource owns `observability`; every kind
  joins it — shared-namespace ownership is structural, not conventional.
- Grafana ships wired to Prometheus, Loki, AND Tempo by reference; team
  dashboards arrive by labeled ConfigMap, never by editing the chart.
- Loki's upstream cache sizing (a ~9.8Gi memory request that silently never
  schedules on smaller nodes and rolls back the whole install) is defused by
  parameters that default to deployable sizes and teach the math.
- Tempo's metrics generator and Prometheus's remote-write receiver ride one
  toggle — the two halves of the service-graph contract flip together.
- Telemetry collection is complete: an OTLP gateway for applications and a
  per-node log DaemonSet with its ServiceAccount and cluster-read RBAC
  composed in-chart.

## Guard Improvements

Two chart CI guards were aligned with the loaders they front:

- The license-footer guard now discovers chart READMEs beside each
  `Chart.yaml` at any nesting depth (previously hardwired to
  `charts/<provider>/<name>/`), and a chart missing its README entirely now
  fails instead of passing silently.
- The structure guard's non-empty-templates check now walks `templates/`
  recursively and accepts `.yml`, matching the offline validator and the
  platform's Infra Stack loader — charts may organize templates in subdirectories.

## Validation

Both charts pass the offline gate with the CLI built from this tree:
defaults plus every bool-toggle flip, plus explicit sweeps across every DNS
provider, every secret-store backend, every workload-identity arm, and the
Gateway API combinations. `chart validate --all charts/` is green, both
guards pass against the real charts, and every icon URL answers 200.
