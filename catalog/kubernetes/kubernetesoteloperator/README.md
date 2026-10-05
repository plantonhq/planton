# Kubernetes OTel Operator

## When NOT to Use This

**This component installs the ENGINE, not a collector.** The
OpenTelemetry Operator reconciles `OpenTelemetryCollector` custom
resources into running collector fleets; those collectors are declared
with KubernetesOtelCollector — one resource per collector. Install the
operator once per Kubernetes cluster (it watches every namespace), then
declare collectors against it.

Also not the right kind when:

- **You want a collector** — that is KubernetesOtelCollector; this
  component is the controller that reconciles it.
- **You cannot run cert-manager** — KubernetesCertManager is a hard
  prerequisite, not a suggestion. The operator's admission webhooks
  need a serving certificate, and the retained CRDs' conversion trust
  must be kept current by cert-manager's CA injector (see below for
  why no self-signed arm exists).
- **You need the ClusterObservability alpha feature** — the
  `operator.clusterobservability` gate is off by default; enabling it
  through `helm_values` also brings its `clusterobservabilities` CRD,
  because the CRDs are derived from the chart with the same values the
  release installs with.

## Overview

**KubernetesOtelOperator** installs the OpenTelemetry Operator — the
controller that turns `OpenTelemetryCollector` declarations into
running collector fleets — from the official `opentelemetry-operator`
Helm chart
(https://open-telemetry.github.io/opentelemetry-helm-charts). Chart
0.120.0 (the pinned default) pairs with operator v0.156.0. The
operator injects a default collector image into CRs that declare none
and derives collector Service ports from the receivers each CR
declares.

**Key design points:**

- **The module owns the CRD lifecycle.** The chart templates its
  opentelemetry.io CRDs as release-owned resources — a Helm-owned
  install would cascade-delete every collector declaration in the
  cluster on uninstall. The modules therefore DERIVE the CRD set from
  the pinned chart at apply time (rendering it with the release's own
  values and the chart's CRD switch on), apply each CRD outside the
  release as a kept resource, and install the release with CRDs
  skipped and `crds.create: false` pinned. The schema always matches
  `chart_version`: a bump re-applies the CRDs at the new pin; destroy
  keeps them (unless `crds.keep_on_uninstall` is false), so removing
  the operator never deletes the fleet's declarations; a reinstall
  re-adopts them; lowering `chart_version` below what the cluster's
  CRDs carry is refused before anything is touched, with the remedy.
  Every CRD carries `planton.ai/crd-source-chart` and
  `planton.ai/crd-source-version` annotations, so `kubectl` shows where
  it came from.
- **Failures explain themselves.** A chart version that is not
  published, a repository unreachable at plan time, a render that
  produces no CRDs, a schema downgrade: each stops with what was
  observed, what it means, and the exact next step.
- **cert-manager is REQUIRED — a consequence of the CRD lifecycle.**
  cert-manager issues and rotates the webhook serving certificate, and
  the collector CRD carries a version-CONVERSION webhook whose trust
  (the `cert-manager.io/inject-ca-from` annotation) must be kept
  current by a RUNNING reconciler — cert-manager's CA injector —
  because the CRDs are retained past the release's lifetime. A
  certificate embedded once at install time would go stale on rotation
  and silently break collector-CR version conversion long after the
  install succeeded — which is why no self-signed/one-shot certificate
  arm exists in this spec.
- **Keep `metadata.name` at 30 characters or fewer.** The chart
  derives a `<name>-controller-manager-service-cert` Secret (33-char
  suffix) and Kubernetes caps names at 63 characters; the modules pin
  the chart's fullname to the resource name and fail loudly over the
  budget.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines). Two keys are re-pinned by the
  modules AFTER this merge, the deliberate exceptions to the escape
  hatch's last-word contract: `crds.create: false` (handing the CRDs
  to Helm would arm the uninstall cascade-delete this design exists to
  prevent) and `admissionWebhooks.certManager.enabled: true`
  (disabling it would leave module-owned CRDs pointing at a
  Certificate that no longer exists and silently break collector-CR
  conversion).

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install the operator into —
  literal or a KubernetesNamespace reference (`create_namespace` to
  own it)

