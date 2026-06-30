# Kubernetes Presets: All 51 Components

**Date**: February 14, 2026
**Type**: Feature
**Components**: Presets, Kubernetes Provider

## Summary

Created production-quality presets for all 51 Kubernetes deployment components in Planton, adding 83 presets (166 YAML + MD files, ~3,886 lines). This completes the largest single provider in the presets system, covering operators, infrastructure addons, core workloads, data services, DevOps/CI-CD platforms, observability tools, identity/security solutions, and distributed storage.

## Problem Statement / Motivation

The Kubernetes provider has the most components (51) of any Planton provider, spanning everything from simple operator deployments to complex multi-container platforms like Harbor, Kafka, and Elasticsearch. Users needed ready-made starting points for each component to reduce the learning curve.

### Pain Points

- 51 components with no presets -- users had to read spec.proto files to understand configuration patterns
- Wide variety of component architectures (simple operators, cloud-provider addons with oneof, stateful databases, CI/CD platforms)
- No guidance on which configurations are most common for each component

## Solution / What's New

83 presets organized into 6 batches by architectural pattern:

### Batch 1: Operators (11 components, 13 presets)
- 7 simple operators with identical structure (altinity, elastic, strimzi-kafka, percona-mongo/mysql/postgres, solr) -- 1 preset each
- 4 enhanced operators (rook-ceph-operator, zalando-postgres-operator, tekton-operator, gha-runner-controller) -- 1-2 presets each

### Batch 2: Cloud-Provider Addons (5 components, 14 presets)
- cert-manager: 4 presets (Cloudflare, GCP Cloud DNS, AWS Route53, Azure DNS)
- external-dns: 4 presets (GKE, EKS, AKS, Cloudflare)
- external-secrets: 3 presets (GKE, EKS, AKS)
- ingress-nginx: 2 presets (internet-facing, internal)
- gateway-api-crds: 1 preset

### Batch 3: Core Workloads (10 components, 22 presets)
- namespace (3), deployment (3), statefulset (2), daemonset (2), cronjob (2), job (2), service (3), secret (3), manifest (1), helm-release (1)

### Batch 4: Data Services (9 components, 17 presets)
- postgres (2), redis (2), mongodb (2), nats (2), neo4j (1), clickhouse (1), solr (1), kafka (3), elasticsearch (3)

### Batch 5: DevOps & Observability (9 components, 10 presets)
- argocd, jenkins, gitlab, tekton, gha-runner-scale-set, grafana, prometheus, signoz, locust

### Batch 6: Security, Storage & Platforms (7 components, 11 presets)
- istio, keycloak, openbao (2), openfga, harbor (2), rook-ceph-cluster (2), temporal

## Implementation Details

### Design Decisions

Four design decisions were made before execution:

1. **DD-1: targetCluster omitted** -- treated as a deployment-time concern like org/env/version, not a configuration pattern
2. **DD-2: Provider-specific addon presets** -- cert-manager, external-dns, external-secrets get one preset per cloud provider variant (oneof provider_config)
3. **DD-3: Simple operators get 1 preset** -- resource sizing is a trivial adjustment; a 2nd preset with only different CPU/memory adds noise
4. **DD-4: KubernetesSecret gets 3 type presets** -- opaque, tls, docker-registry cover the vast majority of real-world use cases

### Known Hack Manifest Mismatches Discovered

- `RedisKubernetes` (should be `KubernetesRedis`)
- `PostgresKubernetes` (should be `KubernetesPostgres`)
- `KafkaKubernetes` (should be `KubernetesKafka`)
- `HelmRelease` (should be `KubernetesHelmRelease`)

All presets use the correct kind values from api.proto.

## Benefits

- All 51 Kubernetes components now have at least 1 production-quality preset
- Cloud-provider addons have per-provider presets reducing configuration guesswork
- Consistent camelCase field naming across all K8s presets
- 166 files providing both deployable YAML and explanatory documentation

## Impact

- **Users**: Can now deploy any of the 51 Kubernetes components using a preset as a starting point
- **Contributors**: Clear patterns established for future Kubernetes component presets
- **Platform**: 83 presets brings the total to 224 presets across AWS (49) + GCP (36) + Azure (55) + Kubernetes (83) + 1 pilot

## Related Work

- T01: Presets foundation (convention, rules, Forge integration)
- T02: AWS presets (25 components, 49 presets)
- T03: GCP presets (19 components, 36 presets)
- T04: Azure presets (29 components, 55 presets)
- T06-T08: OpenStack, Scaleway, and remaining providers (pending)

---

**Status**: Production Ready
**Timeline**: Single session (~2 hours)
