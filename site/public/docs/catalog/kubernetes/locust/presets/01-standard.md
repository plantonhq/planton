---
title: "Standard Locust Load Test"
description: "This preset deploys Locust with a master node, 2 worker nodes, and a simple example load test script. Locust is an open-source load testing tool that lets you define user behavior with Python code."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "locust"
componentTitle: "Locust"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Locust Load Test

This preset deploys Locust with a master node, 2 worker nodes, and a simple example load test script. Locust is an open-source load testing tool that lets you define user behavior with Python code.

## When to Use

- Load testing web applications or APIs from within your Kubernetes cluster
- Performance benchmarking before production releases
- Simulating concurrent user traffic patterns

## Key Configuration Choices

- **1 master + 2 workers** -- the master coordinates the test and provides the web UI; workers generate load
- **Ingress enabled** -- exposes the Locust web UI for controlling and monitoring tests
- **Example load test** -- includes a simple Python locustfile that sends GET requests to `/`; replace with your test scenario

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-locust.example.com>` | Hostname for the Locust web UI | Your DNS provider |
