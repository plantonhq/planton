---
title: "Terraform Parity"
description: "Measured parity of the DigitalOcean catalog against the pinned Terraform provider"
icon: "check-circle"
order: 90
---

<!-- GENERATED FILE -- DO NOT EDIT.
     Rendered from the committed provider schemas, the kind registry, the
     Terraform modules, the per-kind provider-parity manifests, the
     dispositions ledger, and the E2E profiles.
     parameters: provider=digitalocean ga-schema=digitalocean
     Regenerate: make generate-provider-parity-report -->

# DigitalOcean Terraform Parity

This catalog is **built for 100% Terraform parity**: every configurable
argument of the pinned Terraform provider is representable through a kind,
and every provider resource carries exactly one recorded disposition --
omission is a decision, never an accident. This page is the measurement,
generated from the same accounting that gates the repository's CI. It makes
no achieved-parity claim: a kind counts as PROVEN only when live end-to-end
runs pass on both IaC engines, and the tables below show exactly how far
that has progressed.

## Measurement baseline

| | |
|---|---|
| Provider schema (parity baseline) | `digitalocean@2.99.1` |
| Kinds in the catalog | 31 |
| Distinct provider resources consumed | 37 |
| Spec fields authored across all kinds | 675 |
| Module pins on `digitalocean` | `~> 2.99` × 31 |

The GA provider is the parity baseline. Capability that exists only in a
secondary channel (for Google, the `google-beta` provider) enters per kind
through an explicitly enumerated admission list, never wholesale.

## The provider block

Parity covers the provider's own configuration block -- credentials,
role-assumption chains, default tags, endpoint overrides, retry tuning --
under the same total-accounting rule as resources: every configurable,
non-deprecated provider-block argument is matched to a provider-config
field, mapped by recorded judgment, owned by the modules by recorded
judgment, or excluded with a recorded reason -- and arguments set inside
catalog modules' own provider blocks must carry that judgment too.

| Provider-block args | Matched | Mapped | Module-owned | Excluded | Open gaps | Accounted |
|---|---|---|---|---|---|---|
| 9 | 2 | 1 | 0 | 6 | 0 | ✅ |

## Depth: per-kind accounting

Every configurable, non-deprecated provider argument of a kind's consumed
resources must be matched to a spec field, mapped by recorded judgment, or
excluded with a recorded reason -- and every spec field must reach provider
surface. **Accounted** means both directions hold with zero unexplained
gaps. **Proven** means live end-to-end runs passed on both IaC engines.

**31 of 31 kinds are at total accounting; 8 proven live.**

| Kind | Provider args | Matched | Mapped | Excluded | Open gaps | Accounted | Proven |
|---|---|---|---|---|---|---|---|
| DigitalOceanApp | 292 | 0 | 282 | 10 | 0 | ✅ | — |
| DigitalOceanBucket | 28 | 6 | 19 | 3 | 0 | ✅ | — |
| DigitalOceanCdn | 4 | 2 | 2 | 0 | 0 | ✅ | — |
| DigitalOceanCertificate | 6 | 0 | 5 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanContainerRegistry | 6 | 2 | 3 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanDatabaseCluster | 19 | 14 | 5 | 0 | 0 | ✅ | — |
| DigitalOceanDatabaseConnectionPool | 6 | 4 | 2 | 0 | 0 | ✅ | — |
| DigitalOceanDatabaseDb | 2 | 0 | 2 | 0 | 0 | ✅ | — |
| DigitalOceanDatabaseFirewall | 3 | 0 | 2 | 1 | 0 | ✅ | — |
| DigitalOceanDatabaseKafkaSchema | 4 | 3 | 1 | 0 | 0 | ✅ | — |
| DigitalOceanDatabaseKafkaTopic | 27 | 25 | 2 | 0 | 0 | ✅ | — |
| DigitalOceanDatabaseReplica | 7 | 4 | 3 | 0 | 0 | ✅ | — |
| DigitalOceanDatabaseUser | 7 | 1 | 6 | 0 | 0 | ✅ | — |
| DigitalOceanDnsRecord | 10 | 9 | 1 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanDnsZone | 12 | 8 | 3 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanDroplet | 21 | 16 | 4 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanDropletAutoscalePool | 18 | 9 | 8 | 1 | 0 | ✅ | — |
| DigitalOceanFirewall | 17 | 2 | 15 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanFunction | 292 | 0 | 36 | 256 | 0 | ✅ | — |
| DigitalOceanKubernetesCluster | 47 | 31 | 15 | 1 | 0 | ✅ | — |
| DigitalOceanKubernetesNodePool | 13 | 8 | 5 | 0 | 0 | ✅ | — |
| DigitalOceanLoadBalancer | 46 | 31 | 15 | 0 | 0 | ✅ | — |
| DigitalOceanMonitorAlert | 11 | 8 | 3 | 0 | 0 | ✅ | — |
| DigitalOceanProject | 6 | 5 | 1 | 0 | 0 | ✅ | — |
| DigitalOceanReservedIp | 8 | 1 | 3 | 4 | 0 | ✅ | — |
| DigitalOceanSpacesKey | 3 | 0 | 3 | 0 | 0 | ✅ | — |
| DigitalOceanSshKey | 2 | 1 | 1 | 0 | 0 | ✅ | — |
| DigitalOceanUptimeCheck | 14 | 10 | 3 | 1 | 0 | ✅ | — |
| DigitalOceanVolume | 8 | 5 | 3 | 0 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanVpc | 4 | 2 | 1 | 1 | 0 | ✅ | ✅ pulumi, terraform |
| DigitalOceanVpcPeering | 2 | 0 | 2 | 0 | 0 | ✅ | — |

