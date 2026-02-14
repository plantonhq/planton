# Standard Temporal

This preset deploys Temporal -- a durable execution platform for running reliable, long-running workflows. Includes the Temporal server with the web UI exposed via ingress.

## When to Use

- You need durable workflow orchestration (saga patterns, long-running processes, scheduled tasks)
- You want the Temporal web UI for monitoring workflow executions and task queues
- Development or small production Temporal deployments

## Key Configuration Choices

- **Ingress enabled** -- exposes the Temporal web UI at the specified hostname
- **Default resources** -- suitable for moderate workflow volumes; increase for high-throughput environments
- **Self-managed storage** (default) -- Temporal manages its own database; configure external database for production

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-temporal.example.com>` | Hostname for the Temporal web UI | Your DNS provider |
