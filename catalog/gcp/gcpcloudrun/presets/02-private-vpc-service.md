# Private VPC-Connected Backend

This preset creates an internal backend service wired into a VPC: IAM-authenticated callers only, direct VPC egress to private resources, Cloud SQL over managed Unix sockets, credentials from Secret Manager, and a dedicated least-privilege runtime identity.

## When to Use

- Backend APIs that sit behind the composed HTTPS load balancer or serve other internal services
- Services that reach private-IP resources — Cloud SQL, Memorystore, internal load balancers
- Production workloads where the runtime identity, credential handling, and network posture all need to be explicit

## Key Configuration Choices

- **Dedicated `serviceAccount` by reference** — the identity whose permissions the code exercises; never the over-broad Compute Engine default
- **Direct VPC egress (`networkInterfaces`)** — no Serverless VPC Access connector to size, pay for, or maintain; the subnetwork just needs free addresses
- **`egress: PRIVATE_RANGES_ONLY`** — only private-IP traffic routes through the VPC; public egress keeps Cloud Run's own path
- **Cloud SQL volume + `/cloudsql` mount** — GCP manages the socket proxying; no sidecar and no exposed TCP port
- **`valueFromSecret` env** — the secret name rides in the spec, the material stays in Secret Manager
- **`secretValue` env** — a Planton secret the component keeps in a Secret Manager secret of its own, readable only by the runtime identity; a `$secret/` reference in `value` would land in the revision, so the platform refuses it
- **`ingress: INTERNAL_LOAD_BALANCER`** — the run.app URL stops accepting public traffic; requests come through the VPC or your load balancer
- **One warm instance (`minInstanceCount: 1`)** — no cold starts on the request path

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `backend-runtime` | Your `GcpServiceAccount` resource name | Your service-account manifest |
| `backend-db` | Your `GcpCloudSql` resource name | Your Cloud SQL manifest |
| `backend-db-password` | Secret Manager secret holding the DB password | Secret Manager |
| `payments-api-key` | The Planton secret holding the payments API key | `planton secret list -o json` |
| `my-vpc` / `my-subnet` | Your `GcpVpcNetwork` / `GcpSubnetwork` resource names | Your network manifests |
| `us-docker.pkg.dev/my-project/my-repo/backend:1.0.0` | Your container image | Artifact Registry |

## Related Presets

- **01-public-api-service** — the public-facing starting point
- **03-gpu-inference** — GPU-backed model serving

## Related Components

- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — the database this service connects to
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the runtime identity
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) / [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the network the service egresses into
