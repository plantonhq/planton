# KubernetesTektonOperator Guide

The judgment this guide carries: this operator is a fixed-name singleton
that installs Tekton's LIFECYCLE MANAGER and nothing else — the cluster's
actual Tekton is a separate declaration, and the two have a strict
teardown order.

## Manager only; one per cluster

The operator installs into a fixed `tekton-operator` namespace with
upstream-fixed resource names — exactly one per cluster, and it is
installed with automatic component installation DISABLED so the
[KubernetesTekton](../kubernetestekton/GUIDE.md) declaration is the
single owner of the cluster's TektonConfig. Installing this operator alone
deploys no Tekton components. The prerequisite relationship (no diagram
edge) is the
[operator-prerequisite pattern](../../_patterns/operator-prerequisite.md)'s
singleton posture.

## No version field, and destroy order matters

No version selector by design: the operator and the TektonConfig schema
the KubernetesTekton kind renders against are pinned together, so a
user-selectable version would drift the surface (the reference page
explains). And never destroy this operator while a KubernetesTekton
exists — its teardown finalizers need a running operator, or deletion
hangs. The destroy-order rule is stated from the config side too; honor it
in any teardown proposal.

## Pulling from a mirror: `image_registry`, not the image overrides

When ghcr.io is slow or unreachable from a cluster, set `image_registry`
to a mirror or pull-through cache of ghcr.io. It moves every image Tekton
publishes — the operator's own and, through the operator, every
component's, including the images each build pod starts with — at the
same path and digest, and it follows the pinned release on every catalog
upgrade. Do not propose `operator_image`/`webhook_image` for this: each
pins one exact image, moves nothing the operator installs, and freezes
while the next upgrade moves the rest. A private mirror needs
`image_pull_secrets` for the operator's own images; the component images
are pulled in the Tekton namespace with that namespace's credentials, so
a cache readable by the cluster's nodes is the simplest fit. The reference
page lists what does not move (images Tekton does not publish). Prove the
mirror is in use from what the cluster runs, never from the manifest:
`kubectl -n tekton-pipelines get deploy tekton-pipelines-controller -o jsonpath='{.spec.template.spec.containers[0].image} {.spec.template.spec.containers[0].args}'`
names the mirror for the controller and for the `-entrypoint-image`,
`-nop-image`, `-workingdirinit-image` and `-sidecarlogresults-image` it
hands every build.

## Namespace ownership — the infra exception

The fixed `tekton-operator` namespace is a dedicated sole-tenant one — the
[namespace-ownership pattern](../../_patterns/namespace-ownership.md)'s
normal infra case.

## Pairs well with

- KubernetesTekton — the one TektonConfig this operator reconciles (see
  its [guide](../kubernetestekton/GUIDE.md)).
