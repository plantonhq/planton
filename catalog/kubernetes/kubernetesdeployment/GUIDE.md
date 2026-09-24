# KubernetesDeployment Guide

The judgment this guide carries: a Deployment is almost never the whole
architecture. It deliberately exports handles instead of embedding
exposure, autoscaling beyond the basics, or namespace ownership — the
completeness of a proposal depends on composing the right neighbors, not
on filling more of this spec.

## Exposure is composed — finish the chain

This kind creates the Service but nothing that makes it reachable from
outside the cluster. When the user's app needs a public endpoint, the
proposal must also include an exposure kind (KubernetesIngress, or
KubernetesHttpRoute behind a Gateway) pointed at the Service this kind
exports — `status.outputs.kube_endpoint` is the in-cluster handle, and
`status.outputs.selector_labels` feeds NetworkPolicy pod selectors. Public
HTTPS additionally needs the certificate chain; the composition checklist
lives in the [KubernetesClusterIssuer guide](../kubernetesclusterissuer/GUIDE.md).

## Three ways to autoscale — pick by scale signal

- `spec.availability.horizontalPodAutoscaling` (inline) covers CPU and
  memory targets — right for the common case, zero extra nodes.
- The standalone KubernetesHorizontalPodAutoscaler kind carries the full
  autoscaling/v2 surface (custom metrics, external metrics, scaling
  behavior policies) and references this kind's
  `status.outputs.deployment_name`.
- KubernetesKeda is the event-driven lane (queue depth, stream lag,
  cron — and scale-to-ZERO, which neither HPA lane can do).

The full three-lane comparison lives in the
[KubernetesHorizontalPodAutoscaler guide](../kuberneteshorizontalpodautoscaler/GUIDE.md)
— and never declare more than one lane for the same workload: two
controllers fighting over one replica count.

## Namespace ownership

`spec.namespace` is a required foreign key targeting KubernetesNamespace.
`createNamespace: true` makes this workload the namespace's owner in IaC
state — safe only when it is the sole tenant. Any shared namespace wants a
dedicated KubernetesNamespace component instead; the failure story and the
`valueFrom` wiring: [namespace-ownership pattern](../../_patterns/namespace-ownership.md).

## Variables vs secrets is a storage decision

`env.variables` values are written into the Deployment's pod template,
readable by anyone who can read the workload. `env.secrets` values are
collected into one Kubernetes Secret this component owns
(`<name>-env-secrets`) and read through `secretKeyRef`. So a credential goes
in `env.secrets` -- on Planton as `value: $secret/<slug>`, which the
platform resolves just before the module runs. The same reference in
`env.variables` is refused before anything deploys: resolved there, it
would land in the pod template as the secret itself. Reach for
`env.secrets[].secretRef` instead when the Secret already exists and has
its own owner (an external secrets operator, another team's rotation).

## Pipeline-attached services

When this resource is a Service Hub deploy target, the pipeline owns
`spec.version` and `spec.container.app.image` (the deploy-target contract
on [reference.md](v1alpha1/reference.md)). Compose the initial manifest with a real
image, but do not build automation that edits those two paths out-of-band
— the pipeline writes them on every build.

## On the diagram

Each composed piece — namespace, ingress/route, certificate chain,
standalone autoscaler — renders as its own node with reference edges into
this one. An architecture that buries exposure in annotations or namespace
creation in a flag shows a lone deployment node and hides exactly the
infrastructure a reviewer needs to see.

## Pairs well with

- KubernetesNamespace — the namespace owner (pattern above).
- KubernetesIngress / KubernetesHttpRoute — external exposure, wired to
  the exported Service.
- KubernetesHorizontalPodAutoscaler — advanced autoscaling (judgment
  above).
- KubernetesConfigMap / KubernetesSecret / KubernetesExternalSecret —
  configuration and credentials mounted by the containers.
