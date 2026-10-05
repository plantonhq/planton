# Cloudflare Authenticated Origin Pulls

Deploys a zone's Authenticated Origin Pulls surface: the zone-wide toggle that makes Cloudflare present a client certificate when pulling from your origin, plus per-hostname associations that pin specific hostnames to uploaded client certificates. This is half of a security control — the origin completes it by requiring and validating the certificate, and an enabled zone with a non-validating origin protects nothing. Available on free plans. Destroy turns nothing off: the zone toggle is abandoned at its live value and associations are removed by a revert write, so plan teardown deliberately.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Zone Toggle** — one `cloudflare_authenticated_origin_pulls_settings`, created only when `zoneEnabled` is set. Unset means the toggle is not managed at all; the toggle uses the zone's zone-level client certificate unless associations pin uploaded ones.
- **Hostname Associations** — one `cloudflare_authenticated_origin_pulls` per `hostnameAssociations` row. The provider hard-fails any association resource carrying more than one hostname, so the module fans each row out to its own resource, keyed by hostname.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token carries **Zone → SSL and Certificates → Edit**. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone** for `zoneId` — free plans included.
- **An origin that requires and validates Cloudflare's client certificate** — nginx `ssl_verify_client`, Apache `SSLVerifyClient`, or an ALB mTLS listener. Without the origin-side check, enabling AOP is security theater: traffic that bypasses Cloudflare still reaches the origin unchallenged.
- **Uploaded client certificates** (only for per-hostname pinning) — each `hostnameAssociations[].certificateId` references a hostname-scoped upload from the Cloudflare Authenticated Origin Pulls Certificate component.

## Deploy

### Console

Open the deployment store, find **Cloudflare Authenticated Origin Pulls**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target zone, the zone-wide toggle, and the per-hostname association rows. Start from the **Zone-wide toggle** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAuthenticatedOriginPulls
metadata:
  name: www-origin-pulls
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  zoneEnabled: true
```

```shell
planton apply -f origin-pulls.yaml
```

This enables Authenticated Origin Pulls for the whole zone using the zone-level client certificate — Cloudflare starts presenting it on every origin pull. An Infra Job tracks the provisioning in real time.

### InfraChart

When composing the zone and per-hostname certificates in one InfraPipeline, wire the references with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  zoneEnabled: true
  hostnameAssociations:
    - hostname: app.acme.com
      certificateId:
        valueFrom:
          kind: CloudflareAuthenticatedOriginPullsCertificate
          name: app-client-certificate
          fieldPath: status.outputs.certificate_id
```

The InfraPipeline resolves the dependency graph, deploys the zone and the certificate upload first, then manages the AOP surface with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring Authenticated Origin Pulls. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The origin side is your job** — enabling AOP makes Cloudflare present a client certificate; nothing is protected until the origin requires and validates it. Verify the origin-side configuration before treating the control as live.

**Destroy turns nothing off** — the zone toggle has no delete at Cloudflare: destroying this resource abandons whatever value is live. If AOP should be OFF after teardown, set `zoneEnabled: false` and apply before destroying. An abandoned enabled-toggle on an origin that later stops validating fails silently in the wrong direction — Cloudflare keeps presenting, nobody keeps checking.

**Unset means "leave it alone"** — `zoneEnabled` unset does not mean false; it means the module does not manage the toggle at all, so associations can be managed independently. Explicit `false` asserts OFF.

**Association rows default to active** — an association's `enabled` unset is sent as `true`, because Cloudflare treats a null there as "void the association" and a declared row is meant to exist. Set `enabled: false` for present-but-inactive.

**Every association row pins an uploaded certificate** — `certificateId` is required on each row: Cloudflare rejects an association write without a certificate id (400 code 1404), even when zone-level certificate material exists. Zone-level material rides the zone toggle alone; hostname associations always reference a Cloudflare Authenticated Origin Pulls Certificate upload. Pin per hostname when different origins validate against different certificates.

**Keep hostnames unique** — each row fans out to its own provider resource keyed by hostname; duplicate hostnames collapse in the fan-out.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |
| **CloudflareAuthenticatedOriginPullsCertificate** (required per association row) | `hostnameAssociations[].certificateId` | `status.outputs.certificate_id` |

### What This Kind Provides

This component's `status.outputs` only echoes the managed zone's ID back (`zone_id`) — the AOP surface is zone-singleton shaped, so the zone ID is its identity and there is nothing new for downstream resources to consume. Downstream references belong on the zone or the certificate resources themselves.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Zone-wide enablement** — the first step of locking the origin down to Cloudflare-only traffic: one toggle, the zone-level client certificate, every hostname covered. Start from the **Zone-wide toggle** preset.

**Per-hostname pinning** — associations pinning each hostname to its own uploaded certificate, for zones where different origins validate against different client certificates. Wire `certificateId` from the certificate upload via ValueFromRef.

**Staged disablement** — keep a hostname's row present with `enabled: false` while the origin's validation is being reworked, then flip it back — the association survives without being active.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone whose AOP surface this manages; wire `zoneId` via ValueFromRef
- [**Cloudflare Authenticated Origin Pulls Certificate**](/infra-catalog/cloudflare-authenticated-origin-pulls-certificate) — the uploaded client certificates association rows pin
- [**Cloudflare mTLS Certificate**](/infra-catalog/cloudflare-mtls-certificate) — account-level CA material for validating per-hostname client certificates
