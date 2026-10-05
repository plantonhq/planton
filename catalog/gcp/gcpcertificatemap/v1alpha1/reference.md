# GcpCertificateMap

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCertificateMapSpec defines a Certificate Manager certificate map —
the hostname-to-certificate routing table an external HTTPS load
balancer consults at TLS handshake time: each ENTRY binds a hostname
(or the PRIMARY fallback) to up to fifteen certificates, and the map as
a whole attaches to a GcpTargetHttpsProxy via its certificate_map
argument (this kind's map_uri output is exactly that value). Maps are
how a proxy serves MANY certificates — beyond the ~15-certificate limit
of directly attached lists, and with per-hostname selection (SNI) at
scale.

Certificate maps are a GLOBAL resource (no location — the API pins them
to global), usable by global external ALBs and classic HTTPS proxies.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertificateMap
metadata:
  name: my-sample-certificate-map
spec:
  # What this map routes (shown in the console).
  description: Sample map — one hostname entry bound to the certificate fixture.

  entries:
    # Exactly one of hostname / matcher per entry; 1–15 certificates
    # each. The certificate references the fixture the lanes deploy
    # first.
    - entryName: www
      hostname: www.e2e.example.com
      certificates:
        - valueFrom:
            kind: GcpCertManagerCert
            name: planton-oss-e2e-gcpcert-prereq
            fieldPath: status.outputs.certificate_id

  # What a destroy does: DELETE (detach from proxies first), PREVENT
  # (the posture while proxies reference the map), or ABANDON (keep
  # serving unmanaged).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.mapName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.entries` | `[]GcpCertificateMapEntry` |  |  |  |
| `spec.entries[].entryName` | `string` | yes |  |  |
| `spec.entries[].hostname` | `string` |  |  |  |
| `spec.entries[].matcher` | `string` |  |  |  |
| `spec.entries[].certificates` | `[]string \| valueFrom` | yes |  | GcpCertManagerCert (`status.outputs.certificate_id`) |
| `spec.entries[].description` | `string` |  |  |  |
| `spec.entries[].labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the map in. Can be a literal project ID or
a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.mapName

`string`

The map name in GCP. Defaults to metadata.name when left empty.
Immutable: changing it replaces the map (and every entry — detach the
map from proxies first).

### spec.description

`string`

What this map routes — which domains, which environments. Shown in
the console.

### spec.labels

`map<string, string>`

User labels attached to the map (merged with the platform's standard
labels by the module).

### spec.entries

`[]GcpCertificateMapEntry`

The routing entries. A map with no entries is legal (attach entries
later), but a proxy consulting it will fail every handshake until a
matching entry (or a PRIMARY fallback) exists.

- rule: set exactly one of hostname or matcher

### spec.entries[].entryName

`string` · required

The entry name in GCP (unique within the map). Immutable: changing it
replaces the entry.

- rule: {"required":true}

### spec.entries[].hostname

`string`

The hostname this entry serves: a FQDN (example.com) or a wildcard
expression (*.example.com) matched against the client's SNI. Exactly
one of hostname or matcher. Immutable. The API validates COVERAGE at
entry-create time (live-verified 400: certificate "..." does not
cover map entry hostname "...") — every attached certificate's
domain set must cover this hostname, regardless of the
certificate's provisioning state.

### spec.entries[].matcher

`string`

A predefined matcher instead of a hostname. The API's documented
value is "PRIMARY" — the fallback entry used when no hostname entry
matches the SNI (the value list is API-side and not walled here).
Exactly one of hostname or matcher. Immutable.

### spec.entries[].certificates

`[]string | valueFrom` · required

The certificates presented when this entry matches — 1 to 15 (the
API's per-entry cap). Each is a full certificate resource name
(projects/{project}/locations/{location}/certificates/{name}) — a
literal or a reference to a GcpCertManagerCert resource (its
certificate_id output). Mutable: rotate certificates by editing this
list in place.

- references: GcpCertManagerCert (`status.outputs.certificate_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"15"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.certificate_id}} -- a bare string does not parse

### spec.entries[].description

`string`

What this entry serves. Shown in the console.

### spec.entries[].labels

`map<string, string>`

User labels attached to the entry.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed
(applied to the map and every entry):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- entries and map are deleted; a proxy still referencing
               the map fails TLS handshakes (detach first)
  "PREVENT" -- destroy FAILS; protects live TLS routing
  "ABANDON" -- resources are removed from management but keep serving
               in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCertificateMap, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.map_id` | `string` | The full map resource name (projects/{project}/locations/global/certificateMaps/{name}). |
| `status.outputs.map_uri` | `string` | The map as a GcpTargetHttpsProxy consumes it (//certificatemanager.googleapis.com/projects/{project}/locations/global/certificateMaps/{name}) — set the proxy's certificate_map argument to exactly this value. |
| `status.outputs.map_name` | `string` | The short map name (the last segment of map_id). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.entries[].certificates` | GcpCertManagerCert | `status.outputs.certificate_id` |

## See Also

- [Overview](../README.md)
