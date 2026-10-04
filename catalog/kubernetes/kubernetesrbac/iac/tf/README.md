# Kubernetes RBAC - Terraform Module

## Overview

This Terraform module deploys one Kubernetes RBAC grant: a role (created or existing) plus, when subjects are present, a binding that points every subject at that role. One module covers all four Kubernetes RBAC object kinds — Role, ClusterRole, RoleBinding, and ClusterRoleBinding — selected by the grant's scope. It is behaviorally identical to the Pulumi module for this kind.

## Architecture

```
iac/tf/
├── provider.tf     # Terraform and Kubernetes provider requirements
├── variables.tf    # Input variables mirroring spec.proto (StringValueOrRef flattened to string)
├── locals.tf       # Resolves scope, role name/kind, binding name/kind, labels, subjects
├── main.tf         # Count-guarded role/cluster-role/binding/cluster-binding resources
├── outputs.tf      # Exports role_name, role_kind, binding_name, binding_kind, namespace
└── README.md       # This file
```

## How It Works

1. **Variable Input**: The `spec` variable takes exactly one scope block (`namespace_scope` / `cluster_scope`), exactly one role source (`create_role` / `existing_role`), and optional `subjects`
2. **Resolution**: `locals.tf` computes the role's name (defaulting to `metadata.name`), the role kind (also used as the binding's `role_ref` kind), the binding name/kind, and normalizes subjects to the Kubernetes Subject shape
3. **Resource Creation**: `main.tf` creates only the objects the grant calls for, via `count` guards
4. **Output Export**: role/binding names and kinds plus the namespace are exported

## Grant Shapes

| Spec combination | Objects created |
|------------------|-----------------|
| `namespace_scope` + `create_role` + subjects | `kubernetes_role_v1` + `kubernetes_role_binding_v1` |
| `namespace_scope` + `create_role`, no subjects | `kubernetes_role_v1` only |
| `namespace_scope` + `existing_role` + subjects | `kubernetes_role_binding_v1` only (role_ref kind per `is_cluster_role`) |
| `cluster_scope` + `create_role` + subjects | `kubernetes_cluster_role_v1` + `kubernetes_cluster_role_binding_v1` |
| `cluster_scope` + `create_role`, no subjects | `kubernetes_cluster_role_v1` only |
| `cluster_scope` + `existing_role` + subjects | `kubernetes_cluster_role_binding_v1` only |

ClusterRoles additionally get `aggregation_rule` when set (namespaced Roles cannot aggregate — the Kubernetes API has no such field).

## Subject Mapping

| Variable block | Kubernetes Subject |
|----------------|--------------------|
| `service_account` | kind `ServiceAccount`; namespace defaults to the grant's namespace (namespace scope) |
| `user` | kind `User`, apiGroup `rbac.authorization.k8s.io` |
| `group` | kind `Group`, apiGroup `rbac.authorization.k8s.io` |

## Usage

```hcl
module "rbac" {
  source = "./iac/tf"

  metadata = {
    name = "ci-deployer"
  }

  spec = {
    namespace_scope = {
      namespace = "production"
    }

    create_role = {
      rules = [{
        verbs      = ["get", "list", "watch", "create", "update", "patch"]
        api_groups = ["apps"]
        resources  = ["deployments"]
      }]
    }

    subjects = [{
      service_account = {
        name = "ci-bot"
      }
    }]
  }
}
```

Granting a built-in ClusterRole in one namespace:

```hcl
  spec = {
    namespace_scope = { namespace = "team-a" }
    existing_role   = { name = "view", is_cluster_role = true }
    subjects        = [{ group = "team-a-readers" }]
  }
```

## Outputs

| Output | Description |
|--------|-------------|
| `role_name` | Name of the role in the grant (created, or the existing role bound to) |
| `role_kind` | `Role` or `ClusterRole` |
| `binding_name` | Name of the created binding; empty when no subjects |
| `binding_kind` | `RoleBinding` or `ClusterRoleBinding`; empty when no binding |
| `namespace` | The grant's namespace; empty for cluster scope |
