# Foreign-Key Reference Annotation Fixes (GKE, AWS subnets, Hetzner, OpenSearch)

**Date**: June 4, 2026
**Type**: Bug Fix + Enhancement
**Components**: API Definitions (foreignkey annotations), GCP / AWS / Hetzner Cloud providers

## Summary

Corrected and completed `default_kind` / `default_kind_field_path` foreign-key
annotations on several spec protos. Every change was surfaced by a new
FK-annotation audit oracle in planton-web (which walks every Spec descriptor and
cross-checks FK annotations) and verified here against the referenced kind's
`stack_outputs.proto`. These annotations are the single source of truth the
cloud-resource wizard reads at runtime to populate cross-resource references, so
a wrong path silently ships a reference to a non-existent output.

## Fixes (wrong / asymmetric annotations)

- **GcpGkeCluster** `network_self_link` / `subnetwork_self_link` declared
  `default_kind_field_path = "status.outputs.self_link"`, but `GcpVpc` /
  `GcpSubnetwork` export `network_self_link` / `subnetwork_self_link` (there is no
  `self_link` output). Corrected to the real outputs.
- **GcpGkeCluster** `cluster_secondary_range_name` / `services_secondary_range_name`
  pointed at `status.outputs.pods_secondary_range_name` /
  `services_secondary_range_name`, which do not exist — `GcpSubnetwork` exposes a
  repeated `secondary_ranges` message. Corrected to
  `status.outputs.secondary_ranges.[*].range_name`.
- **AWS subnet references** with `default_kind = AwsVpc` but **no field path**
  (asymmetric — unusable for reference resolution): `AwsAlb.subnets`,
  `AwsClientVpn.subnets`, `AwsEcsService.network.subnets`, `AwsLambda.subnets`,
  `AwsNetworkLoadBalancer.subnet_mappings.subnet_id`,
  `AwsTransitGateway.vpc_attachments.subnet_ids`. Added
  `status.outputs.private_subnets.[*].id` (AwsVpc exports `private_subnets[].id`).

## Additions (StringValueOrRef FK fields with no annotation)

- **HetznerCloudFloatingIp** `server_id` → `HetznerCloudServer` /
  `status.outputs.server_id` (matches the field's own doc comment + sibling
  HetznerCloudServer references).
- **AwsOpenSearchDomain** `master_user_arn` → `AwsIamRole` /
  `status.outputs.role_arn`.

## Verification

- `buf build` + `buf lint` clean.
- Each `default_kind_field_path` confirmed to exist on the referenced kind's
  `stack_outputs.proto`.

## Known issues flagged (NOT changed here — need a dedicated, verified pass)

- **OCI camelCase output paths**: OCI FK annotations use camelCase (`keyId`,
  `compartmentId`, `subnetId`, `networkSecurityGroupId`) but OCI
  `stack_outputs.proto` fields are snake_case (e.g. `OciKmsKey` → `key_id`). The
  OCI annotations are systematically suspect and need a per-resource verified
  correction.
- **AwsIamRole** lacks an `instance_profile_arn` output referenced by
  `AwsEc2Instance.iam_instance_profile_arn`; **AwsSecretsManager** exports a
  `map secret_arn_map` rather than a scalar `secret_arn` referenced by
  `AwsOpenSearchDomain.master_user_password`. Add the outputs or change the
  references.
- **AzureNatGateway.subnet_id** annotates `AzureVpc/nodes_subnet_id`; the Azure
  convention is `AzureSubnet/subnet_id`.
- **OciDnsRecord/OciDnsZone.view_id** are asymmetric (path, no kind) with no
  `OciDnsView` kind.
- **KubernetesRookCephCluster.operator_namespace** and `GcpCloudRun.dns.managed_zone`
  are plain `string` fields the UI treats as foreign keys — either promote to
  `StringValueOrRef` or drop the UI FK treatment.
