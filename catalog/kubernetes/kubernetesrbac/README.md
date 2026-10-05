# Kubernetes RBAC

## Overview

**KubernetesRbac** is a catalog kind that models a complete RBAC grant: *"give these subjects these permissions in this scope."* One resource bundles the role definition and its binding — the unit in which RBAC is actually reasoned about — instead of splitting Role, RoleBinding, ClusterRole, and ClusterRoleBinding into four coordinating resources.

## The Grant Model

Three orthogonal choices shape every grant:

1. **Scope** — `namespaceScope` produces Role/RoleBinding objects (permissions inside one namespace); `clusterScope` produces ClusterRole/ClusterRoleBinding objects (cluster-wide permissions, or rules over cluster-scoped resources like nodes and namespaces, or non-resource URLs like `/metrics`). Exactly one must be set.
2. **Role source** — `createRole` defines new permission rules (optionally with ClusterRole aggregation); `existingRole` binds to a role already in the cluster, most commonly the built-in `view`/`edit`/`admin`/`cluster-admin` ClusterRoles. Exactly one must be set.
3. **Subjects** — who receives the permissions: ServiceAccounts (referenced, so charts can create the identity and its grant in one run), users, or groups. When at least one subject is present, a binding is created pointing every subject at the role. When empty, only the role definition is created — useful for publishing a reusable role that other grants bind to later (requires `createRole`; binding an existing role to nobody would deploy nothing).

Any combination of the three axes is a valid grant, and the schema enforces the Kubernetes rules that connect them (aggregation and non-resource URLs need cluster scope; ServiceAccount subjects in cluster scope need an explicit namespace; and so on) at validation time rather than at apply time.

## Kubernetes RBAC Semantics

Understanding three properties of Kubernetes RBAC makes every grant predictable:

- **Deny by default**: a request is allowed only if some Role/ClusterRole rule, reachable through a binding, allows it. No rule → forbidden.
- **Purely additive**: there are no deny rules in Kubernetes RBAC. Grants only ever add permissions; removing access means removing (or narrowing) a grant, never adding a counter-rule. Multiple grants to the same subject union together.
- **Evaluation**: the authorizer checks ClusterRoleBindings first, then RoleBindings in the request's namespace. Rule order within a role is irrelevant — any single matching rule allows the request.

### The built-in roles

Every cluster ships four aggregated, user-facing ClusterRoles, which cover most human-access needs without writing rules:

| Role | Grants | Typical binding |
|------|--------|-----------------|
| `view` | Read-only on most namespaced objects (not Secrets, not RBAC) | Per-namespace, to groups |
| `edit` | Read/write on most namespaced objects (not RBAC) | Per-namespace, to app teams |
| `admin` | `edit` plus namespace-local RBAC management | Per-namespace, to namespace owners |
| `cluster-admin` | Everything, everywhere (`*` on `*`) | Cluster-wide, sparingly |

Bind them with `existingRole: { name: view, isClusterRole: true }` — in namespace scope this creates a RoleBinding to the ClusterRole, which is exactly how built-ins are granted per-namespace.

### Aggregation

A ClusterRole can declare an `aggregationRule` instead of listing rules: the RBAC controller continuously composes its rules from every ClusterRole matching the given label selectors. This is the mechanism behind the built-ins absorbing CRD permissions (operators ship ClusterRoles labeled `rbac.authorization.k8s.io/aggregate-to-view: "true"` and the `view` role grows automatically). Define your own aggregated role to build a plugin-style permission set that teams extend by shipping labeled ClusterRoles. Aggregation exists only on ClusterRoles — namespaced Roles cannot aggregate — and when it is set, directly listed rules are controller-managed and will be overwritten.

## Key value over raw manifests

- **One resource per grant**: the Role+RoleBinding (or ClusterRole+ClusterRoleBinding) pair deploys and is audited as a unit; names stay consistent by construction
- **Kubernetes' cross-object rules validated up front**: scope/aggregation/non-resource-URL/subject-namespace constraints are CEL rules on the spec, so misconfigurations fail at validation instead of half-applying
- **ServiceAccount subjects as references**: a chart can create a KubernetesServiceAccount and its grant in one run, with the name flowing through the graph
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity

