---
title: "Role Assignment"
description: "Role Assignment deployment documentation"
icon: "package"
order: 100
componentName: "openstackroleassignment"
---

# OpenStack Role Assignment

Deploys an OpenStack Identity (Keystone) role assignment, binding a role to a principal (user or group) on a scope (project or domain). This is the fundamental authorization mechanism in OpenStack, controlling what actions a user or group can perform on a given project or domain.

## What Gets Created

When you deploy an OpenStackRoleAssignment resource, OpenMCF provisions:

- **Identity Role Assignment** — an `openstack_identity_role_assignment_v3` resource that binds the specified role to either a user or group on either a project or domain scope

All fields are ForceNew — changing any field causes the assignment to be destroyed and recreated.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **Admin privileges** — role assignments are an admin-level Keystone operation
- **Existing role** — the role UUID must reference a role already created in Keystone (use `openstack role list` to find available roles)
- **Existing principal** — the user or group UUID must reference an existing Keystone user or group
- **Existing scope** — the project or domain must already exist in Keystone

## Quick Start

Create a file `role-assignment.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: my-role-assignment
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackRoleAssignment.my-role-assignment
spec:
  roleId: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  projectId:
    value: "f0e1d2c3-b4a5-6789-0abc-def123456789"
  userId: "11223344-5566-7788-99aa-bbccddeeff00"
```

Deploy:

```shell
openmcf apply -f role-assignment.yaml
```

This assigns the specified role to the user on the given project.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `roleId` | `string` | UUID of the role to assign. Roles are admin-managed Keystone objects (e.g., "admin", "member", "reader"). Use `openstack role list` to find available role UUIDs. ForceNew. |

Additionally, exactly one field from each of the following two pairs must be set:

**Scope (exactly one required):**

| Field | Type | Description |
|-------|------|-------------|
| `projectId` | `StringValueOrRef` | Project to assign the role on. Can reference an OpenStackProject resource via `valueFrom`. Mutually exclusive with `domainId`. ForceNew. |
| `domainId` | `string` | Domain to assign the role on. Mutually exclusive with `projectId`. ForceNew. |

**Principal (exactly one required):**

| Field | Type | Description |
|-------|------|-------------|
| `userId` | `string` | UUID of the user to assign the role to. Mutually exclusive with `groupId`. ForceNew. |
| `groupId` | `string` | UUID of the group to assign the role to. Mutually exclusive with `userId`. ForceNew. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `region` | `string` | provider default | Overrides the region from the provider config for this role assignment. |

### Mutual Exclusion Constraints

This resource enforces two mutual exclusion constraints validated at submission time:

1. **Scope**: exactly one of `projectId` or `domainId` must be set. They define the scope of the role assignment. Setting both or neither is rejected.
2. **Principal**: exactly one of `userId` or `groupId` must be set. They define who receives the role. Setting both or neither is rejected.

## Examples

### User Role on a Project (literal IDs)

Assigns the "member" role to a user on a specific project using literal UUIDs:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: alice-member-on-webteam
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackRoleAssignment.alice-member-on-webteam
spec:
  roleId: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  projectId:
    value: "f0e1d2c3-b4a5-6789-0abc-def123456789"
  userId: "11223344-5566-7788-99aa-bbccddeeff00"
```

### User Role on a Project (referencing OpenStackProject)

Assigns a role to a user on a project managed by an OpenStackProject resource, using `valueFrom` to reference the project ID output:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: bob-admin-on-dataplatform
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackRoleAssignment.bob-admin-on-dataplatform
spec:
  roleId: "deadbeef-1234-5678-9abc-def012345678"
  projectId:
    valueFrom:
      kind: OpenStackProject
      name: data-platform
      fieldPath: status.outputs.project_id
  userId: "aabbccdd-eeff-0011-2233-445566778899"
```

### Group Role on a Domain

Assigns the "reader" role to a group on a domain, granting read access across the entire domain:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: auditors-reader-on-corp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackRoleAssignment.auditors-reader-on-corp
spec:
  roleId: "12345678-abcd-ef01-2345-6789abcdef01"
  domainId: "99887766-5544-3322-1100-ffeeddccbbaa"
  groupId: "abcdef01-2345-6789-abcd-ef0123456789"
```

### Group Role on a Project with Region Override

Assigns a role to a group on a project in a specific region:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackRoleAssignment
metadata:
  name: devs-member-on-staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OpenStackRoleAssignment.devs-member-on-staging
spec:
  roleId: "a1a2a3a4-b1b2-c1c2-d1d2-e1e2e3e4e5e6"
  projectId:
    value: "f1f2f3f4-a1a2-b1b2-c1c2-d1d2d3d4d5d6"
  groupId: "01020304-0506-0708-090a-0b0c0d0e0f10"
  region: RegionOne
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Composite identifier of the role assignment (combination of role, scope, and principal IDs) |
| `role_id` | `string` | UUID of the assigned role |
| `project_id` | `string` | Project scope UUID (empty if domain-scoped) |
| `domain_id` | `string` | Domain scope UUID (empty if project-scoped) |
| `user_id` | `string` | User principal UUID (empty if group assignment) |
| `group_id` | `string` | Group principal UUID (empty if user assignment) |
| `region` | `string` | OpenStack region where the assignment was created |

## Related Components

- [OpenStackProject](/docs/catalog/openstack/openstackproject) — manages the projects that role assignments can scope to
- [OpenStackSecurityGroup](/docs/catalog/openstack/openstacksecuritygroup) — controls network-level access for resources within a project
