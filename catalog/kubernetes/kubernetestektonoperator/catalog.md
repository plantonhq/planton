# Tekton Operator

Installs the Tekton Operator — the lifecycle manager maintained by the Tekton project — from its official single-file release manifest (the in-repo Helm chart is unpublished and is not a distribution channel). The operator reconciles a `TektonConfig` declaration (declared with **Tekton**) into running Tekton components — Pipelines, Triggers, Dashboard, Chains — managing their installation, upgrades, and removal through `TektonInstallerSet` resources.

This component installs the **manager only**. Installing it deploys NO pipeline runtime: automatic component installation is disabled by design, so the KubernetesTekton declaration is the single owner of what Tekton actually runs on the cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module applies the release manifest's documents:

- **The `tekton-operator` namespace** — the manifest's FIXED installation namespace, baked into its own cross-references (the webhook Service, the RBAC subjects); it is not configurable
- **14 `operator.tekton.dev` CRDs** (including `tektonconfigs.operator.tekton.dev`) — documents of the applied manifest, so they install AND delete with this resource; see the destroy ordering under Key Configuration
- **The operator Deployment** (`tekton-operator`, running `ghcr.io/tektoncd/operator/operator-*` at the pinned `v0.80.0` release; its two containers — lifecycle and cluster-operations — share one image) with its RBAC
- **The admission webhook Deployment** (`tekton-operator-webhook`) validating TektonConfig declarations — including the one-per-cluster singleton rule and the immutable target namespace
- **Auto-installation disabled** — upstream, the operator auto-creates a default TektonConfig (profile `all`) at startup; this install always disables that, because two managers writing one object fight through server-side apply and the cluster's Tekton shape would depend on install order

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **No existing install** — exactly ONE operator install per cluster is the upstream contract: its webhooks and CRDs are cluster-scoped singletons with fixed names, and a second install cannot coexist.
- **Registry reachability** — every Tekton image pulls from GitHub Container Registry (`ghcr.io`) unless `imageRegistry` names a mirror of it; air-gapped clusters set `imageRegistry` and, for a private mirror, pull secrets (see Key Configuration).

## Deploy

### Console

Open the deployment store, find **Tekton Operator**, and click **Deploy**. The creation wizard walks you through the installation contract (the fixed namespace, the pinned release, the one-per-cluster rule, the CRD lifecycle), the air-gap image overrides, sizing, and scheduling. Start from the **Tekton Operator preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator
  org: acme-corp
  env: prod
spec: {}
```

```shell
planton apply -f tekton-operator.yaml
```

An empty spec is the complete install: the release manifest's own defaults, in its fixed namespace, with automatic component installation disabled. Declare a **Tekton** resource next to choose what actually runs. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring the Tekton Operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**There is no version field — deliberately** — the installed operator (and the TektonConfig schema the KubernetesTekton kind renders against) is pinned to the release this catalog was designed against (`v0.80.0`). A user-selectable version would silently drift the TektonConfig surface away from what the catalog models. Operator upgrades arrive with catalog releases, not spec edits.

**There is no namespace field either** — the release manifest installs into `tekton-operator`, and that name is baked into the manifest's own cross-references. Exactly one install per cluster is the upstream contract.

**The CRDs delete with this resource** — the 14 `operator.tekton.dev` CRDs are documents of the applied manifest, so destroying the operator removes them, which CASCADE-DELETES any TektonConfig on the cluster. Always destroy the KubernetesTekton resource FIRST: its teardown blocks until the operator finishes removing the components, and the `TektonInstallerSet` finalizers are processed only by a RUNNING operator — removing the operator first strands them.

**`imageRegistry` is the mirror seam** — one registry root replaces `ghcr.io` for every image Tekton publishes: the operator and its webhook, and every component the operator installs (Pipelines, Triggers, Dashboard, Chains, Results, the pruner), including the images every build pod starts with. Each image keeps its path, tag and digest, so a mirror or pull-through cache of `ghcr.io` serves them all and can serve only the bytes the release names; the images move with the pinned release, so a catalog upgrade never freezes one. Images Tekton does not publish keep their registry: the shell images script steps start with (`cgr.dev`, `mcr.microsoft.com`) and Results' bundled Postgres (Docker Hub). Confirm the mirror is in use by reading what the cluster runs: `kubectl -n tekton-pipelines get deploy tekton-pipelines-controller -o jsonpath='{.spec.template.spec.containers[0].image} {.spec.template.spec.containers[0].args}'` names the mirror for the controller and the `-entrypoint-image`, `-nop-image`, `-workingdirinit-image` and `-sidecarlogresults-image` it hands every build.

**Image overrides pin one exact image** — `operatorImage` overrides the image for BOTH containers of the operator Deployment, `webhookImage` the admission webhook's, and each wins over `imageRegistry` for its image. An override freezes that image while the next catalog upgrade moves the rest of the release, so prefer `imageRegistry` and reach for an override only for a rebuilt operator image. `imagePullSecrets` names existing `kubernetes.io/dockerconfigjson` Secrets in the fixed `tekton-operator` namespace — references, never credentials.

**Sizing and placement** — the manifest sets NO resource requests or limits (the operator runs unbounded); set `operatorResources` / `webhookResources` on production clusters with quotas. `nodeSelector` and `tolerations` steer the operator and webhook pods — scheduling for the Tekton COMPONENT pods lives on the KubernetesTekton resource's placement instead.

## Outputs and Dependencies

### What This Component Consumes

This component's spec is self-contained — no fields reference other resources' outputs, and it has no cluster-side prerequisites beyond registry reachability.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in (always `tekton-operator` — fixed by the release manifest) | Composition, debugging |

The operator exports no component handles of its own: the Tekton namespace, profile, and dashboard endpoints are the KubernetesTekton resource's outputs — this installation is only the manager that reconciles it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Operator** — the complete install with an empty spec, which is deliberately tiny: the operator is a lifecycle manager, not the product. Set `imageRegistry` on clusters that pull Tekton from a mirror of `ghcr.io` and resource requests on quota-governed ones. Start from the **Tekton Operator preset**.

## Works With

- [**Tekton**](/cloud-catalog/kubernetes-tekton) — the cluster's TektonConfig declaration this operator reconciles; deploy the operator FIRST, and destroy the declaration FIRST on the way out.
- [**Kubernetes Manifest**](/cloud-catalog/kubernetes-manifest) — Tasks, Pipelines, and their runs are plain custom resources once the Tekton installation converges.
