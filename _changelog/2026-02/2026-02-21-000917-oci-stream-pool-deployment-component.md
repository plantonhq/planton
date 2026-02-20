# OCI Stream Pool Deployment Component

**Date**: February 21, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Terraform Module, OCI Provider

## Summary

Implemented the OciStreamPool deployment component -- OCI's Kafka-compatible managed streaming service bundling a stream pool with its constituent streams under shared Kafka settings, KMS encryption, and optional private networking. First resource of Phase 8 (Messaging and Streaming). Also executed a naming audit renaming 4 pending resources (R30-R33) to drop unnecessary service-name prefixes.

## Problem Statement / Motivation

OCI Streaming provides a fully managed, Kafka-compatible event-streaming service used for real-time data ingestion, log aggregation, and event-driven architectures. Stream pools are the organizational container that groups streams with shared configuration -- Kafka compatibility settings, encryption keys, and networking. Without this component, OpenMCF users would need to manage OCI Streaming resources outside the platform.

### Pain Points

- No managed way to provision OCI Streaming infrastructure through OpenMCF
- Kafka-compatible streaming is a prerequisite for the planned OCI Data Platform infra chart
- Stream pools and streams are tightly coupled (streams inherit pool settings) but must be provisioned separately through the provider

## Solution / What's New

A complete deployment component following the container+sub-resource pattern established by OciNosqlTable (table + indexes):

- `oci_streaming_stream_pool` as the primary resource
- `oci_streaming_stream` as bundled sub-resources, keyed by name

### Key Design Decisions

- **Flat `kmsKeyId` field** instead of nested `customEncryptionKey` block: The provider's block contains only one user-facing input (`kms_key_id`); `key_state` is computed. Flattening gives cleaner YAML.
- **Streams as repeated sub-resource**: Each stream has `name` (ForceNew, map key), `partitions` (ForceNew, gte 1), and `retentionInHours` (ForceNew, optional 24-168). Identical pattern to NoSQL table indexes.
- **KafkaSettings with optional scalars**: `auto_create_topics_enable`, `log_retention_hours`, `num_partitions` all use proto3 `optional` for explicit presence (nil = OCI defaults). `bootstrap_servers` is a computed output.
- **PrivateEndpointSettings entirely ForceNew**: `subnet_id` (StringValueOrRef), `nsg_ids`, `private_endpoint_ip`.

### Naming Audit

Renamed 4 pending resources to drop redundant service-name prefixes:

| Old Name | New Name | Rationale |
|----------|----------|-----------|
| OciStreamingStreamPool | OciStreamPool | "Streaming" is service prefix |
| OciQueueQueue | OciQueue | Redundant double "Queue" |
| OciMonitoringAlarm | OciAlarm | "Monitoring" is service prefix |
| OciLoggingLogGroup | OciLogGroup | "Logging" is service prefix |

4 resources kept as-is: OciDnsZone, OciDnsRecord (domain concepts), OciNetworkFirewall (product name), OciDevopsProject ("Project" too generic alone).

## Implementation Details

### Proto API (4 files)

- `spec.proto`: OciStreamPoolSpec with 5 fields, 3 nested messages (KafkaSettings, PrivateEndpointSettings, Stream), 2 CEL validation rules (log_retention_hours range, retention_in_hours range)
- `api.proto`: OciStreamPool top-level message with `oci.openmcf.org/v1` api-version
- `stack_input.proto`: OciStreamPoolStackInput (target + provider_config)
- `stack_outputs.proto`: 3 outputs (stream_pool_id, endpoint_fqdn, kafka_bootstrap_servers)

### Validation Tests

24 Ginkgo/Gomega tests (12 valid, 12 invalid scenarios), all passing. Coverage includes:
- Minimal valid, full config, Kafka settings, KMS key, private endpoint, streams, retention range validation
- Invalid: wrong api_version/kind, missing required fields, out-of-range retention, missing private endpoint subnet

### Pulumi Module (5 Go files)

- `stream_pool.go`: Creates `streaming.NewStreamPool()` with conditional KafkaSettings, CustomEncryptionKey (wraps flat kmsKeyId), PrivateEndpointSettings. Iterates `spec.Streams` to create `streaming.NewStream()` for each. Extracts `bootstrap_servers` from KafkaSettings output via `ApplyT`.
- `locals.go`: FreeformTags from metadata labels
- `outputs.go`: 3 output key constants

### Terraform Module (6 HCL files)

- `main.tf`: Stream pool with 3 dynamic blocks (kafka_settings, custom_encryption_key, private_endpoint_settings)
- `streams.tf`: `oci_streaming_stream.this` with `for_each` keyed by stream name
- `outputs.tf`: 3 outputs using `try()` for safe kafka_bootstrap_servers extraction

### Kind Registration

OciStreamPool=3370 under new "Messaging and Streaming" section in `cloud_resource_kind.proto`. `kind_map_gen.go` regenerated.

## Benefits

- OpenMCF users can now provision OCI Streaming infrastructure declaratively
- Kafka-compatible bootstrap servers exposed as output for direct Kafka client connectivity
- Private endpoint support enables secure, VCN-scoped streaming
- Streams bundled as sub-resources eliminate the need for separate stream management
- Naming audit improves consistency and reduces cognitive overhead across remaining resources

## Impact

- **Users**: Can deploy OCI Streaming pools with Kafka settings, encryption, and streams via a single YAML manifest
- **Infra Charts**: Unblocks the OCI Data Platform chart which depends on OciStreamPool for data ingestion
- **Downstream Agents**: R30 Start=done enables Docs and Presets agents to begin work on this component

## Related Work

- Follows container+sub-resource pattern from R20 OciNosqlTable (table + indexes)
- Enables Phase 8 (Messaging and Streaming) -- R31 OciQueue is next
- Required by Chart 5: OCI Data Platform (data ingestion via streaming)

---

**Status**: Production Ready
**Validation**: go build clean, go vet clean, 24/24 tests passed, terraform validate success
