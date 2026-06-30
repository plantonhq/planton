# Istio GatewayClass

This preset creates a GatewayClass for the Istio gateway controller. Istio is one of the most widely deployed Gateway API implementations and a common choice for production ingress and service mesh traffic.

## When to Use

- Your cluster runs Istio with the Gateway API controller enabled
- You want Gateways to be provisioned and managed by Istio
- You are following the standard Planton ingress pattern (the leftbin reference setup)

## Key Configuration Choices

- **controllerName** (`istio.io/gateway-controller`) -- the identity Istio's Gateway controller watches for; copied verbatim from Istio's documentation
- **No parametersRef** -- Istio works with sensible defaults; add a ConfigMap reference only if you need controller-specific tuning
- **Cluster-scoped** -- GatewayClass is cluster-wide; no namespace is set

## Prerequisites

- Gateway API CRDs installed (`KubernetesGatewayApiCrds`)
- Istio installed with the Gateway API controller enabled

## Placeholders to Replace

No placeholders -- this preset is directly deployable. Rename `metadata.name` to match your naming convention (the name becomes the GatewayClass that Gateways reference via `gatewayClassName`).
