# OciFunctionsApplication

## Overview

OciFunctionsApplication is an Planton component that deploys an OCI Functions application. It provides a single declarative manifest to create the organizational container for serverless functions with shared networking, processor architecture, environment configuration, security policies, and observability settings.

## Purpose

OCI Functions is a serverless compute platform that runs code in response to events or HTTP requests. The Functions application is the top-level grouping that provides shared execution context — subnets, NSGs, processor architecture, environment variables, and policies — for all functions deployed within it. This component provisions the application infrastructure; individual functions are deployed as code artifacts via `fn deploy` or CI/CD.

## Key Features

- **Processor architecture selection** — x86 (Intel/AMD), ARM (Ampere A1), or multi-architecture.
- **Subnet placement** — functions execute in the specified subnets and can access resources reachable from them.
- **Application config** — shared environment variables available to all functions in the application.
- **Image signature verification** — enforce that only container images signed by specified KMS keys can be deployed.
- **APM tracing** — distributed tracing integration with OCI Application Performance Monitoring.
- **NSG binding** — optional network security groups for fine-grained network access control.
- **Syslog forwarding** — optional syslog URL for centralized log collection.
- **Foreign key references** — `compartmentId`, `subnetIds`, `networkSecurityGroupIds`, and image policy `kmsKeyId` support `valueFrom`.

## Constraints

- `displayName`, `subnetIds`, and `shape` are immutable after creation.
- `keyDetails` must be non-empty when `imagePolicyConfig.isPolicyEnabled` is `true`.
- Application `config` keys must be ASCII letters, digits, and underscores; cannot start with a digit. Max total 4 KB.
- Individual functions are NOT managed by this component — they are deployed separately.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Quick serverless prototyping | Minimal app with one subnet, default x86 |
| Cost-optimized ARM functions | `shape: generic_arm` for Ampere A1 |
| Multi-arch deployment pipeline | `shape: generic_x86_arm` for flexible CI/CD |
| Secure image supply chain | Image signature verification via KMS keys |
| Observability-first functions | APM tracing enabled with domain ID |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **Image signature verification** — prevents deployment of unsigned or tampered container images.
- **APM tracing** — distributed tracing for function invocations across the OCI observability stack.
