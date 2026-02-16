---
title: "Standard SigNoz"
description: "This preset deploys SigNoz -- an open-source observability platform providing traces, metrics, and logs in a single UI. Includes the SigNoz frontend/query service and OpenTelemetry Collector with..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "signoz"
componentTitle: "SigNoz"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard SigNoz

This preset deploys SigNoz -- an open-source observability platform providing traces, metrics, and logs in a single UI. Includes the SigNoz frontend/query service and OpenTelemetry Collector with ingress access.

## When to Use

- You need a unified observability platform (traces + metrics + logs) on Kubernetes
- You want an open-source alternative to Datadog or New Relic
- Your applications instrument with OpenTelemetry

## Key Configuration Choices

- **SigNoz + OTel Collector** -- two containers deployed: the SigNoz query service and the OpenTelemetry Collector for data ingestion
- **Ingress enabled** -- exposes the SigNoz web UI at the specified hostname
- **Self-managed ClickHouse** (default) -- SigNoz uses ClickHouse as its storage backend; deployed automatically unless you configure an external database

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-signoz.example.com>` | Hostname for the SigNoz web UI | Your DNS provider |
