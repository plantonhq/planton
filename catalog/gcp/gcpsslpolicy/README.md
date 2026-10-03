# GCP SSL Policy

Deploys a Compute Engine SSL policy (`google_compute_ssl_policy` / `google_compute_region_ssl_policy`) — the control for which TLS versions and cipher suites a load balancer accepts from clients. Attach its `self_link` to a target HTTPS (or SSL) proxy to enforce modern TLS; without one, GCP's permissive default applies (minimum TLS 1.0, COMPATIBLE ciphers).

## What Gets Created

A single SSL policy. Leave `region` empty for a **global** policy (global external HTTPS load balancers); set it for a **regional** policy (regional external and internal Application Load Balancers). One policy is shared configuration — many proxies can reference it, and tightening it later applies fleet-wide in place.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `ssl-policy.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSslPolicy
metadata:
  name: prod-tls-floor
spec:
  projectId:
    value: my-gcp-project-123
  profile: MODERN
  minTlsVersion: TLS_1_2
  description: TLS 1.2 floor for production frontends
```

Deploy:

```shell
planton apply -f ssl-policy.yaml
```

Reference the policy's `self_link` from a target HTTPS proxy's `sslPolicy` field to harden its client handshakes.

## Configuration Reference

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. Immutable. |
| `sslPolicyName` | `string` | `metadata.name` | Cloud-side name (RFC1035). Immutable. |
| `region` | `string` | `""` (global) | Region for a regional policy; empty means global. Immutable. |
| `description` | `string` | `""` | Why this policy exists. Immutable (a GCP quirk on this resource). |
| `profile` | `string` | `COMPATIBLE` | Cipher profile: `COMPATIBLE`, `MODERN`, `RESTRICTED`, `CUSTOM`, or `FIPS_202205` (requires `minTlsVersion: TLS_1_2`). Mutable. |
| `minTlsVersion` | `string` | `TLS_1_0` | Minimum TLS version: `TLS_1_0`, `TLS_1_1`, `TLS_1_2`, or `TLS_1_3` (requires `profile: RESTRICTED`). Mutable. |
| `customFeatures` | `string[]` | `[]` | Exact cipher suites — required with (and only valid with) `CUSTOM`. Mutable. |
| `postQuantumKeyExchange` | `string` | `DEFAULT` | Post-quantum key exchange (X25519MLKEM768) rollout stance: `DEFAULT`, `ENABLED`, or `DEFERRED`. Mutable. |
| `deletionPolicy` | `string` | `DELETE` | What happens on destroy: `DELETE`, `PREVENT`, or `ABANDON`. |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value a target HTTPS (or SSL) proxy references in `ssl_policy` |
| `ssl_policy_name` | `string` | Name of the SSL policy in GCP |
| `enabled_features` | `string[]` | Cipher suites the policy actually enables, as computed by GCP |
| `region` | `string` | Region of a regional policy; empty for global |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Shared, mutable hardening**: `profile`, `minTlsVersion`, and `customFeatures` update in place and apply to every referencing proxy on the next client handshake — one change hardens a whole fleet.
- **CUSTOM pairing**: the `CUSTOM` profile requires `customFeatures`, and `customFeatures` is rejected on every other profile — validated before deploy.
- **TLS 1.3 is always negotiable**: GCP has no maximum-version control and TLS 1.3 cipher suites are not listable in `customFeatures`; the allowlist governs TLS 1.2 and below. To make TLS 1.3 the *floor*, pair `minTlsVersion: TLS_1_3` with `profile: RESTRICTED`.
- **Compliance pairings validated before deploy**: `FIPS_202205` requires a `TLS_1_2` floor, and a `TLS_1_3` floor requires `RESTRICTED` — both enforced at manifest validation, mirroring what GCP would reject at deploy time.
- **Post-quantum is a rollout stance, not a switch**: `DEFAULT` follows GCP's own X25519MLKEM768 timeline, `ENABLED` opts in now, `DEFERRED` opts out until GCP's later mandatory date.
- **Scope is permanent**: a policy cannot move between global and regional scope, and regional proxies can only reference policies in their own region.

## Related Kinds

- [GcpTargetHttpsProxy](/docs/catalog/gcp/gcptargethttpsproxy) — attaches this policy to harden client handshakes
- [GcpSslCertificate](/docs/catalog/gcp/gcpsslcertificate) — self-managed certificate presented by the same proxy
- [GcpManagedSslCertificate](/docs/catalog/gcp/gcpmanagedsslcertificate) — Google-managed certificate alternative
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the policy

## Additional Resources

- [SSL policies overview](https://cloud.google.com/load-balancing/docs/ssl-policies-concepts)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
