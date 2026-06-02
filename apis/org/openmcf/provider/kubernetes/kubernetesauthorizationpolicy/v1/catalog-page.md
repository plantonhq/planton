# Kubernetes Authorization Policy

Provision an Istio `AuthorizationPolicy` -- the mesh primitive that enforces access
control on your workloads. Allow, deny, or audit requests by source identity,
operation, and conditions, or delegate the decision to an external authorizer.

## What Gets Created

- A namespaced `security.istio.io/v1` `AuthorizationPolicy` custom resource.
- An `action` (ALLOW / DENY / AUDIT / CUSTOM) applied to requests matched by `rules`,
  with an optional workload `selector` or `target_refs`.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to enforce the policy.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesAuthorizationPolicy
metadata:
  name: require-jwt
spec:
  namespace:
    value: finance
  action: ALLOW
  rules:
    - from:
        - source:
            request_principals:
              - "*"
```

```bash
openmcf apply -f authorizationpolicy.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the policy is created in. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `selector.match_labels` | map | Pod labels selecting target workloads; omit for namespace-wide scope. Mutually exclusive with `target_refs`. |
| `target_refs` | list | Gateway / Service / ServiceEntry resources to bind to; required for waypoints. Mutually exclusive with `selector`. |
| `rules` | list | Match rules: `from` (source), `to` (operation), `when` (conditions). |
| `action` | string | `ALLOW` (default), `DENY`, `AUDIT`, or `CUSTOM`. |
| `provider.name` | string | MeshConfig extension provider for the CUSTOM action. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `authorization_policy_name` | Name of the created AuthorizationPolicy (equals metadata.name). |
| `namespace` | Namespace the AuthorizationPolicy was created in. |

## Related Components

- [Kubernetes Request Authentication](kubernetesrequestauthentication)
- [Kubernetes Peer Authentication](kubernetespeerauthentication)
- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
