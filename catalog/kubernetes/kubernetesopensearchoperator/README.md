# Kubernetes OpenSearch Operator

## When NOT to Use This

**This component installs the ENGINE, not a search cluster.** The
OpenSearch Kubernetes Operator reconciles `OpenSearchCluster` custom
resources into running clusters; those clusters are declared with
KubernetesOpenSearch — one resource per cluster. Install the operator
once per Kubernetes cluster (or once per watched namespace), then
declare search clusters against it.

Also not the right kind when:

- **You want an OpenSearch cluster** — that is KubernetesOpenSearch;
  this kind is the controller that reconciles it.
- **You want a managed cloud search service** — use the host cloud
  provider's managed search kinds; this kind is for running
  OpenSearch ON the Kubernetes cluster itself.
- **You need the 3.x operator line today** — the served charts past
  2.8.0 (2.8.3, 2.8.4, 3.0.x) default their manager image to a
  PRERELEASE tag, and the 3.x line migrates the CRDs from the
  `opensearch.opster.io` API group to `opensearch.org`. The pinned
  default (2.8.0) is the newest stable chart/image pairing; pin a 3.x
  chart only after upstream cuts a stable 3.x operator release.

## Overview

**KubernetesOpenSearchOperator** installs the OpenSearch Kubernetes
Operator — the opensearch-project's operator for running OpenSearch
(the Apache-2.0 search and analytics engine) on Kubernetes — from the
official `opensearch-operator` Helm chart
(https://opensearch-project.github.io/opensearch-k8s-operator/). The
operator reconciles `OpenSearchCluster` custom resources into running
search clusters with managed TLS, security bootstrap, safe rolling
upgrades, and OpenSearch Dashboards.

**Key design points:**

- **The module owns the CRD lifecycle.** The chart templates its
  OpenSearch CRDs as release-owned resources with NO keep-on-uninstall
  knob upstream — a Helm-owned install would cascade-delete every
  OpenSearchCluster (and its data) on uninstall. The modules therefore
  DERIVE the CRD set from the pinned chart at deploy time (rendered with
  the release's own values and `installCRDs: true`), apply each CRD
  outside the release as a kept resource, and install the release with
  CRDs skipped and `installCRDs: false` pinned. The schema always
  matches `chart_version`: a bump re-applies the CRDs at the new pin
  (and adds the ones a newer chart brings); destroy keeps them (unless
  `crds.keep_on_uninstall` is false), so removing the operator never
  cascade-deletes clusters; a reinstall re-adopts them; a
  `chart_version` below what the cluster's CRDs carry is refused before
  anything is touched, with the remedy. Every CRD carries
  `planton.ai/crd-source-chart` and `planton.ai/crd-source-version`
  annotations, so `kubectl` shows where it came from.
- **Every CRD failure explains itself.** A chart version that is not
  published, a repository that cannot be reached, a render that
  produces no CRDs, a schema downgrade: each stops with what was
  observed, what it most likely means, and the exact next step.
- **Chart/image pairing is deliberate.** The chart's default manager
  image tag is its appVersion; chart 2.8.0 pairs with the stable
  operator 2.8.0 image. A `chart_version` bump is a license and
  API-group re-check, not a routine upgrade.
- **Watch scope defaults to cluster-wide.** The operator watches ALL
  namespaces by default (cluster-wide RBAC). Set `watch_namespace` to
  fence it to one namespace, and pair with `use_role_bindings` to swap
  ClusterRoleBindings for namespace-scoped RoleBindings on shared
  clusters — the spec rejects `use_role_bindings` without a
  `watch_namespace` (a cluster-wide operator cannot run on
  namespace-scoped permissions).
- **The metrics endpoint ships shielded.** The kube-rbac-proxy sidecar
  (chart default: enabled) puts the operator's metrics endpoint behind
  Kubernetes RBAC.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines). Two keys are off limits:
  `installCRDs` is re-pinned by the modules AFTER this merge (the one
  deliberate exception to the escape hatch's last-word contract), and
  `nameOverride`/`fullnameOverride` break the exported
  `deployment_name` output, which derives from the chart's default
  naming.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install the operator into —
  literal or a KubernetesNamespace reference (`create_namespace` to
  own it)

### Common

- **`spec.chart_version`**: chart pin (default `2.8.0` — the newest
  served chart whose default manager image is a stable operator
  release); a bump moves the module-owned CRDs with it
- **`spec.crds`**: the CRD lifecycle — `install` (default true; false is
  the bring-your-own-CRDs arm) and `keep_on_uninstall` (default true;
  false lets a destroy take every OpenSearchCluster with the CRDs)
- **`spec.watch_namespace`**: restrict the operator to one namespace;
  empty = watch all namespaces (the chart default)
- **`spec.use_role_bindings`**: namespace-scoped RoleBindings instead
  of ClusterRoleBindings — only valid together with `watch_namespace`
- **`spec.log_level`**: `debug`, `info` (default), `warn`, or `error`
- **`spec.dns_base`**: the cluster DNS domain baked into generated
  certificates and discovery addresses (default `cluster.local`) — a
  mismatch produces TLS certificates whose SANs do not match the
  service DNS names nodes advertise
- **`spec.parallel_recovery_enabled`**: recover cluster pods in
  parallel after failures (chart default true; upstream marks it
  experimental)
- **`spec.kube_rbac_proxy_enabled`**: the RBAC shield on the metrics
  endpoint (chart default true)
- **`spec.resources`**: manager container resources — empty = the
  chart defaults (requests 100m/350Mi, limits 200m/500Mi)
- **`spec.node_selector` / `spec.tolerations`**: operator pod
  scheduling
- **`spec.image` / `spec.image_pull_secrets`**: air-gap and
  private-mirror path (empty = `opensearchproject/opensearch-operator`
  at the chart's appVersion)
- **`spec.helm_values`**: the escape hatch (see above for the two
  off-limits keys)

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
| `deployment_name` | The operator controller-manager Deployment — the chart's fullname helper plus `-controller-manager` |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **KubernetesOpenSearch resources depend on this component**: the
  operator must be running and watching their namespace before their
  `OpenSearchCluster` resources reconcile. With the default
  cluster-wide watch, one operator install serves every namespace;
  with `watch_namespace` set, clusters outside that namespace are
  silently ignored by this install.
- **The install is deliberately blocking**: the Helm release waits for
  the operator Deployment to become Available (atomic, 600s timeout),
  so an unpullable image fails THIS apply instead of surfacing later
  as clusters that mysteriously never reconcile.

## Examples

### Standard cluster-wide install

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearchOperator
metadata:
  name: opensearch-operator
spec:
  namespace:
    value: opensearch-operator
  create_namespace: true
```

### Namespace-fenced install on a shared cluster

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearchOperator
metadata:
  name: search-team-operator
spec:
  namespace:
    value: search
  create_namespace: true
  watch_namespace: search
  use_role_bindings: true
```

### Private-mirror image

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearchOperator
metadata:
  name: opensearch-operator
spec:
  namespace:
    value: opensearch-operator
  create_namespace: true
  image:
    repository: my-mirror.example.com/opensearchproject/opensearch-operator
  image_pull_secrets:
    - mirror-pull-secret
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
