# Kubernetes RBAC - Pulumi Module

## Overview

This Pulumi module deploys one Kubernetes RBAC grant: a role (created or existing) plus, when subjects are present, a binding that points every subject at that role. One resource covers all four Kubernetes RBAC object kinds — Role, ClusterRole, RoleBinding, and ClusterRoleBinding — selected by the grant's scope.

## Architecture

```
iac/pulumi/
├── main.go          # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Make targets for preview/up/down/refresh
└── module/
    ├── main.go      # Orchestrator: provider init, role + binding creation, output export
    ├── locals.go    # Derived values: scope, role name/kind, binding name/kind, labels
    ├── role.go      # Creates rbac.v1.Role or rbac.v1.ClusterRole (with aggregation)
    ├── binding.go   # Creates rbac.v1.RoleBinding or rbac.v1.ClusterRoleBinding
    └── outputs.go   # Exports role_name, role_kind, binding_name, binding_kind, namespace
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesRbacIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` resolves the spec's three orthogonal choices:
   - **Scope**: `namespace_scope` → Role/RoleBinding (namespace defaults to `default`); `cluster_scope` → ClusterRole/ClusterRoleBinding
   - **Role**: `create_role` → a new role named `create_role.name` (default: the component's `metadata.name`); `existing_role` → no role is created, the binding references it by name
   - **Binding**: created only when `subjects` is non-empty, named after the component's `metadata.name`
3. **Role Creation**: `role.go` creates the Role or ClusterRole with the policy rules; ClusterRoles additionally get the aggregation rule when set (namespaced Roles cannot aggregate — the Kubernetes API has no such field)
4. **Binding Creation**: `binding.go` creates the binding with an immutable `roleRef` (apiGroup `rbac.authorization.k8s.io`) and maps subjects:
   - `service_account` → kind `ServiceAccount`; namespace defaults to the grant's namespace in namespace scope (spec validation requires it explicitly in cluster scope)
   - `user` / `group` → kind `User`/`Group` under apiGroup `rbac.authorization.k8s.io`
5. **Output Export**: role/binding names and kinds plus the namespace are exported as outputs

## Grant Shapes

| Spec combination | Objects created |
|------------------|-----------------|
| `namespace_scope` + `create_role` + subjects | Role + RoleBinding |
| `namespace_scope` + `create_role`, no subjects | Role only |
| `namespace_scope` + `existing_role` + subjects | RoleBinding only (roleRef kind per `is_cluster_role`) |
| `cluster_scope` + `create_role` + subjects | ClusterRole + ClusterRoleBinding |
| `cluster_scope` + `create_role`, no subjects | ClusterRole only |
| `cluster_scope` + `existing_role` + subjects | ClusterRoleBinding only (roleRef kind always ClusterRole) |

## Outputs

| Output | Description |
|--------|-------------|
| `role_name` | Name of the role in the grant (created, or the existing role bound to) |
| `role_kind` | `Role` or `ClusterRole` |
| `binding_name` | Name of the created binding; empty when no subjects |
| `binding_kind` | `RoleBinding` or `ClusterRoleBinding`; empty when no binding |
| `namespace` | The grant's namespace; empty for cluster scope |

## Usage

```bash
make preview manifest=path/to/manifest.yaml
make up manifest=path/to/manifest.yaml
make down manifest=path/to/manifest.yaml
```
