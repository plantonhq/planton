# OpenStackNetwork Kind and Forge Pipeline Cleanup

**Date**: February 9, 2026
**Type**: Feature
**Components**: OpenStack Provider, Forge Pipeline, Catalog Kind Rules

## Summary

Added the `OpenStackNetwork` catalog kind (enum 2501) -- the foundational networking primitive for OpenStack -- and cleaned up the forge pipeline by removing unused E2E test steps and scoping build/test validation to kind directories instead of the entire monorepo.

## Problem Statement / Motivation

The OpenStack provider integration was bootstrapped with a single kind (`OpenStackKeypair`). To enable the `openstack/developer-environment` InfraChart for ARM, the next critical kind is the Neutron network -- the root resource that every other OpenStack networking kind depends on (Subnet, Router, Port, FloatingIp, Instance all reference `network_id`).

Additionally, the forge pipeline had two pain points:
1. **Unused E2E test steps** (rules 011, 014) that required a live OpenStack environment nobody had
2. **Monorepo-wide `make build` and `make test`** in validation steps that compiled the entire codebase and ran all 100+ kind test suites, wasting minutes on every kind operation

### Pain Points

- Building OpenStack InfraCharts requires networking kinds, starting with Network
- `make build` compiled the entire monorepo (Go + frontend + Gazelle) for single-kind changes
- `make test` ran all tests across 100+ catalog kinds when only 1 kind changed
- E2E rules referenced in the forge pipeline could never actually run

## Solution / What's New

### 1. OpenStackNetwork Kind (2501)

Complete catalog kind following the established Keypair pattern:

```mermaid
flowchart TB
    subgraph proto [Proto API Definitions]
        spec["spec.proto\n9 fields, 3 validations"]
        api["api.proto\nKRM envelope"]
        outputs["outputs.proto\nnetwork_id, name, region"]
        input["iac_input.proto\ntarget + provider_config"]
        tests["spec_test.go\n17 tests"]
    end

    subgraph iac [IaC Modules]
        pulumi["Pulumi Go Module\nnetworking.Network"]
        terraform["Terraform HCL Module\nnetworking_network_v2"]
    end

    subgraph docs [Documentation]
        readme["README.md + examples.md"]
        research["docs/README.md\nResearch documentation"]
    end

    spec --> api
    outputs --> api
    api --> input
    spec --> tests
    input --> pulumi
    input --> terraform
```

### 2. Forge Pipeline Cleanup

```mermaid
flowchart LR
    subgraph before [Before: 21 Steps]
        A["..."]
        B["011: Pulumi E2E"]
        C["014: Terraform E2E"]
        D["018: make build\n~2 min"]
        E["019: make test\n~7 min"]
    end

    subgraph after [After: 19 Steps]
        F["..."]
        G["018: go build ./v1/...\n~3 sec"]
        H["019: go test -v ./v1/\n~2 sec"]
    end

    before -->|cleanup| after
```

## Implementation Details

### OpenStackNetwork spec.proto

9 fields selected via 80/20 analysis of the Terraform provider's 19 arguments:

| Field | Type | Design Rationale |
|-------|------|-----------------|
| `description` | `string` | Stored on OpenStack resource, visible in Horizon |
| `admin_state_up` | `optional bool` | Default `true` via middleware; `optional` needed because proto3 bool defaults to false |
| `shared` | `bool` | Admin-only; proto3 default (false) is correct for tenants |
| `external` | `bool` | Admin-only; proto3 default (false) is correct for tenants |
| `mtu` | `int32` | 0 = let OpenStack decide; validated `gte 0` |
| `dns_domain` | `string` | Validated: must end with `.` if set |
| `port_security_enabled` | `optional bool` | No default -- let deployment config decide |
| `tags` | `repeated string` | Validated: unique |
| `region` | `string` | Region override, same pattern as Keypair |

Excluded fields: `tenant_id`, `segments`, `value_specs`, `availability_zone_hints`, `transparent_vlan`, `qos_policy_id` -- all niche or admin-only.

### Forge Pipeline Changes

**Deleted files (4):**
- `_rules/catalog-kind/forge/flow/011-pulumi-e2e.mdc`
- `_rules/catalog-kind/forge/flow/014-terraform-e2e.mdc`
- `_rules/catalog-kind/_scripts/pulumi_e2e_run.py`
- `_rules/catalog-kind/_scripts/terraform_e2e_run.py`
- `.cursor/info/pulumi_e2e.md`

**Command replacements across 16 rule files:**
- `make build` --> `go build ./apis/dev/planton/provider/<provider>/<kind>/v1/...`
- `make test` --> `go test -v ./apis/dev/planton/provider/<provider>/<kind>/v1/`

These replacements were applied to: forge orchestrator, 5 lifecycle rules (update, fix, rename, delete, complete), 7 README files, and 1 authoring guide.

## Benefits

- **OpenStackNetwork** is the root of the OpenStack networking dependency tree -- enables all downstream kinds (Subnet, Router, Port, FloatingIp, Instance)
- **Build validation drops from ~2 minutes to ~3 seconds** (single kind vs entire monorepo)
- **Test validation drops from ~7 minutes to ~2 seconds** (17 tests vs 1000+ tests)
- **Cleaner forge pipeline** -- 19 steps instead of 21, no dead-code E2E rules

## Impact

- **Downstream kinds**: OpenStackSubnet, OpenStackRouter, OpenStackNetworkPort, OpenStackFloatingIp, OpenStackInstance, and OpenStackContainerClusterTemplate all reference `OpenStackNetwork.status.outputs.network_id` as a foreign key
- **All lifecycle rules** (forge, update, fix, rename, delete, complete, audit) now use kind-scoped validation
- **Developer experience**: Every forge or update operation completes validation in seconds, not minutes

## Related Work

- OpenStack provider integration: `_changelog/2026-02/2026-02-08-215116-openstack-provider-integration.md`
- OpenStackKeypair kind: `_changelog/2026-02/2026-02-08-223027-openstackcomputekeypair-catalog-kind.md`
- Parent project: `planton/_projects/20260209.01.openstack-planton-kinds/`

---

**Status**: Production Ready
**Timeline**: Single session (~2 hours)
