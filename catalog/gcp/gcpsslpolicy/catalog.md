# GCP SSL Policy

Deploys a Compute Engine SSL policy — the control for which TLS protocol versions and cipher suites a load balancer accepts from clients. Without one, GCP's default applies: minimum TLS 1.0 with the permissive COMPATIBLE cipher set. A policy is shared configuration: many proxies reference one policy, and its dials (profile, minimum TLS version, custom suites) update in place — so raising a fleet's TLS floor for PCI DSS is a single-resource change.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine SSL Policy** -- global (blank `region`) for global external Application Load Balancer proxies, or regional for regional external and internal ALB proxies; carries the COMPATIBLE/MODERN/RESTRICTED/FIPS_202205 profile or an explicit CUSTOM cipher-suite allowlist, applied to every referencing proxy's handshakes
- **Compute Engine API enablement** -- `compute.googleapis.com` is enabled in the target project so a fresh project can host the policy; tearing down the policy never disables the API

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the policy will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Compute Engine API itself, so the connection's principal needs permission to enable services (`serviceusage.services.enable`) on a fresh project.

## Deploy

### Console

Open the deployment store, find **GCP SSL Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Modern TLS 1.2 Baseline** preset in the [Presets](#presets) tab to pre-populate the recommended production posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSslPolicy
metadata:
  name: modern-tls12
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  sslPolicyName: "modern-tls12"
  profile: MODERN
  minTlsVersion: TLS_1_2
```

```shell
planton apply -f ssl-policy.yaml
```

This creates the recommended production posture: modern ciphers with a TLS 1.2 floor — the PCI-DSS baseline. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the policy to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the policy — and downstream target proxies reference its `self_link` output.

## Key Configuration

These are the most important decisions when configuring an SSL policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cipher profile** -- COMPATIBLE allows the widest client range (GCP's default); MODERN drops broken ciphers while keeping broad reach — the recommended floor; RESTRICTED narrows to modern-guarantee ciphers for strict compliance; CUSTOM hand-picks exact suites via `customFeatures`; FIPS_202205 pins the FIPS 140-2/3 validated suite set and requires a `TLS_1_2` floor. Untouched records no choice (COMPATIBLE applies). Mutable.

**Minimum TLS version** -- Raise to `TLS_1_2` for PCI DSS and modern compliance regimes; `TLS_1_3` is the strictest floor and requires the RESTRICTED profile. GCP has no maximum-version control — TLS 1.3 is always negotiable when the client supports it, whatever the floor. Untouched records no choice (TLS 1.0 applies). Mutable.

**Post-quantum key exchange** -- A rollout stance for the X25519MLKEM768 hybrid group, not an on/off switch: `DEFAULT` follows GCP's own timeline, `ENABLED` opts in now, `DEFERRED` opts out until GCP's later mandatory date. Mutable.

**Custom cipher suites** -- Required with, and only valid with, the CUSTOM profile: IANA-style suite names from GCP's supported set (unknown names are rejected at deploy). TLS 1.3 suites are never listable — GCP always enables them. A too-narrow list locks out real clients; prefer RESTRICTED unless a compliance regime names exact suites.

**Serving scope** -- Leave `region` empty for a GLOBAL policy; set it for a REGIONAL one. A policy cannot move between scopes or regions. Unusually, even the description is immutable on this resource — a GCP API quirk.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the policy | GcpTargetHttpsProxy `sslPolicy` field via ValueFromRef |
| `ssl_policy_name` | Name as it exists in GCP | Audit, fleet inventory |
| `enabled_features` | The cipher suites GCP actually serves (computed from the profile, or copied from the CUSTOM allowlist) | Compliance audits |
| `region` | Region of a regional policy (empty for global) | Scope verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Modern TLS 1.2 baseline** -- The recommended production posture: broad client reach, no broken ciphers, compliance-ready. Start from the **Modern TLS 1.2 Baseline** preset.

**Restricted strict** -- Only ciphers with modern security guarantees, for regimes stricter than PCI DSS. Start from the **Restricted High-Security Policy** preset.

**Custom allowlist** -- Exact ECDHE AEAD suites for regimes that name ciphers explicitly — demonstrates the CUSTOM↔customFeatures coupling. Start from the **Custom Cipher Allowlist** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the policy is created
- [**GCP Target HTTPS Proxy**](/infra-catalog/gcp-target-https-proxy) -- consumes the policy's `self_link` in its `sslPolicy` field
- [**GCP SSL Certificate**](/infra-catalog/gcp-ssl-certificate) -- the served certificate alongside this policy's negotiation rules
- [**GCP Managed SSL Certificate**](/infra-catalog/gcp-managed-ssl-certificate) -- the Google-issued certificate alternative on the same proxy
