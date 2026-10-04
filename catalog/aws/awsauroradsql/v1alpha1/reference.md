# AwsAuroraDsql

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsAuroraDsqlSpec defines one Aurora DSQL cluster - AWS's serverless,
PostgreSQL-compatible distributed SQL database with no instances,
no capacity dials, and active-active multi-region writes.

A single-region cluster needs nothing beyond this spec's defaults:
AWS names the cluster with a generated identifier (metadata.name
rides in as the Name tag), scales it to zero when idle, and bills
per request and per byte stored. Connect with standard PostgreSQL
drivers at the derived endpoint
"{identifier}.dsql.{region}.on.aws" (the endpoint output) using IAM
authentication tokens - DSQL has no native database passwords.

Multi-region: two peer clusters in different regions plus a witness
region form one logical database with synchronous active-active
writes. Peering is a one-shot pairing performed while a cluster is
still in PENDING_SETUP - the modules order it correctly, but the
PEER clusters must already exist (deploy one instance of this kind
per region, each naming the others in multi_region).

## Example

```yaml
# Canonical AwsAuroraDsql example (hack/dev manifest and refgen
# Example source): a single-region serverless distributed SQL cluster
# with deletion protection - the production posture. Multi-region
# pairing is the multi_region arm (see the reference page).
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAuroraDsql
metadata:
  name: orders-dsql
  id: orders-dsql
  org: test-org
  env: dev
spec:
  region: us-east-1
  deletionProtectionEnabled: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.deletionProtectionEnabled` | `bool` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.kmsEncryptionKey` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.multiRegion` | `AwsAuroraDsqlMultiRegion` |  |  |  |
| `spec.multiRegion.witnessRegion` | `string` | yes |  |  |
| `spec.multiRegion.peerClusterArns` | `[]string \| valueFrom` | yes |  | AwsAuroraDsql (`status.outputs.cluster_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region the cluster lives in. Example: "us-east-1".
(Aurora DSQL is available in a subset of regions - us-east-1,
us-east-2, us-west-2, eu-west-1/2/3, ap-northeast-1/2,
ap-southeast-1/2 as of early 2026.)

- rule: {"string":{"minLen":"1"}}

### spec.deletionProtectionEnabled

`bool`

Refuse deletes while true. Flip to false (and apply) before
destroying, or set force_destroy and let the module do it.

### spec.forceDestroy

`bool`

Let destroy proceed even when deletion protection is on: the
module disables protection first, then deletes. Keep false in
production - it converts deletion protection from a wall into a
speed bump.

### spec.kmsEncryptionKey

`string | valueFrom`

Encrypt the cluster with your own KMS key instead of the
AWS-owned default key. Reference an AwsKmsKey key_arn output or
pass a literal key ARN. Leave empty for the AWS-owned key (the
provider reports it as the sentinel "AWS_OWNED_KMS_KEY").
Switching keys later re-encrypts in place (no replacement).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.multiRegion

`AwsAuroraDsqlMultiRegion`

Make this cluster one half of a multi-region pair. Omit for a
single-region cluster.

### spec.multiRegion.witnessRegion

`string` · required

The region that arbitrates between the peers during a region
failure (stores transaction logs, runs no queries). Must differ
from every peer's own region. Changing it REPLACES the cluster.

- rule: {"string":{"minLen":"1"}}

### spec.multiRegion.peerClusterArns

`[]string | valueFrom` · required

The ARNs of the peer clusters in the OTHER regions. Reference
other AwsAuroraDsql cluster_arn outputs or pass literal ARNs.
Peering completes only while this cluster is still in
PENDING_SETUP (freshly created) - the pairing is one-shot, and
un-pairing at AWS means recreating the cluster.

- references: AwsAuroraDsql (`status.outputs.cluster_arn`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAuroraDsql, name: <that resource's name>, fieldPath: status.outputs.cluster_arn}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsAuroraDsql, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.identifier` | `string` | The AWS-generated cluster identifier - the provider's import ID. |
| `status.outputs.cluster_arn` | `string` | The cluster's ARN - what a peer cluster's multi_region configuration references. |
| `status.outputs.endpoint` | `string` | The PostgreSQL connection host, "{identifier}.dsql.{region}.on.aws" - the chart-ready join key for application database hosts. (AWS exposes no endpoint attribute; both modules derive this documented DNS shape.) |
| `status.outputs.vpc_endpoint_service_name` | `string` | The VPC endpoint service name for private (PrivateLink) connectivity - create an interface VPC endpoint against it to reach the cluster without public egress. |
| `status.outputs.encryption_type` | `string` | How the cluster is encrypted as AWS reports it (AWS_OWNED_KMS_KEY or CUSTOMER_MANAGED_KMS_KEY). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsEncryptionKey` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.multiRegion.peerClusterArns` | AwsAuroraDsql | `status.outputs.cluster_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAuroraDsql | `spec.multiRegion.peerClusterArns` | `status.outputs.cluster_arn` |

## See Also

- [Overview](../README.md)
