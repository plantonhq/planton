# Kubernetes Solr Operator

## When NOT to Use This

**This component installs the ENGINE, not a search cluster.** The
Apache Solr Operator reconciles `SolrCloud` custom resources into
running Solr clusters; those clusters are declared with KubernetesSolr
— one resource per cluster. Install the operator once per Kubernetes
cluster, then declare SolrCloud clusters against it.

Also not the right kind when:

- **You want a Solr cluster** — that is KubernetesSolr; this component
  is the controller that reconciles it.
- **You want a managed cloud search service** — use the host cloud
  provider's managed search kinds; this kind is for running
  SolrCloud ON the Kubernetes cluster itself.
- **A zookeeper-operator already runs on the cluster** — do not
  install a second one: its fixed-name cluster-scoped RBAC conflicts
  on a second install. Set `zookeeper_operator.install: false` with
  `use_existing: true` instead.

## Overview

**KubernetesSolrOperator** installs the Apache Solr Operator — the
Apache Solr project's own operator for running SolrCloud on Kubernetes
— from the official `solr-operator` Helm chart
(https://solr.apache.org/charts). The operator reconciles `SolrCloud`
custom resources into running Solr clusters with managed rolling
updates, scaling with replica movement, and backup repositories — plus
`SolrBackup` and `SolrPrometheusExporter` resources.

**Key design points:**

- **ZooKeeper comes bundled.** SolrCloud requires ZooKeeper, and the
  chart bundles the zookeeper-operator as a dependency (installed by
  default) — so a KubernetesSolr resource can simply declare a
  provided ensemble and the operator provisions it. Disable the
  bundled install only when a zookeeper-operator already runs in the
  cluster (pair `install: false` with `use_existing: true`) or every
  Solr cluster will connect to an EXTERNAL ensemble.
- **The module owns the CRD lifecycle.** The chart carries its CRDs on
  both of Helm's surfaces: the three solr.apache.org CRDs (SolrCloud,
  SolrBackup, SolrPrometheusExporter) in its `crds/` directory, which
  Helm would install once and never upgrade, and the ZookeeperCluster
  CRD templated by the bundled zookeeper-operator subchart, which Helm
  would delete with the release. The modules therefore DERIVE the CRD
  set from the pinned chart at deploy time (rendered with the release's
  own values and the subchart's CRD switch on), apply each CRD outside
  the release as a kept resource, and install the release with CRDs
  skipped and `zookeeper-operator.crd.create: false` pinned. The schema
  always matches `chart_version`: a bump re-applies the CRDs at the new
  pin; destroy keeps them (unless `crds.keep_on_uninstall` is false), so
  removing the operator never cascade-deletes SolrCloud resources; a
  reinstall re-adopts them; a `chart_version` below what the cluster's
  CRDs carry is refused before anything is touched, with the remedy.
  Every CRD carries `planton.ai/crd-source-chart` and
  `planton.ai/crd-source-version` annotations, so `kubectl` shows where
  it came from. When the bundled subchart is not installed, its CRD is
  not derived either: the set follows the chart's own behaviour.
- **Every CRD failure explains itself.** A chart version that is not
  published, a repository that cannot be reached, a render that
  produces no CRDs, a schema downgrade: each stops with what was
  observed, what it most likely means, and the exact next step.
- **Chart and operator versions pair.** Chart 0.9.1 ships operator
  v0.9.1 (the chart version has no `v` prefix, the operator image tag
  carries one), and the operator is pre-1.0: minor versions can change
  the CRD API, which is exactly what the derived CRDs follow on a bump.
- **mTLS client identity is modeled.** When KubernetesSolr clusters
  enforce client certificates on their TLS listeners
  (`client_auth: Need`), the operator needs its own certificate to
  reach them — `mtls.client_cert_secret` (required whenever the block
  is declared), an optional CA secret, hostname-verification and
  rotation-watch knobs.
- **Watch scope defaults to cluster-wide.** `watch_namespaces` fences
  the operator to an explicit set; the watched namespaces need only
  exist by the time SolrCloud resources appear in them.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines); a safety valve, never the
  primary interface.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install the operator into —
  literal or a KubernetesNamespace reference (`create_namespace` to
  own it)

### Common

- **`spec.chart_version`**: chart pin (default `0.9.1`); a bump moves
  the module-owned CRDs with it
- **`spec.crds`**: the CRD lifecycle — `install` (default true; false is
  the bring-your-own-CRDs arm) and `keep_on_uninstall` (default true;
  false lets a destroy take every SolrCloud with the CRDs)
- **`spec.zookeeper_operator`**: the bundled dependency — `install`
  (default true) and `use_existing` (tell the Solr operator a
  zookeeper-operator is present even though this release does not
  install one)
- **`spec.watch_namespaces`**: watch exactly these namespaces; empty =
  all (the chart default)
- **`spec.replicas`**: operator replicas (default 1; extras are
  leader-elected warm standbys)
