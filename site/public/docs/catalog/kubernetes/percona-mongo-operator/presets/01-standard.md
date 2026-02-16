---
title: "Standard Percona Operator for MongoDB"
description: "This preset deploys the Percona Operator for MongoDB with recommended default resources. The operator automates the creation, scaling, and management of Percona Server for MongoDB clusters on..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "percona-mongo-operator"
componentTitle: "Percona Mongo Operator"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Percona Operator for MongoDB

This preset deploys the Percona Operator for MongoDB with recommended default resources. The operator automates the creation, scaling, and management of Percona Server for MongoDB clusters on Kubernetes, including backups, restores, and rolling upgrades.

## When to Use

- You need to run MongoDB on Kubernetes with operator-managed lifecycle
- You want automated backups, point-in-time recovery, and rolling upgrades
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`percona-system`) -- shared namespace for Percona operators; isolates operators from managed database clusters
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`100m` CPU, `256Mi` memory) -- conservative baseline for the operator pod
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- headroom for reconciliation of multiple MongoDB clusters

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
