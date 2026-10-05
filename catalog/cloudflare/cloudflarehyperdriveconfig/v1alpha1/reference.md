# CloudflareHyperdriveConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareHyperdriveConfigSpec configures a Cloudflare Hyperdrive: a connection
pooler and global cache that lets a Worker reach a regional SQL database with
low latency. A Worker binds to this config (the `hyperdrive` binding) to query
the origin without paying the full connection-setup round trip on every request.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareHyperdriveConfig
metadata:
  name: test-hyperdrive
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: app-prod-pg
  origin:
    database: app_production
    scheme: postgres
    user: app_user
    host: db.example.com
    port: 5432
    password:
      value: "REPLACE_WITH_DB_PASSWORD"
  caching:
    disabled: false
    maxAge: 60
    staleWhileRevalidate: 15
  originConnectionLimit: 5
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.origin` | `CloudflareHyperdriveOrigin` | yes |  |  |
| `spec.origin.database` | `string` | yes |  |  |
| `spec.origin.scheme` | `enum` | yes |  |  |
| `spec.origin.user` | `string` | yes |  |  |
| `spec.origin.host` | `string` |  |  |  |
| `spec.origin.port` | `int32` |  | `5432` |  |
| `spec.origin.password` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.origin.accessClientId` | `string` |  |  |  |
| `spec.origin.accessClientSecret` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.origin.serviceId` | `string` |  |  |  |
| `spec.caching` | `CloudflareHyperdriveCaching` |  |  |  |
| `spec.caching.disabled` | `bool` |  |  |  |
| `spec.caching.maxAge` | `int64` |  |  |  |
| `spec.caching.staleWhileRevalidate` | `int64` |  |  |  |
| `spec.mtls` | `CloudflareHyperdriveMtls` |  |  |  |
| `spec.mtls.caCertificateId` | `string` |  |  |  |
| `spec.mtls.mtlsCertificateId` | `string` |  |  |  |
| `spec.mtls.sslmode` | `string` |  |  |  |
| `spec.originConnectionLimit` | `int64` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this Hyperdrive config.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string` · required

Human-readable name for the Hyperdrive config (shown in the dashboard and
used as the binding's target).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63"}}

### spec.origin

`CloudflareHyperdriveOrigin` · required

Origin database connection details. Hyperdrive opens and reuses pooled
connections to this origin on behalf of the Worker.

- rule: {"required":true}

### spec.origin.database

`string` · required

Name of the database to connect to (e.g. "app_production").

- rule: {"required":true}

### spec.origin.scheme

`enum` · required

Wire protocol / engine of the origin database.

- rule: scheme must be one of postgres, postgresql, or mysql
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `scheme_unspecified` -- Unspecified scheme (invalid).
- `postgres` -- PostgreSQL over the postgres wire protocol.
- `postgresql` -- PostgreSQL (alias accepted by the API; identical behavior to postgres).
- `mysql` -- MySQL over the mysql wire protocol.

### spec.origin.user

`string` · required

Database user Hyperdrive authenticates as.

- rule: {"required":true}

### spec.origin.host

`string`

Hostname or IP of the origin database. Must be reachable from Cloudflare's
network (a public address, or a private address fronted by Cloudflare Access
/ a Cloudflare Tunnel — see access_client_id/secret).

### spec.origin.port

`int32`

TCP port of the origin database. Leave 0 to use the engine default
(5432 for PostgreSQL, 3306 for MySQL).

- default: `5432`
- rule: port must be 0 (engine default) or between 1 and 65535

### spec.origin.password

`string | valueFrom` · required · sensitive

Password for the database user. Provide a managed-secret reference; the
platform resolves it just-in-time at deploy and never stores it in plaintext.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.origin.accessClientId

`string`

Cloudflare Access client ID, when the origin is published behind Cloudflare
Access (the host is fronted by an Access-protected hostname). Pairs with
access_client_secret.

### spec.origin.accessClientSecret

`string | valueFrom` · sensitive

Cloudflare Access client secret for the service token in access_client_id.
Provide a managed-secret reference; resolved just-in-time at deploy.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.origin.serviceId

`string`

Identifier of the Workers VPC Service to connect through. When set,
Hyperdrive egresses to the origin over that VPC Service (private
connectivity) instead of dialing the public host. Leave empty for a
public/Access-fronted origin. Mutually exclusive with the spec-level mtls
block, since TLS is managed on the VPC Service.

### spec.caching

`CloudflareHyperdriveCaching`

Query-result caching behavior. Caching reduces origin load and latency for
repeated read queries; omit to accept Cloudflare's defaults (enabled, 60s
max age, 15s stale-while-revalidate).

### spec.caching.disabled

`bool`

Disable caching entirely (every query hits the origin). When false, cached
results are served within the freshness window below.

### spec.caching.maxAge

`int64`

Maximum age in seconds a cached query result is served before it is
considered stale. Leave 0 to use Cloudflare's default (60s).

- rule: max_age must be 0 (default) or a positive number of seconds

### spec.caching.staleWhileRevalidate

`int64`

Window in seconds after max_age during which a stale result is served while
the cache refreshes in the background. Leave 0 to use Cloudflare's default (15s).

- rule: stale_while_revalidate must be 0 (default) or a positive number of seconds

### spec.mtls

`CloudflareHyperdriveMtls`

Optional mutual-TLS configuration for connecting to origins that require
client certificates. Omit when the origin does not use mTLS.

### spec.mtls.caCertificateId

`string`

ID of the CA certificate (uploaded to Cloudflare) used to verify the origin's
server certificate.

### spec.mtls.mtlsCertificateId

`string`

ID of the client certificate (uploaded to Cloudflare) Hyperdrive presents to
the origin during the mTLS handshake.

### spec.mtls.sslmode

`string`

TLS verification mode for the origin connection. One of "require" (encrypt,
do not verify), "verify-ca" (verify the CA), or "verify-full" (verify the CA
and the hostname). Leave empty to use the engine default.

- rule: sslmode must be one of "require", "verify-ca", "verify-full"

### spec.originConnectionLimit

`int64`

Maximum number of pooled connections Hyperdrive opens to the origin. Leave 0
to accept the plan default. The Cloudflare minimum is 5; ceilings depend on
plan (20 on the free plan, up to 100 on paid).

- rule: origin_connection_limit must be 0 (plan default) or between 5 and 100

## Validation Rules

- `mtls.not_with_vpc_service`: mtls cannot be combined with a VPC Service origin (origin.service_id); TLS is managed on the VPC Service

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareHyperdriveConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.hyperdrive_id` | `string` | The Cloudflare-assigned identifier of the Hyperdrive config. A Worker's `hyperdrive` binding references this value. |
| `status.outputs.name` | `string` | The Hyperdrive config name (echoed for convenience). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflarePagesProject | `spec.deploymentConfigs.preview.hyperdriveBindings[].configId` | `status.outputs.hyperdrive_id` |
| CloudflarePagesProject | `spec.deploymentConfigs.production.hyperdriveBindings[].configId` | `status.outputs.hyperdrive_id` |
| CloudflareWorker | `spec.hyperdriveConfigs[].configId` | `status.outputs.hyperdrive_id` |

## See Also

- [Overview](../README.md)