## Breadth: every GA resource, one disposition

All resources of `digitalocean@2.99.1` land in exactly one class:

| Disposition | Resources | Meaning |
|---|---|---|
| Modeled | 37 | consumed by a kind's Terraform module today |
| IAM-covered | 0 | per-resource IAM member/binding/policy triplets, covered by the owning kinds' additive `iam_members` fields |
| Composed | 17 | capability covered through an existing kind's surface rather than a kind of its own |
| Planned | 6 | judged to be covered by a planned kind or planned composition, not built yet |
| Deferred | 17 | deliberately not offered, each with the recorded reason |
| Excluded as deprecated | 2 | deprecated or superseded provider surface |
| **Total** | **79** | |

## The enumerated record

The full per-resource record, so the accounting above is verifiable
rather than trusted.

### Modeled (37)

| Resource | Consuming kinds |
|---|---|
| `digitalocean_app` | consumed by DigitalOceanApp, DigitalOceanFunction |
| `digitalocean_cdn` | consumed by DigitalOceanCdn |
| `digitalocean_certificate` | consumed by DigitalOceanCertificate |
| `digitalocean_container_registry` | consumed by DigitalOceanContainerRegistry |
| `digitalocean_container_registry_docker_credentials` | consumed by DigitalOceanContainerRegistry |
| `digitalocean_database_cluster` | consumed by DigitalOceanDatabaseCluster |
| `digitalocean_database_connection_pool` | consumed by DigitalOceanDatabaseConnectionPool |
| `digitalocean_database_db` | consumed by DigitalOceanDatabaseDb |
| `digitalocean_database_firewall` | consumed by DigitalOceanDatabaseFirewall |
| `digitalocean_database_kafka_schema_registry` | consumed by DigitalOceanDatabaseKafkaSchema |
| `digitalocean_database_kafka_topic` | consumed by DigitalOceanDatabaseKafkaTopic |
| `digitalocean_database_replica` | consumed by DigitalOceanDatabaseReplica |
| `digitalocean_database_user` | consumed by DigitalOceanDatabaseUser |
| `digitalocean_domain` | consumed by DigitalOceanDnsZone |
| `digitalocean_droplet` | consumed by DigitalOceanDroplet |
| `digitalocean_droplet_autoscale` | consumed by DigitalOceanDropletAutoscalePool |
| `digitalocean_firewall` | consumed by DigitalOceanFirewall |
| `digitalocean_kubernetes_cluster` | consumed by DigitalOceanKubernetesCluster |
| `digitalocean_kubernetes_node_pool` | consumed by DigitalOceanKubernetesNodePool |
| `digitalocean_loadbalancer` | consumed by DigitalOceanLoadBalancer |
| `digitalocean_monitor_alert` | consumed by DigitalOceanMonitorAlert |
| `digitalocean_project` | consumed by DigitalOceanProject |
| `digitalocean_record` | consumed by DigitalOceanDnsRecord, DigitalOceanDnsZone |
| `digitalocean_reserved_ip` | consumed by DigitalOceanReservedIp |
| `digitalocean_reserved_ipv6` | consumed by DigitalOceanReservedIp |
| `digitalocean_reserved_ipv6_assignment` | consumed by DigitalOceanReservedIp |
| `digitalocean_spaces_bucket` | consumed by DigitalOceanBucket |
| `digitalocean_spaces_bucket_cors_configuration` | consumed by DigitalOceanBucket |
| `digitalocean_spaces_bucket_logging` | consumed by DigitalOceanBucket |
| `digitalocean_spaces_bucket_policy` | consumed by DigitalOceanBucket |
| `digitalocean_spaces_key` | consumed by DigitalOceanSpacesKey |
| `digitalocean_ssh_key` | consumed by DigitalOceanSshKey |
| `digitalocean_uptime_alert` | consumed by DigitalOceanUptimeCheck |
| `digitalocean_uptime_check` | consumed by DigitalOceanUptimeCheck |
| `digitalocean_volume` | consumed by DigitalOceanVolume |
| `digitalocean_vpc` | consumed by DigitalOceanVpc |
| `digitalocean_vpc_peering` | consumed by DigitalOceanVpcPeering |

