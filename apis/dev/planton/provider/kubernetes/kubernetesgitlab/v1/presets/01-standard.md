# Standard GitLab

This preset deploys a self-hosted GitLab instance on Kubernetes with ingress for external access. GitLab provides a complete DevOps platform including Git repositories, CI/CD, container registry, and issue tracking.

## When to Use

- You need a self-hosted Git platform with integrated CI/CD
- You want the GitLab web UI accessible via a hostname
- Your organization requires on-premises source code management

## Key Configuration Choices

- **Ingress enabled** -- exposes GitLab at the specified hostname
- **Higher resources** (`4000m` CPU, `8Gi` memory limits) -- GitLab is resource-intensive; it bundles Rails, Sidekiq, Gitaly, and PostgreSQL
- **Namespace** (`gitlab`) -- dedicated namespace isolates GitLab components

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gitlab.example.com>` | Hostname for the GitLab web UI | Your DNS provider |
