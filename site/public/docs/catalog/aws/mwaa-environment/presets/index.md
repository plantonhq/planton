---
title: "Presets"
description: "Ready-to-deploy configuration presets for MWAA Environment"
type: "preset-list"
componentSlug: "mwaa-environment"
componentTitle: "MWAA Environment"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-basic-private-airflow"
    rank: "01"
    title: "Preset: Basic Private Airflow Environment"
    excerpt: "A minimal MWAA environment with private webserver access, suitable for development and small-scale DAG workloads within a VPC."
  - slug: "02-production-encrypted-logging"
    rank: "02"
    title: "Preset: Production Encrypted Airflow with Full Logging"
    excerpt: "A production-grade MWAA environment with customer-managed KMS encryption, all five Airflow logging modules enabled, graceful worker replacement, and a defined maintenance window."
  - slug: "03-public-access-with-plugins"
    rank: "03"
    title: "Preset: Public Access with Plugins and Custom Packages"
    excerpt: "An MWAA environment with public webserver access, custom plugins, Python requirements, a startup script, and aggressive worker auto-scaling. Demonstrates the full breadth of MWAA's extensibility..."
---

# MWAA Environment Presets

Ready-to-deploy configuration presets for MWAA Environment. Each preset is a complete manifest you can copy, customize, and deploy.