### Composed (17)

| Resource | Recorded reason |
|---|---|
| `digitalocean_container_registries` | the plural resource is the newer multi-registry API for the same product surface; DigitalOceanContainerRegistry absorbs it as an implementation upgrade rather than a second registry kind |
| `digitalocean_database_advanced_mysql_config` | extended per-cluster MySQL settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_advanced_postgresql_config` | extended per-cluster PostgreSQL settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_kafka_config` | per-cluster Kafka broker settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_logsink_rsyslog` | same ship-cluster-logs concept as the opensearch logsink, different transport; folds into the planned DigitalOceanDatabaseLogsink kind's oneof destination rather than duplicating the kind |
| `digitalocean_database_mongodb_config` | per-cluster MongoDB settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_mysql_config` | per-cluster MySQL settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_opensearch_config` | per-cluster OpenSearch settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_postgresql_config` | per-cluster PostgreSQL settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_redis_config` | per-cluster Redis/Caching settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_database_valkey_config` | per-cluster Valkey settings singleton -- exactly one per cluster, lifecycle identical to the parent; folds into DigitalOceanDatabaseCluster as per-engine settings |
| `digitalocean_dedicated_inference_token` | access tokens are many-per-endpoint credential rows tightly coupled to the endpoint they authorize; folds into the planned DigitalOceanDedicatedInference kind |
| `digitalocean_nfs_access_point` | export paths and access policies are many-per-share config rows, meaningless without the share; folds into the planned DigitalOceanNfs kind |
| `digitalocean_nfs_attachment` | additional-VPC attachments are membership rows on the share; folds into the planned DigitalOceanNfs kind |
| `digitalocean_project_resources` | membership assignment rows (URN lists) with no meaning outside the project; the DigitalOceanProject kind carries the full membership list on the project resource itself (its resources argument), and this non-authoritative partial-ownership variant conflicts with that whole-list read-back by design -- deliberately not created by the modules |
| `digitalocean_reserved_ip_assignment` | attach/detach is the droplet FK on the DigitalOceanReservedIp kind, which assigns v4 addresses through the reservation's own genuinely mutable droplet_id argument (assign/re-point/unassign all in place) -- this standalone resource is all-ForceNew with a synthetic timestamped id that can never round-trip an import, so it is deliberately never created by the modules |
| `digitalocean_volume_attachment` | attachment is a nullable droplet FK on DigitalOceanVolume (the droplet kind's volume_ids covers it today); attach/detach is a field change, not a separate object lifecycle worth a kind |

### Planned (6)

| Resource | Recorded reason |
|---|---|
| `digitalocean_custom_image` | judged as a planned DigitalOceanCustomImage kind (imported images are first-class and referenced by droplets and autoscale pools) |
| `digitalocean_database_logsink_opensearch` | anchor of a planned DigitalOceanDatabaseLogsink kind modeling ship-cluster-logs-somewhere with a oneof destination config (FK cluster; optionally a Planton-managed OpenSearch cluster) |
| `digitalocean_dedicated_inference` | judged as a planned DigitalOceanDedicatedInference kind (GPU-backed inference endpoints with model deployments and accelerator scaling -- first-class, referenceable infrastructure; FK project) |
| `digitalocean_nfs` | anchor of a planned DigitalOceanNfs kind (the share is the parent object with size, performance tier, and primary VPC; FK vpc) |
| `digitalocean_partner_attachment` | judged as a planned DigitalOceanPartnerAttachment kind (GA but specialist private connectivity via Megaport-class providers with BGP config; first-class because nothing else could honestly absorb it; FK VPCs) |
| `digitalocean_vector_database` | judged as a planned DigitalOceanVectorDatabase kind (Weaviate-powered managed cluster, managed independently from standard database clusters; FK project, vpc) |

### Deferred (17)

| Resource | Recorded reason |
|---|---|
| `digitalocean_byoip_prefix` | bring-your-own-IP needs an out-of-band provisioning workflow (signed request, DigitalOcean-side verification, later advertisement) and serves a small compliance-driven audience -- deferred until the workflow can be represented honestly |
| `digitalocean_database_online_migration` | action-style virtual resource (start/stop a replication job that runs up to two weeks), not durable declarative state -- gated on an action/workflow surface |
| `digitalocean_droplet_snapshot` | a point-in-time capture is operationally an action, not durable declared state -- gated with volume_snapshot on an action/snapshot-policy surface |
| `digitalocean_gradientai_agent` | core object of the GradientAI family, deferred wholesale: the provider surface is visibly immature (copy-paste description bugs in its own schemas, churning release to release) -- revisit as one wave when it settles |
| `digitalocean_gradientai_agent_knowledge_base_attachment` | attachment row of a deferred parent -- the GradientAI family is deferred wholesale on documented provider immaturity |
| `digitalocean_gradientai_agent_route` | agent-to-agent routing rows whose schema carries placeholder-grade descriptions -- the GradientAI family is deferred wholesale on documented provider immaturity |
| `digitalocean_gradientai_custom_model` | depends on the deferred agent surface -- the GradientAI family is deferred wholesale on documented provider immaturity |
| `digitalocean_gradientai_function` | agent function attachment, deferred with its parent -- the GradientAI family is deferred wholesale on documented provider immaturity |
| `digitalocean_gradientai_indexing_job_cancel` | action-style (cancel a running indexing job) -- would be deferred even in a mature family; the GradientAI family is deferred wholesale besides |
| `digitalocean_gradientai_knowledge_base` | likely second anchor when the GradientAI family matures; deferred with the family today on documented provider immaturity |
| `digitalocean_gradientai_knowledge_base_data_source` | data-source rows of a deferred parent -- the GradientAI family is deferred wholesale on documented provider immaturity |
| `digitalocean_gradientai_openai_api_key` | credential row for a deferred family -- the GradientAI family is deferred wholesale on documented provider immaturity |
| `digitalocean_nfs_snapshot` | point-in-time capture, action-style -- same reasoning as droplet and volume snapshots; gated on an action/snapshot-policy surface |
| `digitalocean_spaces_bucket_object` | uploading object content is data-plane work, not infrastructure; a declarative kind for file contents is marginal and invites abuse as a deployment mechanism |
| `digitalocean_tag` | Droplets and volumes create tags implicitly when they declare them, so a tagged fleet needs no standalone kind; but tag CONSUMERS do not -- a firewall naming a tag no Droplet carries yet is rejected by the API (422 'tag does not exist', live-verified), so a tag-first workflow (firewall before its first Droplet) has no handle today and must target a tag some Droplet already carries. Revisit as a re-tier if that ordering constraint proves too costly for charts |
| `digitalocean_volume_snapshot` | point-in-time capture, action-style -- gated with droplet_snapshot on an action/snapshot-policy surface |
| `digitalocean_vpc_nat_gateway` | the provider docs mark it currently in Private Preview -- gated on GA; product value is high, promote to P1 the moment the gate lifts |

### Excluded as deprecated (2)

| Resource | Recorded reason |
|---|---|
| `digitalocean_floating_ip` | docs-deprecated in favor of reserved_ip (the schema DeprecationMessage is commented out in provider code, so the checker cannot compute it) |
| `digitalocean_floating_ip_assignment` | docs-deprecated in favor of reserved_ip_assignment (the schema DeprecationMessage is commented out in provider code, so the checker cannot compute it) |
