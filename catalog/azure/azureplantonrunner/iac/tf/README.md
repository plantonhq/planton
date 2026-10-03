# Terraform Module to Deploy AzurePlantonRunner

This module provisions a standing Planton runner appliance on Azure
Container Apps: one single-revision Container App inside the referenced
Container App Environment, pinned to exactly one replica, with the
runner token in the app's own secret store (`runner-token`) and no
ingress at all. The resource group and the environment are referenced
resources -- the module never creates or mutates them (the environment
decides the network boundary; a VNet-integrated one gives the runner
reach into private endpoints).

ENROLLMENT IS TOKEN-FIRST: the app ships the runner TOKEN (in the app's
own secret store), never an identity. The runner joins the control plane
on first boot, registers itself, and receives its own individually
revocable identity; replica replacement re-joins with the same token
(its lineage re-admits the runner it originally admitted).

`variables.tf` reflects the proto schema for `AzurePlantonRunner`.

## Environment contract

The container's environment, exactly as the module wires it:

| Variable | Value |
|----------|-------|
| `PLANTON_RUNNER_TOKEN` | Referenced by name from the app secret `runner-token` -- never a plain env value, so reading the app definition reveals nothing |
| `PLANTON_RUNNER_NAME` | The registration name: `<env>-<metadata.name>`, or `metadata.name` outside an environment |
| `PLANTON_RUNNER_ENDPOINT` | Only when `control_plane_endpoint` is set; omitted, the runner's built-in hosted default applies |
| `PORT` | `50051` (the gRPC/CloudOps server) |
| `LOG_LEVEL` | `info` |

The startup probe reads the runner's health server at `/healthz` on port
8093; it answers independently of control-plane reachability, so a
runner whose control plane is momentarily unreachable still starts.

## The singleton law

The app is pinned to exactly one replica (min = max = 1) in
single-revision mode. A runner's identity is minted for ONE live replica
-- a second replica joining under the same name would revoke the first's
key (token lineage: re-admission re-mints and revokes). Never enable
scaling here without redesigning enrollment for fleets.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

See [`e2e/manifest.yaml`](../../e2e/manifest.yaml) for a minimal test manifest.
