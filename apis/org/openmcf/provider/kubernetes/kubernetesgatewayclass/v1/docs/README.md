# KubernetesGatewayClass: Research Documentation

## Introduction

The `GatewayClass` is the root of the Kubernetes Gateway API resource model. It
is a cluster-scoped resource that declares which controller implementation is
responsible for a class of Gateways, much like a `StorageClass` declares which
provisioner backs a class of volumes. This document explains the Gateway API
role model, why GatewayClass deserves to be a first-class OpenMCF component, the
design decisions behind this component, and how it composes with the rest of the
OpenMCF Gateway API family.

## The Gateway API Role-Oriented Model

The Gateway API was designed by Kubernetes SIG-Network to fix the structural
limitations of the Ingress API: annotation sprawl, provider lock-in, HTTP-only
routing, and the lack of separation between infrastructure and application
concerns. It splits responsibilities across three personas:

```
Infrastructure Provider  ->  GatewayClass   (which controller, with what defaults)
Cluster Operator         ->  Gateway        (where traffic enters, listeners, TLS)
Application Developer     ->  HTTPRoute/...  (how traffic is routed to backends)
```

`GatewayClass` is the infrastructure-provider artifact. It names a controller and
optionally points at a controller-specific configuration resource. Every
`Gateway` references exactly one `GatewayClass` through `spec.gatewayClassName`,
and the controller that owns the class is the one that provisions the Gateway's
data plane.

### Where GatewayClass sits

```
GatewayClass (cluster-scoped)         "istio" -> istio.io/gateway-controller
   ^ referenced by gatewayClassName
Gateway (namespaced)                  listeners :443 HTTPS, TLS termination
   ^ referenced by parentRefs
HTTPRoute / GRPCRoute / TLSRoute / TCPRoute (namespaced)
```

## Anatomy of GatewayClassSpec

Upstream `GatewayClassSpec` (gateway-api v1.5.1, `apis/v1/gatewayclass_types.go`)
has exactly three fields:

| Field | Upstream type | Required | Notes |
|-------|---------------|----------|-------|
| `controllerName` | `GatewayController` (string) | Yes | Domain-prefixed path; immutable; 1-253 chars |
| `parametersRef` | `*ParametersReference` | No | Reference to controller-specific config (ConfigMap or CRD) |
| `description` | `*string` | No | Max 64 characters |

### controllerName

The controller name is a domain-prefixed path such as
`istio.io/gateway-controller` or `gateway.envoyproxy.io/gatewayclass-controller`.
It is the contract between a GatewayClass and the controller that watches for it:
a controller only reconciles GatewayClasses whose `controllerName` matches its own
identity. Upstream marks this field immutable
(`+kubebuilder:validation:XValidation: self == oldSelf`): once a GatewayClass
exists, its controller cannot be changed -- you must delete and recreate.

### parametersRef

Some controllers accept additional configuration via a referenced resource. For
example, Envoy Gateway uses an `EnvoyProxy` custom resource; other controllers
use a `ConfigMap`. `parametersRef` is a structured reference
(group/kind/name/namespace), not a single-value identifier, so it is modeled as a
shared structured message rather than a foreign key (see Design Decisions).

### description

A free-form, human-friendly label, capped at 64 characters upstream.

## Why GatewayClass Is a First-Class OpenMCF Component

A reasonable objection is that controllers (Istio, Envoy Gateway) often install
their own default GatewayClass, so why model it separately? Three reasons:

1. **Visibility and inventory.** Making GatewayClass a resource gives platform
   teams a declarative record of which classes exist across a fleet of clusters.
2. **Self-managed classes.** Teams that want custom classes (custom parameters,
   non-default controllers) can declare them as code rather than `kubectl apply`.
3. **Foreign-key wiring.** `KubernetesGateway.spec.gateway_class_name` can
   reference `KubernetesGatewayClass.status.outputs.gateway_class_name`, letting
   an InfraChart deploy the class before the Gateway and wire them automatically.

The spec is tiny (three fields), so the cost of providing this surface is low and
the composability payoff is high.

## Design Decisions

### 100% upstream fidelity

Per the project's design decision DD-001, this component mirrors the upstream
spec exactly rather than subsetting it. Gateway API is an external standard, so
"the what" the user wants to express *is* the upstream spec. All three fields are
present with upstream semantics and validation.

### Cluster-scoped: no namespace

GatewayClass is `+kubebuilder:resource:scope=Cluster` upstream. The OpenMCF spec
therefore carries `target_cluster` but deliberately has **no** `namespace` field.
The Pulumi and Terraform modules never set or create a namespace.

### controllerName immutability is documented, not enforced at the proto layer