- **`spec.mtls`**: the operator's client identity for mutual-TLS Solr
  clusters — `client_cert_secret` (tls.crt/tls.key), `ca_cert_secret`
  + `ca_cert_secret_key` (default `ca-cert.pem`),
  `insecure_skip_verify` (chart default true — pod-IP calls rarely
  match certificate SANs), `watch_for_updates` (restart on rotation,
  default true)
- **`spec.leader_election_enabled` / `spec.metrics_enabled`**: chart
  defaults true; disable leader election only for a single-replica dev
  install
- **`spec.resources`**: empty = the chart defaults (none set — the
  operator is lightweight)
- **`spec.node_selector` / `spec.tolerations`**: operator pod
  scheduling
- **`spec.image` / `spec.image_pull_secret`**: air-gap path (empty =
  `apache/solr-operator` at the chart's matching tag; the chart
  accepts exactly ONE pull secret)
- **`spec.helm_values`**: the escape hatch

## When it fails

Every refusal on the CRD path says three things, in this order and in stable words a person or an agent can act on: what was observed (with the value), what it most likely means (one root cause), and the exact next step. The set the module anticipates, and where each is refused:

- **The pinned `chart_version` is not published.** Refused at plan, before anything is created: the version is not in the repository index. Next step: pin a version the index lists.
- **The chart repository cannot be reached from where the plan runs.** Refused at plan: the host does not resolve or egress is blocked; the install would fail the same way. Next step: check DNS and egress with the `curl -I` line the message gives.
- **A CRD schema downgrade.** The cluster's CRDs carry a higher chart version than the manifest asks for. Refused before anything is touched: an older schema over a newer one can strip fields from existing custom resources. Next step: pin the cluster's version or higher, or delete the CRD deliberately first.
- **A CRD already exists and belongs to someone else.** One of the chart's CRDs is on the cluster without this module's stamp (a hand-run `helm install`, a `kubectl apply`, another tool, another Planton module deriving the same name). Refused before anything is written, naming the owner. Next step: `crds.install: false` to leave the definitions with their owner (the release still uses them), or the two printed `kubectl` commands to hand them to this module once you know they match the pinned version (for a Helm-owned CRD, after freeing it from that release).
- **The deploy's identity may not write CRDs.** A namespace-admin identity cannot patch cluster-scoped CRDs; the module applies the chart's CRDs itself, outside the Helm release. Pulumi refuses at preview, Terraform at the first apply, in the same words: the identity, the verb, and the rules to grant from `iac/permissions.yaml`. Next step: grant them, or `crds.install: false` and have a cluster administrator apply the CRDs (`helm template --include-crds` renders them).
- **The render produced no CRDs.** Upstream renamed the chart's CRD switch or stopped shipping CRDs at this version. Next step: read the chart's values at the version and update the module's override.
- **A stale Helm repository entry on your own machine.** Helm consults the local repository list even for a URL-addressed chart; a missing index cache fails every render and install. Next step: `helm repo update` or remove the entry. The runner never meets this.

Two things the messages say on purpose. A kept CRD the module re-adopts on reinstall shows as `create` in a Terraform plan, because the state has no record of it; the apply adopts it in place and the Pulumi log says so. And a chart with no CRDs is never refused for a CRD right it does not need: the ownership read and the permission probe run only over what the render produced.

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the operator is installed into |
| `release_name` | Helm release name of the operator install (= `metadata.name`) |
| `deployment_name` | The Solr operator Deployment (the chart's fullname helper) |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`);
  **`mtls.client_cert_secret`** / **`ca_cert_secret`** reference
  existing Secrets (KubernetesSecret by default) — issue the client
  certificate with a KubernetesCertificate and reference its Secret.
- **KubernetesSolr resources depend on this component**: the operator
  (and, for provided ensembles, the bundled zookeeper-operator) must
  be running before SolrCloud resources reconcile. With
  `watch_namespaces` set, clusters outside the list are silently
  ignored.
- **The install is deliberately blocking**: the Helm release waits for
  the operator to become Available (atomic, 600s timeout), so an
  unpullable image fails THIS apply instead of surfacing later as
  SolrCloud resources that never reconcile.

## Examples

### Standard install (bundled zookeeper-operator)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolrOperator
metadata:
  name: solr-operator
spec:
  namespace:
    value: solr-operator
  create_namespace: true
```

### Cluster already runs a zookeeper-operator

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolrOperator
metadata:
  name: solr-operator
spec:
  namespace:
    value: solr-operator
  create_namespace: true
  zookeeper_operator:
    install: false
    use_existing: true
```

### Fenced to team namespaces, with an mTLS client identity

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolrOperator
metadata:
  name: solr-operator
spec:
  namespace:
    value: solr-operator
  create_namespace: true
  watch_namespaces:
    - search
    - analytics
  mtls:
    client_cert_secret:
      value: solr-operator-client-tls
    ca_cert_secret:
      value: solr-ca
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
