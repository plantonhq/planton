# OciDevopsProject

## Overview

OciDevopsProject is an Planton component that deploys an OCI DevOps project. It provides a single declarative manifest to create the organizational container for CI/CD pipelines, code repositories, deployment environments, artifacts, and triggers.

## Purpose

OCI DevOps is a managed CI/CD platform that builds, tests, and deploys software to OCI infrastructure. The DevOps project is the top-level grouping — all pipelines, repositories, connections, and artifacts live within a project. The project also provides a namespace used in container registry paths and a notification topic for pipeline event delivery. This component provisions the project; downstream DevOps resources (build pipelines, deploy pipelines, repositories) reference the project by its OCID.

## Key Features

- **Pipeline event notifications** — routes build and deployment events to an ONS topic for alerting, Slack integration, or custom automation.
- **Namespace** — the project provides a namespace used in OCIR (OCI Container Registry) paths and artifact references, exported as a stack output.
- **Foreign key references** — `compartmentId` and `notificationTopicId` support `valueFrom` for infra-chart composability.

## Constraints

- Project `name` (from `metadata.name`) is immutable after creation.
- `compartmentId` is updatable (supports compartment moves).
- `notificationTopicId` is updatable.
- Individual DevOps resources (build pipelines, deploy pipelines, repositories, connections) are NOT managed by this component.
- The ONS topic must exist before deploying the project.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Team CI/CD project | Minimal project with compartment + notification topic |
| Multi-team platform | Separate projects per team in isolated compartments |
| Infra-chart composition | Reference compartment via `valueFrom` for automated wiring |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **Namespace output** — the project namespace is exported for use in container registry and artifact paths.
