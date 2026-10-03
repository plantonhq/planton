# Keycloak Operator

Installs the official Keycloak Operator from the pinned keycloak-k8s-resources release manifests. Keycloak ships **no official Helm chart** — the operator IS the first-party Kubernetes distribution. It reconciles Keycloak declarations (declared with **Keycloak**) into running Keycloak StatefulSets, managing their Services, network policy, and the one-time bootstrap admin credential.

This component installs the **manager only**. Installing it deploys NO Keycloak server: declare `KubernetesKeycloak` resources and the operator turns each into a running server.

## What Gets Created

When you deploy this Infra Component, the IaC module fetches the release manifests tag-pinned, stamps your namespace onto every namespaced document (upstream expects kustomize to do this), and applies them as ordered groups:

- **16 plain-YAML documents** — the `keycloak-operator` ServiceAccount, RBAC, the metrics/health Service (port 80 → 8080), and the operator Deployment. No admission webhooks, no cert-manager dependency, no install hooks
- **4 `k8s.keycloak.org` CRDs** (`keycloaks`, `keycloakrealmimports`, `keycloakoidcclients`, `keycloaksamlclients`) — documents of the applied manifest, so they install AND delete with this resource; see the destroy ordering under Key Configuration
- **The namespace** — created with the standard Planton governance labels when `createNamespace` is true; otherwise it must already exist
- **Fixed names throughout** — every object carries upstream's own `keycloak-operator` name (not derived from this resource's name), so exactly ONE operator install fits per namespace

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **No existing install in the namespace** — the bundle's fixed object names mean one operator per namespace; a second install would fight over the same objects. Cluster-wide installs (see Key Configuration) are additionally a one-per-cluster singleton.
- **Registry reachability** — the images pull from `quay.io/keycloak/*`. Air-gapped clusters set the two image overrides (see Key Configuration).

## Deploy

### Console

Open the deployment store, find **Keycloak Operator**, and click **Deploy**. The creation wizard walks you through the installation contract (the pinned release, the fixed names, the CRD lifecycle), namespace placement, the watch scope, the air-gap image mirrors, sizing, and scheduling. Start from the **Operator preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeycloakOperator
metadata:
  name: keycloak-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: keycloak
  createNamespace: true
```

```shell
planton apply -f keycloak-operator.yaml
```

The namespace is the only required field: everything else is an optional override of the bundle's own defaults. Declare a **Keycloak** resource next — in this same namespace under the default watch — to get a server. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the installation namespace from the namespace resource the whole identity stack shares:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: identity
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline creates the namespace first, then installs the operator into it — the Keycloak declarations and their database follow in the same namespace.

## Key Configuration

These are the most important decisions when configuring the Keycloak Operator installation. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**There is no version field — deliberately** — the `KubernetesKeycloak` declaration kind's CR rendering is built against the CRD schema this bundle installs (release `26.7.0`). A user-selectable operator version would silently drift that schema away from what the declaration kind renders. The module pins the release; upgrades arrive as module updates, not spec edits.

**The CRDs delete with this resource** — the 4 `k8s.keycloak.org` CRDs are ordinary manifest documents with no keep annotation: destroying the operator removes them, which CASCADE-DELETES every KubernetesKeycloak declaration (and its server) on the cluster. This is the OPPOSITE posture from kept-CRD operators like Gatekeeper or the GHA runner controller. Always destroy the Keycloak declarations FIRST, while the operator still runs to process their finalizers.

**Watch scope is the one topological fork** — `cluster_wide: false` (the default) watches ONLY the operator's own namespace: KubernetesKeycloak resources must live beside it, and several teams can run isolated operator+Keycloak stacks in separate namespaces. `true` applies upstream's cluster-wide bundle variant (per-controller ClusterRoleBindings, `JOSDK_ALL_NAMESPACES`): run at most one per cluster — and know the upstream constraint that cluster-wide mode REFUSES custom ServiceAccounts on Keycloak pod templates.

**Image overrides are the air-gap seam** — `operatorImage` overrides the operator container (default `quay.io/keycloak/keycloak-operator` at the pinned release); `defaultKeycloakImage` overrides the bundle's `RELATED_IMAGE_KEYCLOAK` — the DEFAULT server image the operator stamps into Keycloak StatefulSets whose declaration sets no image. Mirror the repository, keep the tag: a different tag drifts the operator away from the CRD schema this bundle installs.

**Sizing and placement** — upstream ships real defaults for the operator container (requests `300m`/`450Mi`, limits `700m`/`450Mi`); empty `resources` keeps them. `scheduling.nodeSelector` and `scheduling.tolerations` steer the operator pod — scheduling for the Keycloak SERVER pods lives on each KubernetesKeycloak declaration instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in — where namespaced-watch Keycloak declarations must also live | Composition, co-locating KubernetesKeycloak resources |
| `deployment` | The operator Deployment name (upstream-fixed: `keycloak-operator`) | Debugging, monitoring |
| `service` | The operator's metrics/health Service (upstream-fixed: `keycloak-operator`, port 80 → 8080) | Scrape configuration, health checks |

The operator exports no Keycloak server handles of its own: each server's Services, endpoints, and admin credential Secret are the KubernetesKeycloak declaration's outputs — this installation is only the manager that reconciles them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Operator** — the default posture: installed into the `keycloak` namespace with the namespaced watch, so the operator reconciles only Keycloak declarations living beside it. Air-gapped clusters point the two image overrides at their mirrors, keeping the pinned tags. Start from the **Operator preset**.

## Works With

- [**Keycloak**](/infra-catalog/kubernetes-keycloak) — the server declarations this operator reconciles; deploy the operator FIRST (in the same namespace under the default watch), and destroy the declarations FIRST on the way out
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the installation namespace by reference
- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) — the production database a Keycloak declaration composes; co-locate it with the operator+server namespace so credential Secrets are readable
