# Kubernetes Request Authentication

Provision an Istio `RequestAuthentication` -- the mesh primitive that defines which
JSON Web Tokens (JWTs) are accepted on your workloads. Validate end-user / caller
tokens from one or more issuers at the mesh layer, surface verified identities to
authorization policies, and forward selected claims to your backends.

## What Gets Created

- A namespaced `security.istio.io/v1` `RequestAuthentication` custom resource.
- A set of `jwt_rules` plus an optional workload `selector` or `target_refs`.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to enforce the policy.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRequestAuthentication
metadata:
  name: jwt-auth
spec:
  namespace:
    value: finance
  jwt_rules:
    - issuer: https://accounts.example.com
      jwks_uri: https://accounts.example.com/.well-known/jwks.json
```

```bash
openmcf apply -f requestauthentication.yaml
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
| `jwt_rules` | list | JWT rules: `issuer`, `jwks_uri`/`jwks`, token locations, audiences, claim forwarding, `timeout`. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `request_authentication_name` | Name of the created RequestAuthentication (equals metadata.name). |
| `namespace` | Namespace the RequestAuthentication was created in. |

## Related Components

- [Kubernetes Peer Authentication](kubernetespeerauthentication)
- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
