---
title: "DevOps Project"
description: "DevOps Project deployment documentation"
icon: "package"
order: 100
componentName: "ocidevopsproject"
---

# OCI DevOps Project

Deploys an Oracle Cloud Infrastructure DevOps project — the organizational container for CI/CD pipelines, code repositories, deployment environments, artifacts, and triggers. The project provides a shared namespace and an ONS notification topic for pipeline event delivery.

## What Gets Created

When you deploy an OciDevopsProject resource, OpenMCF provisions:

- **DevOps Project** — a `devops.Project` resource in the specified compartment with a notification topic for pipeline events (build completions, deployment successes, failures). The project name is derived from `metadata.name`.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the project will be created — either a literal value or a reference to an OciCompartment resource
- **An ONS topic OCID** for receiving DevOps pipeline events — the topic must already exist in OCI Notifications

## Quick Start

Create a file `devops-project.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDevopsProject
metadata:
  name: my-project
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDevopsProject.my-project
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  notificationTopicId:
    value: "ocid1.onstopic.oc1..example"
```

Deploy:

```shell
openmcf apply -f devops-project.yaml
```

This creates a DevOps project in the specified compartment with pipeline events routed to the ONS topic. The project OCID and namespace are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the project will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `notificationTopicId` | `StringValueOrRef` | OCID of the ONS topic for pipeline event notifications (build started, deployment succeeded, etc.). Can reference an ONS topic via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the project's purpose. |

## Examples

### Minimal Project

A DevOps project with direct OCID values:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDevopsProject
metadata:
  name: my-project
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDevopsProject.my-project
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  notificationTopicId:
    value: "ocid1.onstopic.oc1..example"
```

### Project with Compartment Reference

A project referencing an OciCompartment for composability in infra charts:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDevopsProject
metadata:
  name: platform-cicd
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDevopsProject.platform-cicd
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: cicd-compartment
      fieldPath: status.outputs.compartmentId
  notificationTopicId:
    value: "ocid1.onstopic.oc1..example"
  description: "Platform team CI/CD pipelines for production workloads"
```

### Full-Featured with Description

A production project with a descriptive purpose:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDevopsProject
metadata:
  name: backend-services
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: backend
    pulumi.openmcf.org/stack.name: prod.OciDevopsProject.backend-services
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  notificationTopicId:
    valueFrom:
      kind: OciCompartment
      name: notifications-topic
      fieldPath: status.outputs.compartmentId
  description: "Backend microservices build and deployment pipelines"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `project_id` | `string` | OCID of the DevOps project |
| `namespace` | `string` | Namespace associated with the project, used in container registry paths and artifact references |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciContainerEngineCluster](/docs/catalog/oci/container-engine-cluster) — OKE clusters are common deployment targets for DevOps pipelines
- [OciFunctionsApplication](/docs/catalog/oci/functions-application) — Functions applications are deployment targets for serverless pipelines
