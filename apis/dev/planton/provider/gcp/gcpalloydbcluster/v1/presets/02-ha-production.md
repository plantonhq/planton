# HA Production

Production-ready AlloyDB cluster with high availability, automated backups with 7-day retention, initial user, deletion protection, and ENCRYPTED_ONLY SSL. Uses 4 CPUs and REGIONAL deployment for automatic failover.

## Description

This preset provisions a production-ready AlloyDB cluster suitable for application databases requiring high availability and automated backups. It uses a regional deployment for automatic failover across zones and enforces TLS-only client connections.

## Use Case

- Production application databases requiring high availability
- Workloads that need automated backups with a defined retention window
- Environments where accidental deletion must be prevented
- Applications that require encrypted client connections (TLS only)
- Standard production PostgreSQL workloads with moderate compute needs

## What This Preset Configures

- **4 CPUs** — Moderate production capacity; GCP selects the machine family
- **REGIONAL availability** — Multi-zone deployment with automatic failover
- **deletion_protection: true** — Prevents accidental cluster destruction
- **initial_user** — Creates a `postgres` superuser with the specified password for direct database access
- **automated_backup_policy** — Enabled with 7-day time-based retention (604800 seconds)
- **ssl_mode: ENCRYPTED_ONLY** — All client connections must use TLS
- **Default continuous backup** — 14-day PITR window when not specified
- **No CMEK** — Data encrypted with Google-managed keys

## When to Use It

Use this preset for production workloads that need high availability and automated backups but do not require customer-managed encryption. For enterprise compliance or CMEK requirements, use the enterprise-encrypted preset.