### Common

- **`spec.chart_version`**: chart pin (default `0.120.0` — pairs with
  operator v0.156.0; bumping it re-applies the module-owned CRDs at the
  new pin; lowering it below the cluster's CRDs is refused)
- **`spec.crds.install`**: default true; false is the bring-your-own-CRDs
  arm — set ONLY when the CRDs are owned elsewhere (a GitOps-managed
  bundle); with the CRDs absent the operator cannot start
- **`spec.crds.keep_on_uninstall`**: default true; false lets a destroy
  delete the CRDs and, with them, every collector declaration
- **`spec.webhook.issuer_ref`**: cert-manager Issuer/ClusterIssuer to
  sign the webhook certificate; empty = the chart creates its own
  self-signed Issuer (the right choice for almost everyone)
- **`spec.default_collector_image`**: fleet-wide default collector
  image the operator injects into CRs that declare none; empty = the
  operator's compiled-in default
  (`ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-k8s`
  at the operator's paired version)
- **`spec.replicas`**: operator replicas (default 1; 2 gives a warm
  standby behind leader election)
- **`spec.resources`**: operator container resources — empty = the
  chart defaults
- **`spec.service_monitor_enabled`**: ServiceMonitor for the
  operator's own metrics (requires the monitoring.coreos.com CRDs —
  deploy KubernetesKubePrometheusStack first)
- **`spec.image_registry`**: registry replacing the manager image's
  registry part (air-gap/private-mirror path); does NOT rewrite the
  default collector image — mirror that via `default_collector_image`
- **`spec.image_pull_secrets`**: names of existing image-pull Secrets
  in the namespace
- **`spec.scheduling`**: node selector, tolerations, priority class
  for the operator pods
- **`spec.helm_values`**: the escape hatch (see above for the two
  re-pinned keys)

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
| `namespace` | Namespace the operator runs in |
| `release_name` | Helm release name (= `metadata.name`; the chart's fullname is pinned to it) |
| `webhook_service` | The operator's webhook Service (`<name>-webhook`, port 443) — where the API server sends admission reviews for collector CRs |
| `webhook_cert_secret_name` | The Secret holding the webhook serving certificate (`<name>-controller-manager-service-cert`) |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **KubernetesCertManager is a prerequisite**: the manager pod mounts
  the cert-manager-issued webhook Secret, and the retained CRDs'
  conversion trust rides cert-manager's CA injector.
- **KubernetesOtelCollector resources depend on this component**: the
  operator must be running before their `OpenTelemetryCollector`
  resources reconcile. It watches all namespaces — one install serves
  the whole cluster.
- **The install is deliberately blocking**: the Helm release waits for
  the operator to become Available (atomic, 600s timeout), so an
  install without a working cert-manager — or with an unpullable
  image — fails THIS apply with a readiness timeout instead of
  surfacing later as collectors that mysteriously never reconcile.

## Examples

### Standard install

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOtelOperator
metadata:
  name: otel-operator
spec:
  namespace:
    value: otel-operator
  createNamespace: true
```

### Explicit cert-manager issuer

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOtelOperator
metadata:
  name: otel-operator
spec:
  namespace:
    value: otel-operator
  createNamespace: true
  webhook:
    issuerRef:
      kind: ClusterIssuer
      name: internal-ca
```

### Private-mirror images

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOtelOperator
metadata:
  name: otel-operator
spec:
  namespace:
    value: otel-operator
  createNamespace: true
  imageRegistry: mirror.example.com
  defaultCollectorImage: mirror.example.com/otel/opentelemetry-collector-k8s:0.156.0
  imagePullSecrets:
    - mirror-pull
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