## Essential Configuration Fields

### Required

- **Scope**: exactly one of `spec.namespaceScope` (with an optional `namespace` value-or-ref, defaulting to `default`) or `spec.clusterScope` (empty marker)
- **Role source**: exactly one of `spec.createRole` (with `rules` and/or `aggregationRule`, optional `name` defaulting to the component's metadata name) or `spec.existingRole` (with `name` and `isClusterRole`)

### Rules (`createRole.rules[]`)

Each rule independently grants `verbs` (`get`, `list`, `watch`, `create`, `update`, `patch`, `delete`, `deletecollection`, or `*`) over either:

- **Resources**: `apiGroups` (e.g. `""` for core, `apps`, `batch`) + `resources` (lowercase plurals, subresources as `pods/log`) + optional `resourceNames` whitelist (not valid with `create`/`deletecollection` — Kubernetes evaluates those before a name exists), or
- **Non-resource URLs**: paths like `/metrics`, `/healthz`, `/api/*` — cluster scope only, never both in one rule

### Subjects (`spec.subjects[]`)

Each entry is exactly one of:

- **`serviceAccount`**: `name` (value or KubernetesServiceAccount reference) + `namespace` (defaults to the grant's own namespace in namespace scope; required in cluster scope)
- **`user`**: a name as asserted by the cluster's authenticator (OIDC claim, certificate CN, cloud IAM mapping) — Kubernetes has no User objects
- **`group`**: an authenticator-asserted group, e.g. an OIDC groups claim or `system:authenticated`

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`role_name`** / **`role_kind`**: the role in the grant — the created Role/ClusterRole, or the existing role bound to
- **`binding_name`** / **`binding_kind`**: the created RoleBinding/ClusterRoleBinding; empty when the grant defined a role with no subjects
- **`namespace`**: the namespace the grant applies to; empty for cluster-scoped grants

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the scope namespace and any ServiceAccount subject references
2. Create the Role or ClusterRole (when `createRole` is set) with the given rules or aggregation rule
3. Create the RoleBinding or ClusterRoleBinding pointing every subject at the role (when subjects are present)
4. Export the role and binding names and kinds

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesRbac** when you need:

- Least-privilege permissions for a workload's ServiceAccount (created in the same chart)
- Team access via built-in roles (`view`/`edit`/`admin`) bound to OIDC groups per namespace
- Operator-style cluster permissions, including cluster-scoped resources and `/metrics`-style endpoints
- A reusable or aggregated role published without any binding

**Do NOT use** when:

- You need deny semantics — Kubernetes RBAC has none; use admission policy (ValidatingAdmissionPolicy, OPA/Kyverno) for "must not" rules
- The permission target is outside the cluster — cloud IAM is granted on the cloud identity resource, not here

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials with permission to create RBAC objects — note Kubernetes' escalation prevention: you can only grant permissions you yourself hold (or hold the `escalate` verb for)
- **Namespace**: For namespace-scoped grants, the target namespace must exist (or be created in the same chart via a reference)
- **Subjects**: ServiceAccounts should exist or deploy in the same chart; users and groups are plain strings matched at authentication time and need no object

## Best Practices

1. **Grant the least verbs that work**: `get,list,watch` for readers; add write verbs individually. Avoid `*` outside true admin roles
2. **Prefer built-ins for humans**: `view`/`edit`/`admin` bound per-namespace to groups covers most team access with zero maintained rules
3. **Prefer namespace scope**: reach for `clusterScope` only for genuinely cluster-wide needs, cluster-scoped resources, or non-resource URLs
4. **Bind to groups, not users**: membership changes then happen in the identity provider, not in RBAC edits
5. **Treat Secrets access as privileged**: even `view` excludes Secrets; grant `get` on `secrets` deliberately and narrowly (use `resourceNames` where possible)
6. **One grant per intent**: many small KubernetesRbac resources audit and revoke cleanly; mega-roles do neither

## References

- [Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [User-Facing Roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles)
- [Aggregated ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles)
- [Privilege Escalation Prevention](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#privilege-escalation-prevention-and-bootstrapping)
- [RBAC Good Practices](https://kubernetes.io/docs/concepts/security/rbac-good-practices/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
