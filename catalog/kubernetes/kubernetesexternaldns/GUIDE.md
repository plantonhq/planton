# KubernetesExternalDns Guide

The judgment this guide carries: a public endpoint has TWO silent
prerequisites — the certificate chain signs it, and this controller makes
its name resolvable. Both fail the same way: nothing in the endpoint's own
manifest errors; the architecture is simply incomplete. When a proposal
includes a hostname, it must include the machinery that publishes it.

## Where it sits in the public-endpoint chain

ExternalDNS watches the cluster's Ingress and route resources and writes
their hostnames into a DNS provider's zone. Without it, the hostname on an
Ingress is a label nothing publishes — the endpoint serves traffic only
for clients that already know the load balancer's address. It belongs in
the shared-cluster chart beside the
[ingress controller](../kubernetesingressnginx/GUIDE.md) and the
[certificate chain](../kubernetesclusterissuer/GUIDE.md).

## Routes on listener sets need the listener-set switch

When components bring their own hostname and certificate to a shared
Gateway as a [KubernetesListenerSet](../kuberneteslistenerset/README.md)
(a monitoring hub beside a platform on one Gateway is the common case),
the controller writes their records only with `gateway_listener_sets:
true`. Without it, a route attached to the listener set reads Accepted and
no record ever appears: the silent prerequisite again. The switch needs a
`gateway-*` entry in `sources`, because the chart grants the permission to
read ListenerSets only alongside a Gateway API route source (the spec
refuses the switch without one), and the Gateway itself must admit
listener sets (`allowed_listeners`).

## One instance per DNS provider, each owning its records

One installation writes to exactly one DNS provider; clusters publishing
to several run several instances (the instancing and `txtOwnerId`
ownership rules are on [reference.md](v1alpha1/reference.md)). The judgment: when
adding a second instance, give it its own `txtOwnerId` FIRST — instances
sharing an owner id fight over record ownership, and the damage lands in
the DNS zone, far from either manifest.

## Namespace ownership — the infra exception

A dedicated "external-dns" namespace with `createNamespace: true` is the
normal single-tenant shape — the
[namespace-ownership pattern](../../_patterns/namespace-ownership.md)'s
sole-tenant case.

## On the diagram

ExternalDNS renders as a shared-cluster node. It is the piece reviewers
use to answer "how does this hostname become real?" — an architecture
with public hostnames and no DNS publisher is showing names it cannot
resolve.

## Pairs well with

- KubernetesIngressNginx — the entry point whose Ingress hostnames this
  controller publishes.
- KubernetesCertManager + KubernetesClusterIssuer — the HTTPS half of the
  same completeness story.
