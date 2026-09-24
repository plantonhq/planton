# Tekton

Declares the cluster's Tekton installation — which components run (Pipelines, Triggers, Dashboard, Chains), where they install, the pipeline feature flags and execution defaults, and the cleanup policy — as a `TektonConfig` custom resource that the Tekton Operator reconciles into running components and keeps converged.

This component shapes the **engine**, never your pipelines: Tasks, Pipelines, and their runs are plain custom resources once this converges — declare them via Kubernetes Manifest resources, your platform, or the Tekton CLI. Exactly one KubernetesTekton per cluster is allowed (the operator's own admission rule), and the operator (declared with **Tekton Operator**) must be installed first.

## What Gets Created

When you deploy this Cloud Resource, the IaC module renders the TektonConfig and the operator turns it into running components:

- **The TektonConfig** — always named `config` (the operator's singleton rule; your resource name keys the Planton record only), carrying the profile, target namespace, pipeline configuration, and pruner policy
- **The target namespace** (`tekton-pipelines` unless overridden) — CREATED and OWNED by the operator, including deletion: teardown removes the namespace with the components, so it must never carry anything else
- **The profile's components** — `lite` installs Pipelines only; `basic` adds Triggers; `all` (the operator default) adds the Dashboard; Chains additionally installs on `basic`/`all` unless disabled
- **The pruner CronJob** — only when the pruner is declared; deletes completed PipelineRuns/TaskRuns past the retention rule
- **Component pods scheduled per your placement** — node selector, tolerations, and priority class applied to EVERY Tekton component pod the operator deploys (pipeline RUN pods schedule per their own runs)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **The Tekton Operator** — a HARD prerequisite: declare a **Tekton Operator** resource first; without it, nothing reconciles this declaration.
- **No existing TektonConfig** — the operator's admission webhook allows exactly one per cluster.
- **A dedicated target namespace name** — the operator creates it, owns it, and DELETES it on teardown. Never point `targetNamespace` at a namespace carrying anything else.

## Deploy

### Console

Open the deployment store, find **Tekton**, and click **Deploy**. The creation wizard walks you through the singleton contract, the profile ladder, the immutable target namespace, placement, the pipeline surface (execution defaults, feature flags, resolvers, metrics, performance), the per-component steps (Triggers, Dashboard, Chains — shown only on profiles that install them), the pruner, and the additional-params escape surface. Start from the **CI standard preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTekton
metadata:
  name: tekton
  org: acme-corp
  env: prod
spec:
  pruner:
    schedule: "0 8 * * *"
    resources:
      - pipelinerun
      - taskrun
    keep: 100
```

```shell
planton apply -f tekton.yaml
```

This is the CI-standard shape: profile `all` by absence (Pipelines + Triggers + Dashboard, plus Chains), the upstream default `tekton-pipelines` namespace, and the one piece of configuration no production cluster should skip — a pruner. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Tekton installation. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The profile ladder** — `lite` = Pipelines only (the embedded-engine shape for platforms that create PipelineRuns programmatically); `basic` = + Triggers (webhook-driven execution); `all` = + Dashboard (the operator's own default when unset). Chains rides along on `basic`/`all` unless `chain.disabled`.

**`targetNamespace` is IMMUTABLE and operator-owned** — the operator's webhook rejects changing it on an existing installation (destroy and recreate to move), and teardown DELETES the namespace with the components (Tekton Results archival finalizers can hold it Terminating for a while). Empty = `tekton-pipelines`.

**One CloudEvents sink per cluster** — `pipeline.cloudEventsSinkUrl` is Tekton's single, cluster-global event destination: every run in every namespace reports there, and per-namespace sinks do not exist. Planton's own runners need none -- a build-capable runner watches its build namespace's PipelineRuns and TaskRuns directly and reports to its own control plane -- so a cluster serving several Planton control planes leaves this unset. Set it only for an event consumer of your own, with a fan-out service at the URL if more than one consumer needs the stream (each event carries its source namespace). Must be an `http://` or `https://` URL.

**Feature flags are tri-states** — every pipeline feature flag left unset keeps Tekton's own default for the pinned release; a pinned value holds even if a future release changes its default. `keepPodOnCancel` is alpha-gated (takes effect only with `enableApiFields: alpha`); `resultsFrom: sidecar-logs` pairs with `maxResultSize`.

**Resolvers default ON — including the internet-reaching ones** — all four remote resolvers (bundles, hub, git, cluster) are enabled by Tekton's default; `git` and `hub` reach the public internet from the resolvers deployment, so locked-down clusters disable exactly those.

**Tekton HA is sharding** — `performance.replicas` and `performance.buckets` are one decision: extra replicas only take work when buckets shards it across them (upstream bucket maximum 10). The usual throughput ceiling is the controller's Kubernetes API budget — raise `kubeApiQps`/`kubeApiBurst` together with run volume.

**The dashboard has NO built-in authentication** — in its default writable mode, anyone who reaches it can run and delete pipelines. Set `dashboard.readonly`, and expose the ClusterIP Service only through first-class kinds (KubernetesIngress, Gateway API kinds) with an authenticating layer — or keep it unexposed and use the exported port-forward command.

**No pruner = unbounded growth** — completed runs keep their pods until something deletes them. Declare the pruner with a required cron `schedule`, at least one of `pipelinerun`/`taskrun`, and EXACTLY one retention rule — `keep` (newest N) or `keepSince` (younger than N minutes), never both (the operator's webhook enforces it).

**Destroy while the operator lives** — deleting this resource makes the operator tear down every component, and the deletion BLOCKS until that finishes. Never destroy the KubernetesTektonOperator resource first — the teardown finalizers strand without a running operator.

## Outputs and Dependencies

### What This Component Consumes

This component's spec is self-contained — no fields reference other resources' outputs. Its one dependency is environmental: a running Tekton Operator on the target cluster (see Before You Deploy).

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the Tekton components run in (`tekton-pipelines` unless overridden) | Composition, debugging |
| `profile` | The installed profile (`lite`, `basic`, or `all`) | Operational tooling |
| `dashboard_service` | Name of the dashboard Service (`tekton-dashboard`) in the target namespace; empty unless profile is `all` | Backend handle for KubernetesIngress / KubernetesHttpRoute exposure |
| `dashboard_kube_endpoint` | In-cluster endpoint of the dashboard; empty unless profile is `all` | In-cluster integrations |
| `port_forward_command` | Command to port-forward the dashboard to a workstation; empty unless profile is `all` | Workstation access without exposure |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CI Standard** — the full control plane (profile `all` by absence) with a daily pruner keeping the newest 100 runs of each kind. Start from the **CI standard preset**.

**Pipelines Engine** — Tekton as an embedded engine: profile `lite`, every run's lifecycle streaming to one CloudEvents receiver, the internet-reaching resolvers off, the controller sharded 2×2, and aggressive two-day retention. Start from the **Pipelines engine preset**.

## Works With

- [**Tekton Operator**](/cloud-catalog/kubernetes-tekton-operator) — the hard prerequisite; deploy the operator FIRST, destroy this declaration FIRST.
- [**Kubernetes Ingress**](/cloud-catalog/kubernetes-ingress) — exposes the dashboard over the exported `dashboard_service` handle, always behind an authenticating layer.
- [**Kubernetes HTTPRoute**](/cloud-catalog/kubernetes-http-route) — the Gateway API alternative for dashboard exposure, same authenticating-layer caveat.
- [**Kubernetes Manifest**](/cloud-catalog/kubernetes-manifest) — declares Tasks, Pipelines, EventListeners, and runs once the installation converges.
