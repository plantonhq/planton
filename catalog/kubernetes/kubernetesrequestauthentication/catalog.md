# Istio Request Authentication

Defines an Istio RequestAuthentication: a namespaced policy that decides which JSON Web Tokens (JWTs) are accepted on your workloads. Use it to validate end-user or service tokens from one or more identity providers at the mesh proxy, extract the verified identity for authorization, and reject requests that present an invalid token. Know its one behavioral trap up front: it validates a token when one is PRESENT -- a request with no token passes through, so requiring a token takes a paired Authorization Policy.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **A RequestAuthentication policy** -- a namespaced Istio policy that tells the mesh how to validate JWTs presented to the workloads it selects: which issuers to trust (one JWT rule per issuer, each with its own signing keys and allowed audiences), where to read the token from (Authorization header, custom headers, query parameters, or cookies), and what to do with the verified claims (forward the token, emit the payload, or copy claims into request headers).
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Istio installed** -- the Istio CRDs (Istio Base CRDs) must be present and the Istio control plane (istiod) running. The policy is only enforced where istiod is active and the selected workloads are part of the mesh.
- **Target namespace exists** -- the policy is created in a specific namespace; reference an existing one or create it first.
- **Reachable signing keys** -- istiod must be able to fetch the issuer's JWKS (via discovery or the configured URL) to verify token signatures.

## Deploy

### Console

Open the deployment store, find **Istio Request Authentication**, and click **Deploy**. The creation wizard walks you through the namespace, the scope (namespace-wide, a workload selector, or target references), and the JWT rules, with guidance at each step. Start from the **Validate JWTs Across a Namespace** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRequestAuthentication
metadata:
  name: jwt-auth
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  jwtRules:
    - issuer: https://accounts.google.com
      audiences:
        - my-app.acme.com
```

```shell
planton apply -f request-authentication.yaml
```

This validates Google-issued JWTs for every workload in the `prod-apps` namespace. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the policy to its Planton-managed namespace:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: prod-apps-namespace
      fieldPath: spec.name
  jwtRules:
    - issuer: https://accounts.google.com
```

The InfraPipeline deploys the namespace first, then creates the policy inside it.

## Key Configuration

These are the most important decisions when configuring a RequestAuthentication policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Validation is not requirement** -- a valid token yields an authenticated principal and an invalid token is rejected, but a request with NO token passes through untouched. Deploying this policy alone does not protect anything; pair it with an Authorization Policy that demands `requestPrincipals` for the endpoints that must be authenticated. This is the most common misconfiguration in Istio JWT setups.

**Scope: namespace-wide, selector, or target references -- pick one deliberately** -- with neither a selector nor target references, the policy covers every workload in its namespace (and the whole mesh if created in the Istio root namespace). A workload selector narrows it by pod labels; target references attach it to a specific Gateway, Service, or ServiceEntry. At most one of the two can be set, and waypoint proxies IGNORE label-selector policies -- ambient-mode waypoints require target references.

**One JWT rule per issuer, keys must be reachable** -- each rule names the issuer (matched against the token's `iss` claim), optional audiences, and the JWKS source: OIDC discovery from the issuer URL, a pinned JWKS URL, or inline keys. istiod fetches the keys -- an unreachable JWKS endpoint means tokens from that issuer can never validate, failing requests that carry them.

**Inline keys for a door you own** -- when you mint the tokens yourself (each cluster or service sending into a gateway you run), hold the public keys inline in `jwks` so istiod fetches nothing. An inline key set requires `issuer`. Put one key per sender in the one set, each with its sender's name as key id, and let each token name its sender as subject: the principal reads `<issuer>/<sender>`, the paired ALLOW policy lists principals, and revoking one sender is deleting its key. A token the keys cannot verify is refused with 401, a request with no token reaches the AuthorizationPolicy, which answers 403, and the token is removed before the backend sees it unless `forward_original_token` is set. A token needs no `exp` to validate.

**Namespace is fixed at creation** -- the policy's scope is defined relative to the namespace it is created in; moving it means creating a new policy.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources and operators can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `request_authentication_name` | Name of the created RequestAuthentication (equals `metadata.name`) | Ordering resources that depend on the policy being in place |
| `namespace` | The namespace the policy was created in | Confirming where the policy applies |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Validate JWTs across a namespace** -- accept tokens from a single trusted issuer for every workload in a namespace. The canonical starting point. Start from the **Validate JWTs Across a Namespace** preset.

**Workload JWT validation with claim forwarding** -- validate tokens for one selected workload, read the token from a custom header, restrict it to an audience, and forward a verified claim to the backend as an HTTP header. Start from the **Workload JWT Validation with a Custom Header and Claim Forwarding** preset.

## Works With

- [**Istio Base CRDs**](/cloud-catalog/kubernetes-istio-base-crds) -- the CRDs this policy type registers under (prerequisite).
- [**Istio**](/cloud-catalog/kubernetes-istio) -- the running control plane (istiod) that fetches signing keys and enforces validation in the data plane.
- [**Istio Authorization Policy**](/cloud-catalog/kubernetes-authorization-policy) -- the natural pair: RequestAuthentication establishes WHO the caller is (a verified JWT principal); AuthorizationPolicy decides whether that principal may proceed. Without it, tokenless requests pass.
- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) -- the placement target the policy scopes to.
