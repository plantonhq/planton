# Standard Argo CD

This preset deploys Argo CD with ingress enabled for external access to the web UI. Argo CD is a declarative GitOps continuous delivery tool for Kubernetes.

## When to Use

- You need a GitOps platform to sync Kubernetes resources from Git repositories
- You want a web UI for managing application deployments across clusters
- Standard resource allocation is sufficient for the Argo CD control plane

## Key Configuration Choices

- **Ingress enabled** -- exposes the Argo CD web UI at the specified hostname
- **Namespace** (`argocd`) -- the conventional namespace for Argo CD installations
- **Default resources** -- proto recommended defaults; increase for clusters with many managed applications

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-argocd.example.com>` | Hostname for the Argo CD web UI | Your DNS provider |
