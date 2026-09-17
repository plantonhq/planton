# Kubernetes Planton Runner

Deploys a standing Planton runner appliance on a Kubernetes cluster -- an always-on, outbound-only worker that receives deploy operations from the control plane and executes them from inside the cluster's network. The module installs the official `planton-runner` chart (OCI, ghcr.io/plantonhq/charts) as a real Helm release, so the deployed runner is byte-identical to a hand-installed one. Enrollment is token-first: the runner joins with a token, registers itself, and receives its own individually revocable identity.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true` (with the standard Planton governance labels) and deleted with the resource; otherwise the runner installs into an existing namespace
- **Runner token Secret** -- the `<name>-token` Opaque Secret holding the runner token, created before the release; the chart reads it by name (its existingSecret form), so the token never rides rendered chart values
- **Helm Release** -- the `planton-runner` chart, pinned to `chartVersion` (default 0.5.0), rendering:
  - The runner Deployment at exactly 1 replica with a Recreate strategy -- two live pods under one runner name would revoke each other's keys
  - The ephemeral identity volume the runner persists its minted identity into -- container restarts reuse it; pod recreation re-joins with the token
  - The runner's health endpoints
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied to the module-created namespace and Secret

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- the credential used to install the release into the target cluster. Required.
- **Runner token** -- nothing to create by hand: the platform mints a runner token and writes it at exactly the managed-secret reference the manifest declares, before the infrastructure applies. Choose a secret slug and reference it as `$secret/<slug>` in the `token` field; never inline plaintext. (Self-service alternative: `planton runner token create`, or the console under Organization Settings → Runner Tokens.)

### Kubernetes Cluster

- **The namespace** -- must already exist when `createNamespace` is `false`; set it `true` to have the module own the namespace's lifecycle.
- **Tekton Pipelines (only for the build worker)** -- required when `build.enabled` is `true`; the runner then executes container-image build pipelines through Tekton on this cluster.

## Deploy

### Console

Open the deployment store, find **Kubernetes Planton Runner**, and click **Deploy**. The creation wizard walks you through environment and connection configuration and the spec fields. Start from the **Cluster Runner** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonRunner
metadata:
  name: cluster-runner
  org: acme-corp
  env: prod
spec:
  namespace:
    value: planton-runner
  createNamespace: true
  token: $secret/cluster-runner-token
```

```shell
planton apply -f runner.yaml
```

This minimal manifest installs chart version 0.5.0 tracking the latest runner release from the official image repository, at the chart's own default sizing (requests 100m/256Mi, limits 1/1Gi). The runner registers itself as `prod-cluster-runner` (`<env>-<metadata.name>`) the moment it joins, and `planton runner list` shows it. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the runner to a namespace managed by another Cloud Resource:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: platform-runners
      fieldPath: spec.name
  createNamespace: false
  token: $secret/cluster-runner-token
```

The InfraPipeline deploys the namespace first, then installs the runner into it.

## Key Configuration

These are the most important decisions when configuring the runner. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Token** -- `token` authorizes the runner to JOIN and is never its identity: the runner registers itself on first boot and receives its own individually revocable identity, and revoking the token never touches runners it already admitted. The module stores it in the `<name>-token` Secret; it never appears in rendered chart values.

**Runner name** -- leave `runnerName` unset: it defaults to `<env>-<metadata.name>` (`metadata.name` outside an environment), the same derivation the platform uses for records that reference this runner. Re-deploying with the same name and the same token re-admits the runner (lost-disk recovery); set it explicitly only when deliberately adopting an existing enrollment.

**Control-plane endpoint** -- `controlPlaneEndpoint` (host:port) is only for self-hosted control planes; leave it unset for Planton's hosted endpoint. It is the one bootstrap coordinate the join cannot deliver -- everything else arrives in the join response, so the runner self-configures its execution mode on arrival and no mode knob exists.

**Chart version** -- empty `chartVersion` installs the pinned default (0.5.0), the version this catalog release was validated against; since 0.5.0 the pod is probed over gRPC on the runner's port, the contract every execution mode serves, so a runner enrolled with a tunnel-less (self-hosted) instance stands Ready. Versions below 0.4.0 predate token enrollment and are refused loudly -- older charts would silently ignore the enrollment values and the runner would deploy with no way to join.

**Sizing** -- when `resources` is omitted, the chart's own defaults apply (requests 100m/256Mi, limits 1/1Gi), comfortable for the runner's control loops plus typical IaC operations. Memory pressure shows up as failed IaC operations mid-apply; size memory up before CPU.

**Build worker** -- `build.enabled: true` registers the runner as a build worker executing container-image build pipelines through Tekton (Tekton Pipelines must be installed); `build.tektonNamespace` defaults to the runner's own namespace.

**Escape hatch** -- `helmValues` merges raw values YAML over what the spec renders (Helm `-f` semantics), for chart knobs the spec does not model (nodeSelector, tolerations, extra env). The enrollment block is re-pinned after the merge and cannot be overridden -- the token never rides rendered values.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | The namespace the runner is installed in | Operational verification |
| `release_name` | The Helm release name (equals metadata.name) | Helm management and debugging |
| `token_secret_name` | The Kubernetes Secret holding the runner token | Auditing secret access; rotation tooling |
| `runner_name` | The name the runner registers itself under with the control plane | Console and `planton runner list` lookups |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard in-cluster runner** -- two decisions (a namespace and a token reference) make a cluster's private targets deployable: the runner dials out to the control plane, receives deploy operations, and executes them from inside the network with zero inbound exposure. Start from the **Cluster Runner** preset.

**Deploys plus image builds** -- the same runner additionally registered as a build worker: container-image build pipelines execute through Tekton on this cluster, without shipping build context out of the network. Tekton Pipelines is a prerequisite the module does not install. Start from the **Build Runner (Deploys + Image Builds)** preset.

**Enrolling against your own control plane** -- one `controlPlaneEndpoint` (host:port, no scheme) plus a token minted by YOUR instance is the entire difference from the standard shape -- a token minted by one control plane is meaningless to another. Start from the **Self-Hosted Control Plane** preset.

## Works With

- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) -- own the runner's namespace as a first-class resource and wire it through ValueFromRef
