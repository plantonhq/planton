# Istio Authorization Policy

Defines an Istio AuthorizationPolicy: a namespaced resource that enforces request-level access control on your mesh workloads. It decides whether a request is allowed, denied, or audited based on who the request is from, what it is doing, and additional conditions -- or hands the decision to an external authorizer. The spec is 100% faithful to the upstream `security.istio.io/v1` AuthorizationPolicy (pinned to the Istio 1.30 line), flattened directly after the Planton namespace envelope -- this is how you express zero-trust access rules inside the mesh.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **An AuthorizationPolicy** -- a namespaced Istio policy that applies an action (Allow, Deny, Audit, or Custom) to the requests matched by its rules, scoped to selected workloads or to the whole namespace.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Istio installed** -- the Istio CRDs (**Istio Base CRDs**) must be present and the Istio control plane (istiod) running. The policy is only enforced where istiod is active.
- **Target namespace exists** -- the policy is created in a specific namespace; reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Istio Authorization Policy**, and click **Deploy**. The creation wizard walks you through the namespace, the scope, the action, and the matching rules, with guidance at each step.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAuthorizationPolicy
metadata:
  name: allow-frontend-to-api
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  selector:
    matchLabels:
      app: api
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
              - cluster.local/ns/prod-apps/sa/frontend
      to:
        - operation:
            methods: ["GET", "POST"]
            paths: ["/api/*"]
```

```shell
planton apply -f authorization-policy.yaml
```

This allows only the `frontend` service account to call `GET`/`POST` on `/api/*` of the `api` workloads in `prod-apps`; everything else to those workloads is denied. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace to a resource managed by another Cloud Resource:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: prod-apps-namespace
      fieldPath: spec.name
  action: ALLOW
```

The InfraPipeline creates the namespace first, then registers the policy in it. The workload selector is a plain label match resolved by istiod at runtime, not a foreign key -- to order the policy after the workload it protects, express the dependency through `metadata.relationships`.

## Key Configuration

These are the most important decisions when configuring an authorization policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Empty rules are a verdict, not a no-op.** The action and the rules work together, and the empty cases are consequential: with **ALLOW**, once any allow policy applies to a workload, only requests that match a rule are permitted -- so an ALLOW policy with **no rules denies all traffic** to the selected workloads. With **DENY**, a policy with no rules does nothing.

**Evaluation order is fixed and Deny always wins.** Across all policies on a workload, CUSTOM is evaluated first, then DENY, then ALLOW -- a Deny overrides an Allow no matter the order they were created in. Design deny rules as guardrails and allow rules as grants; they compose predictably because of this ordering.

**Scope: selector XOR target references.** At most one of a workload `selector` (pod labels) and `targetRefs` (Gateway, Service, ServiceEntry -- up to 16) may be set; both omitted means namespace-wide, or mesh-wide when the policy lives in the Istio root namespace. Waypoint proxies IGNORE label-selector policies -- attaching to a waypoint requires `targetRefs`.

**Rule matching is any-source, any-operation, ALL conditions.** A request matches the policy if it matches any rule; within a rule it must match at least one `from` source, at least one `to` operation, and every `when` condition. A rule that should match every request says `matchAll: true` (the empty rule on the Istio wire) -- with DENY that is a total lockout of the selected workloads; a rule that names neither a matcher nor `matchAll` is refused, so matching everything is always a written choice.

**CUSTOM needs a provider that already exists.** The CUSTOM action delegates to an extension provider named in the mesh's MeshConfig (`provider.name`). istiod enforces the CUSTOM-provider coupling at runtime, not at admission -- a typo in the provider name surfaces as failing requests, not a rejected manifest.

**Identity-based rules need authentication in place.** Rules matching `principals` require mTLS (compose an **Istio Peer Authentication**); rules matching `requestPrincipals` or JWT-claim conditions require an **Istio Request Authentication** that validates the token. Without the authentication layer, identity fields never populate and identity rules never match.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `authorization_policy_name` | Name of the created AuthorizationPolicy (equals `metadata.name`) | Ordering resources that depend on the policy being in place |
| `namespace` | The namespace the policy was created in | Confirming where the policy applies |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Require a valid JWT principal** -- allow only requests that carry a validated request principal, locking down a workload to authenticated callers. Pair with a Request Authentication that validates the token. Start from the **Require an Authenticated JWT Principal** preset.

**Front a workload with external authorization** -- use the CUSTOM action to delegate the decision to an external authorizer (for example, an ext-authz service) on ingress traffic. Start from the **Delegate to an External Authorizer (CUSTOM) on the Ingress Gateway** preset.

## Works With

- [**Istio Base CRDs**](/cloud-catalog/kubernetes-istio-base-crds) -- the CRD prerequisite; the policy kind must exist before this resource can be applied.
- [**Istio**](/cloud-catalog/kubernetes-istio) -- the control plane that enforces the policy; nothing is enforced where istiod is not active.
- [**Istio Request Authentication**](/cloud-catalog/kubernetes-request-authentication) -- validates JWTs so rules can match `requestPrincipals` and JWT-claim conditions.
- [**Istio Peer Authentication**](/cloud-catalog/kubernetes-peer-authentication) -- enforces mTLS so rules can match peer `principals` and `namespaces`.
- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) -- the placement target the policy is registered in.
