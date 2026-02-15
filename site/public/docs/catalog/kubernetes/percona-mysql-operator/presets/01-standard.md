---
title: "Standard Percona Operator for MySQL"
description: "This preset deploys the Percona Operator for MySQL with recommended default resources. The operator automates the creation, scaling, and management of Percona XtraDB Cluster and Percona Server for..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "percona-mysql-operator"
componentTitle: "Percona MySQL Operator"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Percona Operator for MySQL

This preset deploys the Percona Operator for MySQL with recommended default resources. The operator automates the creation, scaling, and management of Percona XtraDB Cluster and Percona Server for MySQL on Kubernetes, including automated backups and self-healing.

## When to Use

- You need to run MySQL on Kubernetes with operator-managed lifecycle
- You want automated provisioning of Percona XtraDB Cluster or Percona Server for MySQL
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`percona-system`) -- shared namespace for Percona operators; isolates operators from managed database clusters
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`100m` CPU, `256Mi` memory) -- conservative baseline for the operator pod
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- headroom for reconciliation of multiple MySQL clusters

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
