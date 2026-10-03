# OPA Gatekeeper

Deploys OPA Gatekeeper -- the Open Policy Agent's Kubernetes admission controller -- from the official `gatekeeper` chart. Policies are ConstraintTemplates (Rego or Kubernetes CEL) instantiated as typed Constraint resources, enforced at admission and continuously audited against what already runs: the framework of choice where policy teams standardize on OPA across more than just Kubernetes. This component installs the ENGINE only -- the webhook controller manager, the audit controller, and the constraint-framework CRDs. The policies themselves are declared separately, and one Gatekeeper serves the whole cluster. Uses a Kubernetes Provider Connection for cluster access.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; the module then ALSO declares the `admission.gatekeeper.sh/ignore` exemption label on the namespace object itself, so day-2 applies never strip what the chart's hook stamped
- **Helm Release** -- the `gatekeeper` chart, creating:
  - Deployment for the webhook controller manager (3 replicas by the chart default -- the webhook sits on the cluster's write path), with anti-affinity across hosts and `system-cluster-critical` priority out of the box
  - Deployment for the audit controller -- always a single replica; re-evaluates existing resources against constraints and records violations in each constraint's status
  - The ValidatingWebhookConfiguration and MutatingWebhookConfiguration -- templated WITH the release (unlike engines that register webhooks at runtime), so they install and delete with it
  - The engine CRDs (constrainttemplates, configs, expansiontemplates, mutators, ...) from the chart's `crds/` directory -- installed on FIRST install, never upgraded or deleted by Helm; the chart's pre-install/pre-upgrade hook Job keeps their schema in step with the chart version
  - Lifecycle hook Jobs: the namespace-labeling hook (stamps the exemption label on Gatekeeper's own namespace so the engine never polices itself), the webhook probe, and the CRD upgrade Job
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster-admin-grade permissions** -- the install creates cluster-scoped CRDs and webhook configurations.
- **A pre-issued TLS Secret, only if you choose external webhook certificates** -- the default posture needs nothing: Gatekeeper's embedded cert controller generates and rotates the webhook certificate. Declaring `externalCert` requires the Secret (typically materialized by cert-manager) to exist in the install namespace BEFORE the install -- the chart mounts it, and a missing Secret holds the rollout.
- **Air-gapped clusters need THREE images mirrored** -- the engine (`openpolicyagent/gatekeeper`, overridable via the typed `image` field), the CRD hook (`openpolicyagent/gatekeeper-crds`), and the webhook probe (`curlimages/curl`); the last two ride `helmValues`.

## Deploy

### Console

Open the deployment store, find **OPA Gatekeeper**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Audit-first preset** for Gatekeeper as it ships (fail-open, audit-only until constraints say otherwise), the **Production enforce preset** for the fail-closed posture with a tuned audit loop, or the **cert-manager TLS preset** to serve the webhook with an externally issued certificate, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGatekeeper
metadata:
  name: gatekeeper
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "gatekeeper-system"
  createNamespace: true
  replicas: 3
  validatingWebhook:
    failurePolicy: Fail
    timeoutSeconds: 5
  audit:
    intervalSeconds: 120
    matchKindOnly: true
    constraintViolationsLimit: 50
  exemptNamespacePrefixes:
    - kube-
  engine:
    logDenies: true
```

```shell
planton apply -f gatekeeper.yaml
```

This deploys the fail-closed enforcement posture: three webhook replicas (the count the spec's own guidance names for choosing Fail), a 5-second timeout so a sick engine degrades admissions instead of hanging them, the audit loop tuned for a real cluster, and every blocked admission written to the controller log. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire Gatekeeper to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: gatekeeper-namespace
      fieldPath: spec.name
  createNamespace: false
  externalCert:
    secretName:
      valueFrom:
        kind: KubernetesCertificate
        name: gatekeeper-webhook-cert
        fieldPath: spec.secret_name
```

The InfraPipeline deploys the namespace and the certificate first, then provisions Gatekeeper against them.

## Key Configuration

These are the most important decisions when configuring Gatekeeper. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The deliberate fail-OPEN / fail-CLOSED pair** -- `validatingWebhook.failurePolicy` defaults `Ignore` (fail-open: an engine outage never blocks admission -- but a request that slips through during one is not evaluated), while `validatingWebhook.checkIgnoreFailurePolicy` -- the webhook guarding Gatekeeper's own exemption label -- defaults `Fail` (the label cannot be smuggled onto a namespace during that same outage; its blast radius is namespace label edits only). Flipping the policy webhook to `Fail` closes the smuggling window and blocks every matched admission when the engine is down: only responsible with the webhook highly available (3 replicas) and a short timeout.

**Exemption is a two-key system** -- `exemptNamespaces` and `exemptNamespacePrefixes` AUTHORIZE namespaces to carry the `admission.gatekeeper.sh/ignore` label -- they do not exempt by themselves. Exemption takes both keys: authorization on the list AND the label on the namespace, applied separately, with the fail-closed label guard enforcing the authorization. The chart's post-install hook labels only Gatekeeper's own namespace (`hooks.labelNamespace`, default true -- leave it on, or the engine polices its own pods and can deadlock itself).

**The CRD posture** -- The engine CRDs ship in the chart's `crds/` directory: installed on first install, never upgraded or deleted by Helm. Destroying the engine KEEPS them -- and every ConstraintTemplate, Constraint, and runtime-generated constraint CRD with them; a later install adopts the kept CRDs and the policy library enforces again. `hooks.upgradeCrds` (default true) is the chart's own answer to upgrades -- a pre-install/pre-upgrade Job applies the CRDs at the chart's version; disabling it leaves CRDs frozen at their first-install schema.

**Declared lists REPLACE chart defaults** -- `engine.disabledBuiltins` carries the chart default `["{http.send}"]` (Rego in ConstraintTemplates must not make arbitrary network calls from the admission path; external data providers are the sanctioned alternative). Declaring the field REPLACES the default: re-include `{http.send}` unless dropping it is a deliberate acceptance.

**The audit loop, tuned as typed fields** -- `audit.intervalSeconds` (default 60; 0 is a real position: run the audit exactly once at startup), `constraintViolationsLimit` (default 20; raising it grows constraint objects in etcd), `matchKindOnly` (only list kinds some constraint actually matches), `chunkSize`, and `fromCache` (requires syncing kinds via a Gatekeeper Config resource). The audit controller records violations -- it never blocks anything.

**Webhook TLS** -- By default Gatekeeper's embedded cert controller generates and rotates the webhook certificate in the chart-fixed `gatekeeper-webhook-server-cert` Secret: zero prerequisites. Declaring `externalCert` disables the embedded rotation and mounts your Secret instead -- and the chart only auto-disables rotation on the audit controller, so the module sets `disableCertRotation` on the controller manager explicitly; without that the embedded rotator would keep overwriting your issued certificate.

**The pre-delete webhook cleanup** -- `hooks.deleteWebhookConfigurationsOnUninstall` (default false -- chart-owned webhook configurations already delete with the release) adds a pre-delete hook Job for teardown-ordering edge cases. The raw chart value alone fails at uninstall (it reads a service-account name from a key the chart never renders), so the module renders the hook's service-account name alongside.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesCertificate** | `externalCert.secretName` | `spec.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the engine runs in | Locating the engine for diagnostics |
| `release_name` | Helm release name (equals metadata.name) | Helm management and debugging |
| `webhook_service_name` | The Service the webhook configurations point at | Webhook diagnostics and network policy |
| `webhook_cert_secret_name` | The Secret carrying the webhook server certificate | TLS diagnostics and rotation checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Audit first** -- Gatekeeper as it ships: three webhook replicas, the policy webhook fail-open, and the audit controller re-checking existing resources every 60 seconds. The right first posture for adopting policy on a running cluster. Start from the **Audit-first preset**.

**Production enforce** -- The fail-closed policy webhook with a 5-second timeout, the audit loop tuned for scale, kube-* prefixes authorized for exemption, and every deny logged. Start from the **Production enforce preset**.

**cert-manager TLS** -- The webhook serving a certificate issued by cert-manager instead of the embedded rotator, with the Secret reference composing the two kinds in the resource graph. Start from the **cert-manager TLS preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the Gatekeeper install
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) -- materializes the TLS Secret when webhook issuance is switched to cert-manager
- [**Kubernetes Manifest**](/infra-catalog/kubernetes-manifest) -- applies the ConstraintTemplates and Constraints the engine enforces
