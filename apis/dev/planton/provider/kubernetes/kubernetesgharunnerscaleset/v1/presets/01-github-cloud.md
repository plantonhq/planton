# GitHub Actions Runner Scale Set (GitHub Cloud)

This preset deploys a self-hosted GitHub Actions runner scale set that connects to GitHub.com. Runners scale from 0 to 10 based on queued workflow jobs, using ephemeral pods that are created per job and destroyed after completion.

## When to Use

- Self-hosted runners for GitHub.com organizations or repositories
- Workloads that need custom tools, larger resources, or private network access not available on GitHub-hosted runners
- Cost optimization by running CI/CD on your own infrastructure

## Key Configuration Choices

- **Scale 0-10** (`minRunners: 0`, `maxRunners: 10`) -- runners scale to zero when idle; adjust max based on expected concurrency
- **Runner resources** (`500m`/`1Gi` requests, `2000m`/`4Gi` limits) -- sized for typical CI builds; increase for heavy compilation or Docker builds
- **Ephemeral runners** -- each job gets a fresh runner pod; no state leaks between jobs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-org>` | GitHub organization (e.g., `my-company`) or `<your-org>/<your-repo>` for repo-level runners | GitHub Settings > Organizations |
| `<your-github-app-secret-name>` | Kubernetes secret containing GitHub App credentials for authentication | Created via `kubectl create secret` with GitHub App private key |
| `<your-runner-label>` | Label used in workflow `runs-on` to target these runners (e.g., `self-hosted-k8s`) | Your workflow YAML files |
