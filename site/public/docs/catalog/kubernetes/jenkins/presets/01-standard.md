---
title: "Standard Jenkins"
description: "This preset deploys Jenkins with ingress for external web UI access. Jenkins is an open-source automation server for CI/CD pipelines."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "jenkins"
componentTitle: "Jenkins"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Jenkins

This preset deploys Jenkins with ingress for external web UI access. Jenkins is an open-source automation server for CI/CD pipelines.

## When to Use

- You need a self-hosted CI/CD server on Kubernetes
- You want the Jenkins web UI accessible via a hostname
- Pipeline agents will be dynamically provisioned as Kubernetes pods

## Key Configuration Choices

- **Ingress enabled** -- exposes the Jenkins web UI at the specified hostname
- **Higher resources** (`2000m` CPU, `4Gi` memory limits) -- Jenkins controller is memory-intensive, especially with many plugins and concurrent builds
- **Namespace** (`jenkins`) -- dedicated namespace for Jenkins and its agent pods

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-jenkins.example.com>` | Hostname for the Jenkins web UI | Your DNS provider |
