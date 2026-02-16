---
title: "OpenFGA with PostgreSQL Datastore"
description: "This preset deploys OpenFGA with a PostgreSQL backend for storing authorization models and relationship tuples. OpenFGA is a high-performance authorization engine based on Google's Zanzibar paper."
type: "preset"
rank: "01"
presetSlug: "01-with-postgres-datastore"
componentSlug: "openfga"
componentTitle: "OpenFGA"
provider: "kubernetes"
icon: "package"
order: 1
---

# OpenFGA with PostgreSQL Datastore

This preset deploys OpenFGA with a PostgreSQL backend for storing authorization models and relationship tuples. OpenFGA is a high-performance authorization engine based on Google's Zanzibar paper.

## When to Use

- You need fine-grained authorization (ReBAC, ABAC, or RBAC) for your applications
- You have an existing PostgreSQL database to use as the OpenFGA datastore
- You want the OpenFGA HTTP/gRPC API accessible via ingress

## Key Configuration Choices

- **PostgreSQL datastore** -- production-grade persistent storage for authorization data; alternative: `mysql` or `memory` (testing only)
- **Ingress enabled** -- exposes the OpenFGA API for authorization checks and model management
- **Single replica** -- sufficient for moderate authorization check throughput; scale for higher QPS

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-postgres-host>` | PostgreSQL server hostname or ClusterIP service name | Your PostgreSQL deployment or cloud database |
| `<your-postgres-username>` | PostgreSQL username with access to the `openfga` database | Your database credentials |
| `<your-postgres-password>` | PostgreSQL password | Your database credentials |
| `<your-openfga.example.com>` | Hostname for the OpenFGA API | Your DNS provider |
