# GCP Cloud Run Domain Mapping

Maps a custom domain (like `app.example.com`) directly onto a Cloud Run service — Cloud Run serves the domain itself and provisions/renews the TLS certificate, no load balancer required. The mapping emits the DNS records your domain's zone must publish as outputs, ready to wire into [GcpDnsRecord](/docs/catalog/gcp/gcpdnsrecord) or your external DNS host. This is the scale-appropriate path for "one service, one domain"; the production-grade path for high-traffic or multi-service domains remains the load-balancer composition (serverless NEG → backend service → URL map → HTTPS proxy → forwarding rule).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Domain Mapping** -- a `google_cloud_run_domain_mapping` pointing the verified domain at the Cloud Run service, with a managed TLS certificate in the default `AUTOMATIC` mode
- **Cloud Run API enablement** -- `run.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Domain Verification (one-time, out-of-band)

**The domain must already be verified by the deploying identity** — GCP rejects the mapping otherwise. Verify it once via [Google Search Console](https://search.google.com/search-console/welcome) or `gcloud domains verify <domain>`; subdomains of a verified domain need no separate verification. No Terraform or Pulumi resource performs this step.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Cloud Run service** in the same region and project — reference a [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) resource or name an existing service.
- **IAM**: the deploying identity needs `roles/run.admin` or broader on the project, and must be a verified owner of the domain.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunDomainMapping
metadata:
  name: app-domain
spec:
  region: us-central1
  domain: app.example.com
  route:
    valueFrom:
      kind: GcpCloudRun
      name: my-service
      fieldPath: status.outputs.service_name
```

```shell
planton apply -f mapping.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | GCP region of the Cloud Run service being mapped. Mappings are regional and must live in the SAME region as their target service. | Regional name (e.g. `us-central1`); `global` rejected |
| `domain` | `string` | The custom domain to map — this IS the mapping's name in GCP. Must be verified by the deploying identity. | Lowercase FQDN, max 253 chars |
| `route` | `StringValueOrRef` | The Cloud Run service this domain routes to. Reference a GcpCloudRun resource's `service_name` output or provide the name literally. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference a GcpProject resource. |
| `certificateMode` | `string` | `AUTOMATIC` | `AUTOMATIC` (Cloud Run provisions and renews the certificate) or `NONE` (no managed certificate — for migrations where DNS must be published first). |
| `forceOverride` | `bool` | `false` | Overrides an existing mapping of the same domain without warning. Leave unset for the safe conflict error; set only after that error confirmed the override is intended. |
| `namespace` | `string` | project ID | Cloud Run namespace — GCP requires the project ID or project NUMBER. Set only when a numbered-namespace convention requires the number. |
| `labels` | `map<string,string>` | `{}` | Labels stored on the mapping object (non-authoritative Knative metadata). |
| `annotations` | `map<string,string>` | `{}` | Annotations stored on the mapping object. The Cloud Run API adds server-side annotations of its own — those never show as drift. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (the domain stops routing; the service is untouched), `PREVENT` (destroy fails), or `ABANDON` (keeps serving, dropped from management). |

### Validation Rules

- **The whole mapping is immutable**: every field except `deletionPolicy` is create-only at the provider — any change replaces the mapping. Replacement is cheap (free object, seconds to re-create) with a brief serving gap while the certificate re-issues.
- **Certificate mode is a closed set**: `AUTOMATIC` or `NONE`.
- **The domain must be a lowercase FQDN** with at least two labels (`app.example.com`, not `localhost`).

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `domain` | `string` | The mapped domain (the mapping's name in GCP) |
| `region` | `string` | GCP region the mapping lives in |
| `resource_records` | `[]record` | The DNS records the domain's zone must publish: `record_type` (A/AAAA/CNAME), `record_name` (relative name, CNAME only), `rrdata` (the value). A root domain receives A/AAAA sets; a subdomain one CNAME (`ghs.googlehosted.com.`) |
| `mapped_route_name` | `string` | The Cloud Run route (service) the mapping currently points to |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The mapping exists but does not serve until DNS is published.** After deploy, read the `resource_records` output and create those records in the domain's zone — until then the domain does not resolve and the managed certificate cannot finish issuing.
- **Domain verification is per-identity**: the service account or user running the deploy must be a verified owner. Add robot accounts as additional verified owners in Search Console when deploys run through them.
- **Certificate issuance takes minutes after DNS propagates** — a fresh mapping serving TLS immediately is the exception, not the rule.
- **Deleting the mapping only stops the domain from routing** — the Cloud Run service itself is untouched.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) — the service being mapped; its `service_name` output feeds `route`
- [GcpDnsRecord](/docs/catalog/gcp/gcpdnsrecord) — publishes the `resource_records` output in a Cloud DNS zone
- [GcpDnsZone](/docs/catalog/gcp/gcpdnszone) — the managed zone those records live in
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and API enablement

## Additional Resources

- [Mapping custom domains (Cloud Run)](https://cloud.google.com/run/docs/mapping-custom-domains)
- [Webmaster Central domain verification](https://support.google.com/webmasters/answer/9008080)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
