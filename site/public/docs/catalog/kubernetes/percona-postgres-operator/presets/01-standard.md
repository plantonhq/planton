---
title: "Standard Percona Operator for PostgreSQL"
description: "This preset deploys the Percona Operator for PostgreSQL with recommended default resources. The operator automates the creation, scaling, and management of PostgreSQL clusters on Kubernetes using..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "percona-postgres-operator"
componentTitle: "Percona Postgres Operator"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Percona Operator for PostgreSQL

This preset deploys the Percona Operator for PostgreSQL with recommended default resources. The operator automates the creation, scaling, and management of PostgreSQL clusters on Kubernetes using Percona Distribution for PostgreSQL, including automated backups, high availability, and connection pooling.

## When to Use

- You need to run PostgreSQL on Kubernetes with operator-managed lifecycle
- You want automated backups, failover, and rolling upgrades for PostgreSQL
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`percona-system`) -- shared namespace for Percona operators; isolates operators from managed database clusters
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`100m` CPU, `256Mi` memory) -- conservative baseline for the operator pod
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- headroom for reconciliation of multiple PostgreSQL clusters

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
