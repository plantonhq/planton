---
title: "Project"
description: "Project deployment documentation"
icon: "package"
order: 100
componentName: "openstackproject"
---

# OpenStack Project

Deploys an OpenStack Identity (Keystone) project, the fundamental organizational unit in OpenStack that provides resource isolation, quota boundaries, and access control scoping for all cloud resources such as instances, volumes, and networks.

## What Gets Created

When you deploy an OpenStackProject resource, OpenMCF provisions:

- **Keystone Project** — an `openstack_identity_project_v3` resource with the configured description, domain, enabled state, parent hierarchy, and tags

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **Admin role** — creating projects is an admin-level operation; the credentials must have the `admin` role or equivalent permissions in Keystone

## Quick Start

Create a file `project.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: my-project
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackProject.my-project
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackproject/v1/iac/pulumi/module
spec: {}
```

Deploy:

```shell
openmcf apply -f project.yaml
```

This creates a Keystone project named `my-project` in the default domain with enabled state set to `true`.

## Configuration Reference

### Required Fields

All spec fields are optional. The project name is derived from `metadata.name`.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the project. Visible in the OpenStack API, Horizon dashboard, and CLI output. |
| `domainId` | `string` | provider default | Keystone domain to which this project belongs. ForceNew: changing the domain recreates the project. Most single-domain deployments can leave this empty. |
| `enabled` | `bool` | `true` | Whether the project is active. When `false`, all users in the project lose access to its resources, but the resources are not deleted. |
| `parentId` | `string` | — | UUID of the parent project in the project hierarchy. ForceNew: changing the parent recreates the project. Used for nested quota management and organizational structuring. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API. |
| `region` | `string` | provider default | Overrides the region from the provider config. Keystone is typically a global service, so this is rarely needed. |

## Examples

### Basic Tenant Project

A simple project for a development team with default settings:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: dev-team
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackProject.dev-team
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackproject/v1/iac/pulumi/module
spec:
  description: Development team project
```

### Project in a Specific Domain

A project assigned to a custom Keystone domain, useful in multi-domain deployments:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: engineering
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: staging.OpenstackProject.engineering
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackproject/v1/iac/pulumi/module
spec:
  description: Engineering department project
  domainId: abcdef12-3456-7890-abcd-ef1234567890
  tags:
    - engineering
    - staging
```

### Nested Project Hierarchy

A child project under a parent project for organizational structuring and nested quota management:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackProject
metadata:
  name: backend-team
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackProject.backend-team
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackproject/v1/iac/pulumi/module
spec:
  description: Backend team project under engineering
  parentId: 12345678-abcd-ef01-2345-678901abcdef
  tags:
    - backend
    - production
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `project_id` | `string` | UUID of the created Keystone project |
| `name` | `string` | Name of the project, derived from `metadata.name` |
| `domain_id` | `string` | Keystone domain to which this project belongs (computed if not specified) |
| `enabled` | `bool` | Whether the project is currently active |
| `region` | `string` | OpenStack region where the project was created |

## Related Components

- [OpenStack Network](/docs/catalog/openstack/network) — creates Neutron networks within the project
- [OpenStack Security Group](/docs/catalog/openstack/security-group) — defines firewall rules for instances in the project
- [OpenStack Router](/docs/catalog/openstack/router) — provides routing and external connectivity for project networks
- [OpenStack Instance](/docs/catalog/openstack/instance) — launches compute instances within the project
