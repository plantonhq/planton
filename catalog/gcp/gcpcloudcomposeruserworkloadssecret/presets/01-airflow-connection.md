# Airflow Connection

A single connection URI delivered as a Kubernetes Secret into a Composer
environment — the standard way DAGs reach a database without the URI
living in DAG code or environment variables.

## When to use

Any DAG that talks to a database or service via an Airflow connection.
Reference the Secret from an Airflow connection (secret backend) or
mount it into `KubernetesPodOperator` tasks.

## What to customize

- `environment` — reference your `GcpCloudComposerEnvironment` resource
  (its `environment_name` output).
- `data.connection` — base64 of your real connection URI:
  `echo -n 'postgresql://user:pass@host:5432/db' | base64`.
- `secretName` — the name DAGs reference; immutable after creation.

## Security notes

- Values must be base64-encoded; the API rejects raw strings.
- The decoded material never appears in outputs and is held as a
  secret in IaC state.

## Composes with

`GcpCloudComposerEnvironment` upstream (reference its
`environment_name` output). Pair with `02-api-credentials` for token
material, or `GcpCloudComposerUserWorkloadsConfigMap` for the
non-secret half of a DAG's configuration.
