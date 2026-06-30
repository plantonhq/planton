# Dev Basic

Minimal development cluster for local development, CI/CD pipelines, and prototyping. Uses 2 CPUs, ZONAL availability, no CMEK, and deletion protection disabled for easy teardown.

## Description

This preset provisions a minimal AlloyDB cluster suitable for development and testing. It keeps costs low with a single-zone deployment and the smallest viable instance size. Deletion protection is disabled so clusters can be destroyed without explicit override.

## Use Case

- Local development and integration testing
- CI/CD pipelines that need a temporary PostgreSQL-compatible database
- Proof-of-concept or prototyping environments
- Cost-sensitive dev/test workloads where high availability is not required

## What This Preset Configures

- **2 CPUs** — Smallest viable primary instance size; GCP selects the machine family automatically
- **ZONAL availability** — Single-zone deployment; lower cost, single zone of failure
- **No CMEK** — Data encrypted with Google-managed keys
- **deletion_protection: false** — Cluster can be destroyed without explicit override
- **Default backups** — GCP applies its default automated backup policy (enabled, daily, 14-day retention) when not specified
- **Default continuous backup** — 14-day PITR window when not specified
- **No initial user** — Access must be configured via AlloyDB Auth Proxy with IAM or manually after provisioning

## When to Use It

Use this preset when you need a quick, low-cost AlloyDB cluster for non-production workloads. Choose a different preset if you require high availability, customer-managed encryption, extended backup retention, or production-grade security controls.
