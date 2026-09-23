# Personal Runtime

## Use Case

A runtime waiting for one person on the team's standard template, which they start and stop themselves in Colab Enterprise.

## When to Use

- Onboarding a new data scientist
- Workshops and courses: one runtime per participant

## What This Creates

- A runtime in `us-central1` assigned to `alice@example.com` from the `standard-cpu` template, started at creation; the template's idle shutdown stops it when unused

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `runtimeUser` | `alice@example.com` | The person the runtime belongs to. |
| `runtimeTemplate` | `standard-cpu` | A GPU or private template for other roles. |
| `desiredState` | unset | `STOPPED` to park it until needed (enforced on every apply). |
