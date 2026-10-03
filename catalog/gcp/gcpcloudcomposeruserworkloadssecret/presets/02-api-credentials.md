# API Credentials

Two token entries — an API key and a webhook signing secret — delivered
as one Kubernetes Secret. DAG tasks that call the external service mount
both from a single named Secret.

## When to use

DAGs that integrate with an external API needing more than one piece of
secret material (key + signing secret, client id + client secret,
username + password).

## What to customize

- `environment` — reference your `GcpCloudComposerEnvironment` resource
  (its `environment_name` output).
- `data` entries — base64 of each real token:
  `echo -n 'sk-live-...' | base64`. Add or rename keys freely; the map
  needs at least one entry.
- `secretName` — the name DAGs reference; immutable after creation.

## Security notes

- Values must be base64-encoded; the API rejects raw strings.
- The decoded material never appears in outputs and is held as a
  secret in IaC state.
- Rotating a token is a data update — it applies in place without
  recreating the Secret.

## Composes with

`GcpCloudComposerEnvironment` upstream (reference its
`environment_name` output). Pair with `01-airflow-connection` for
database URIs.
