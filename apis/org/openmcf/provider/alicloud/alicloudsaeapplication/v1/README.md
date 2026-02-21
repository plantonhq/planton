# AliCloudSaeApplication

Manages an Alibaba Cloud Serverless App Engine (SAE) application.

## Overview

SAE is a fully managed, container-based serverless compute platform that combines the simplicity of PaaS with the flexibility of containers. Applications can be deployed as container images, JAR packages, WAR packages, or Python/PHP ZIP archives. SAE handles provisioning, scaling, load balancing, and log collection automatically.

This component creates and manages a single `alicloud_sae_application` resource.

## What Gets Created

- **SAE Application** — a container-based serverless application with configurable CPU, memory, replicas, health checks, and deployment strategy

## Supported Package Types

| Type | Description |
|------|-------------|
| `Image` | Container image from ACR or any Docker-compatible registry |
| `FatJar` | Executable JAR file (Java) |
| `War` | Web application archive (Java) |
| `PythonZip` | Python application ZIP archive |
| `PhpZip` | PHP application ZIP archive |

## Configuration Reference

See [catalog-page.md](catalog-page.md) for the full configuration reference, or [examples.md](examples.md) for YAML examples.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `app_id` | The SAE application ID assigned by Alibaba Cloud |
| `app_name` | The application name |

## Related Components

- [AliCloudVpc](../alicloudvpc/v1/) — VPC for network isolation
- [AliCloudVswitch](../alicloudvswitch/v1/) — Subnet for VPC-based deployment
- [AliCloudSecurityGroup](../alicloudsecuritygroup/v1/) — Network access rules
- [AliCloudContainerRegistry](../alicloudcontainerregistry/v1/) — Private image registry
- [AliCloudFunction](../alicloudfunction/v1/) — Alternative serverless compute (event-driven)
