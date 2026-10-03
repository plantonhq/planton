# Kubernetes Keycloak Operator

## When NOT to Use This

**One resource is ONE Keycloak Operator install** — the official
operator that reconciles `KubernetesKeycloak` declarations into
running Keycloak StatefulSets. Every bundle resource carries
upstream's fixed names (`keycloak-operator` and friends): exactly one
install per namespace, by construction.

Not the right kind when:

- **You want Keycloak itself** — that is `KubernetesKeycloak`: the
  server declaration this operator reconciles. Installing the
  operator alone deploys NO Keycloak server — by design.
- **You want to pick the operator version** — there is no version
  field by design: the `KubernetesKeycloak` declaration renders its
  CR against the CRD schema this bundle installs, and a selectable
  version would drift them apart. The module pins the release;
  upgrades arrive as module updates.
- **You want several watchers** — run namespaced operators (one per
  namespace, each watching its own) or at most ONE cluster-wide
  operator per cluster.

## Distribution

Keycloak ships NO official Helm chart — the operator IS the
first-party Kubernetes distribution. This module installs the
keycloak-k8s-resources release manifests tag-pinned: 16 plain-YAML
documents (ServiceAccount, RBAC, Service, the operator Deployment)
plus 4 CRDs. No admission webhooks, no cert-manager dependency, no
install hooks; documents apply as ordered groups so creation and
teardown order correctly by construction.

## The namespace stamp

Upstream expects kustomize to set the namespace; the module stamps
the target namespace onto every namespaced document AND every RBAC
binding subject — one ClusterRoleBinding subject is baked to the
`keycloak` namespace upstream, which the stamp fixes.

## Watch scope

Default: the operator watches ONLY its own namespace —
`KubernetesKeycloak` resources live beside it, and several teams run
isolated operator+Keycloak stacks in separate namespaces.
`cluster_wide: true` switches to upstream's cluster-wide variant; the
ONLY differences are the binding kinds (per-controller
ClusterRoleBindings) and the JOSDK env values. Know the upstream
constraint: in cluster-wide mode the operator refuses custom
ServiceAccounts on Keycloak pod templates.

## The default server image

`default_keycloak_image` overrides the bundle's
`RELATED_IMAGE_KEYCLOAK` — the server image the operator stamps into
every Keycloak StatefulSet whose declaration sets no image of its
own. Air-gapped clusters point it (and `operator_image`) at their
mirrors, keeping the pinned tags.

## Destroy

The 4 CRDs delete WITH the resource — which cascade-deletes any
Keycloak declarations still in the cluster. Destroy
`KubernetesKeycloak` resources first, this operator after.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
