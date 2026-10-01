# Cloud Run Target

## Use Case

Deploy a delivery pipeline's releases as Cloud Run services into one region of an environment's project.

## When to Use

- A Cloud Run application promoted through staging and production
- A delivery project that deploys into separate per-environment projects

## What This Creates

- The Cloud Deploy API on the delivery project
- A target that deploys Cloud Run services into `acme-staging`'s us-central1 region

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region of the delivery pipeline that deploys here. |
| `run.location` | `projects/acme-staging/locations/us-central1` | The project and region the services run in. |
| `requireApproval` | unset | Set `true` for production. |