`buf.validate` evaluates a single message instance and has no access to the prior
value (`oldSelf`), so upstream's `self == oldSelf` rule cannot be expressed as a
proto validation. Immutability is documented in the field comment and enforced by
the Gateway API admission webhook at apply time. Enforcing it inside the OpenMCF
control plane would be a separate concern, out of scope for this component.

### parametersRef is a structured reference, not a foreign key

OpenMCF foreign keys (`StringValueOrRef`) are for single-value identifiers or
outputs of *other OpenMCF resources*. `parametersRef` is a multi-field reference
(group/kind/name/namespace) to an arbitrary Kubernetes object that is usually not
an OpenMCF resource (a ConfigMap, or an implementation CRD like `EnvoyProxy`).
Wrapping it in a foreign key would distort the upstream structure and has no
single default kind to point at, so it reuses the shared
`KubernetesGatewayApiParametersReference` message instead.

### Typed crd2pulumi resource

The Pulumi module uses the typed `gatewayv1.NewGatewayClass` from the
crd2pulumi-generated SDK rather than an untyped `CustomResource`. This catches
field-name and structure errors at compile time and matches how every other
OpenMCF ingress component consumes the Gateway API types.

## Controller Landscape

| Controller | controllerName | parametersRef kind |
|------------|----------------|--------------------|
| Istio | `istio.io/gateway-controller` | (none / ConfigMap) |
| Envoy Gateway | `gateway.envoyproxy.io/gatewayclass-controller` | `EnvoyProxy` |
| NGINX Gateway Fabric | `gateway.nginx.org/nginx-gateway-controller` | `NginxProxy` |
| Cilium | `io.cilium/gateway-controller` | (none) |
| GKE Gateway | `networking.gke.io/gateway` | (none) |

Controller names are copied verbatim from each implementation's documentation;
this is why the OpenMCF field preserves the exact upstream string rather than an
enum.

## How It Works

1. The module renders a cluster-scoped `GatewayClass` CR named after
   `metadata.name`, with `controllerName`, optional `description`, and optional
   `parametersRef`.
2. The matching controller observes the new GatewayClass and sets the `Accepted`
   status condition to `True` (or `False` with a reason such as
   `InvalidParameters` or `Unsupported`).
3. Downstream `KubernetesGateway` resources reference the class by name. The
   InfraChart DAG, using the `gateway_class_name` stack output as a foreign-key
   target, guarantees the class is created before any Gateway that uses it.

## 80/20 Scoping

**In scope:** the full GatewayClass spec (controllerName, parametersRef,
description), cluster-scoped lifecycle, typed IaC in Pulumi and Terraform,
foreign-key output for Gateway wiring.

**Out of scope:** installing the controller itself (use the controller's own
component or a Helm release), installing the CRDs (use
`KubernetesGatewayApiCrds`), and reading back the live `Accepted`/`SupportedVersion`
status conditions (these are controller-managed and observed via `kubectl`, not
stored in OpenMCF outputs).

## Common Pitfalls

### GatewayClass stuck without an Accepted condition

**Problem:** The GatewayClass is created but never accepted.
**Cause:** No controller in the cluster owns the specified `controllerName`, or
the controller is not yet running.
**Solution:** Install and run the matching controller; verify `controllerName`
exactly matches the controller's identity string.

### Trying to change controllerName

**Problem:** Updating `controllerName` is rejected.
**Cause:** The field is immutable upstream.
**Solution:** Delete and recreate the GatewayClass with the new controller.

### Setting parametersRef.namespace for a cluster-scoped parameters resource

**Problem:** The controller rejects the parametersRef.
**Cause:** `namespace` must be unset for cluster-scoped parameter resources and
set for namespace-scoped ones.
**Solution:** Only populate `parametersRef.namespace` for namespaced resources
such as ConfigMaps.

## Conclusion

GatewayClass is the small but foundational root of the Gateway API model.
Modeling it as a first-class OpenMCF component yields fleet-wide visibility,
declarative management of custom classes, and clean foreign-key wiring into
`KubernetesGateway`. With a three-field spec mirrored at 100% fidelity and typed
IaC in both Pulumi and Terraform, this component closes the first gap in the
OpenMCF Gateway API networking layer.

## References

- [Gateway API GatewayClass](https://gateway-api.sigs.k8s.io/api-types/gatewayclass/)
- [Gateway API Official Documentation](https://gateway-api.sigs.k8s.io/)
- [Gateway API GitHub Repository](https://github.com/kubernetes-sigs/gateway-api)
- [Gateway API Implementations](https://gateway-api.sigs.k8s.io/implementations/)
