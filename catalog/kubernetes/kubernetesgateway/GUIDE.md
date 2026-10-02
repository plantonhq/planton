# KubernetesGateway Guide

The judgment this guide carries: the Gateway API family is fully modeled
in this catalog as exact upstream mirrors — so the platform knowledge is
not in any one kind's fields, it is in the CHAIN: which pieces must exist,
in what order, and above all WHICH CONTROLLER actually implements the
gateway. A declared Gateway with no implementing controller is valid,
inert, and silently unreachable.

## The chain, in order

1. **KubernetesGatewayApiCrds** — the CRDs are not on clusters by
   default; this kind installs them, and it is a registry prerequisite of
   both GatewayClass and Gateway (the invisible-edge mechanism:
   [operator-prerequisite pattern](../../_patterns/operator-prerequisite.md)).
2. **An implementing controller** — the crown judgment, next section.
3. **KubernetesGatewayClass** — names the controller
   (`controllerName`, immutable once created — its reference page has the
   vocabulary).
4. **This kind** — binds listeners (ports, protocols, TLS) under that
   class; TLS listeners consume certificate Secrets that
   [KubernetesCertificate](../kubernetescertificate/GUIDE.md)
   materializes.
5. **Route kinds** — HTTPRoute, GRPCRoute, TCPRoute, UDPRoute, TLSRoute,
   ListenerSet all attach through `parentRefs`, a foreign key the
   platform draws as a real edge. Their per-kind semantics are upstream
   Gateway API, faithfully mirrored — their reference pages are the
   complete authority; no per-route guide exists on purpose.

## The implementing controller — check before declaring anything

The catalog's two grounded answers:

- **KubernetesIstio** implements Gateway API natively — an Istio cluster
  deploys NO separate ingress; declaring a KubernetesGateway with
  `gatewayClassName: istio` is how istiod provisions the actual gateway
  (Istio's own reference page states this). On Istio clusters the Gateway
  family IS the north-south path.
- **KubernetesCilium** provides it behind its `gatewayApi` flag, which
  creates the "cilium" GatewayClass — and requires `kubeProxyReplacement`
  plus the CRDs (its reference page carries both constraints).

On Istio, each Gateway gets its own Deployment and LoadBalancer Service,
named `<gateway>-istio`, which is how a second front door gets its own
load balancer. Those pods are Istio's to create, so their priority and
resources come from `infrastructure.parameters_ref`: a
`KubernetesConfigMap` in the Gateway's namespace whose `deployment` key is
a strategic-merge overlay on the generated Deployment (for example
`spec.template.spec.priorityClassName`, or the `istio-proxy` container's
resources). Name it as a plain string and keep an explicit `depends_on`
on the ConfigMap, since no reference carries the edge. An Istio policy
attached through `target_refs` must live in the Gateway's own namespace.

On a cluster with neither — for example, one whose entry point is
ingress-nginx — Gateway declarations sit unclaimed forever. That cluster
wants the Ingress lane instead: the comparison lives in the
[KubernetesIngress guide](../kubernetesingress/GUIDE.md).

## On the diagram

Routes draw real `parentRefs` edges into this Gateway, and TLS listeners
draw certificate edges — the north-south topology is visible. What is
NOT drawn is the implementing-controller coupling (class name matching a
controller that exists) — the reviewer verifies that deliberately, same
as every prerequisite.

## Pairs well with

- KubernetesGatewayApiCrds + KubernetesGatewayClass — the chain below
  this kind.
- KubernetesHttpRoute and the other route kinds — attachment via
  `parentRefs`.
- KubernetesIstio / KubernetesCilium — the implementing controllers.
- KubernetesCertificate — listener TLS Secrets.
- KubernetesExternalDns — publishes the hostnames routes serve.
