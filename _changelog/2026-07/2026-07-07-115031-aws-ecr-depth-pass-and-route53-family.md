# AWS ECR Depth Pass + the Route 53 Family: Registry and DNS to 90/10

**Date**: July 7, 2026
**Type**: Feature (breaking, zero users)
**Components**: API Definitions, AWS Provider, Terraform Modules, Pulumi Modules, E2E Harness, Workflow Rules

## Summary

`AwsEcrRepo` is rebuilt from its 8-field spec to the full repository surface —
exclusion-filtered tag immutability, dual-layer encryption, structured
lifecycle rules, a folded repository policy — closing two live Terraform
defects along the way (a hardcoded `scan_on_push` and a silently dropped KMS
key). The Route 53 family comes to the bar in the same pass: the zone drops
its duplicated inline records and gains the full hosted-zone surface (private
VPC associations, DNSSEC, query logging, delegation sets, accelerated
recovery), the DNS record grows to the complete RRType set with all seven
routing policies, and `AwsRoute53HealthCheck` is forged as the availability
probe records reference. All four kinds gain first-ever E2E coverage; all
eight live dual-engine lanes green with a zero-orphan sweep.

## Problem Statement

- **The ECR Terraform module was lying twice.** It hardcoded
  `scan_on_push = true` regardless of the spec, and it read
  `try(var.spec.kms_key_id.value, null)` on a generator-flattened flat
  string — the expression errors, `try()` swallows it, and a
  customer-managed KMS key silently never reached AWS. Every offline gate
  stayed green.
- **The ECR spec was shallow.** A boolean `image_immutable` hid AWS's four
  mutability modes (the exclusion-filtered pair lets release tags freeze
  while `latest` floats); a 2-knob lifecycle "policy" hid the real rule
  model; dual-layer encryption and repository policies were unrepresentable.
- **Route 53 carried a duplicated, worse-shaped record surface.** The zone
  spec folded a `records` list (plain-string alias targets, no
  mutual-exclusion CELs, a divergent enum) while the standalone
  `AwsRoute53DnsRecord` kind modeled the same records properly. The zone's
  Terraform module implemented NEITHER private zones, DNSSEC, nor query
  logging — the spec fields deployed as silent no-ops.
- **The record kind stopped at nine RRTypes and four routing policies**,
  and its `health_check_id` referenced a kind that did not exist.
- **No DNS or registry kind had ever run live.**

## What Was Done

### AwsEcrRepo (breaking rebuild)

Full surface: provider-string `image_tag_mutability` (MUTABLE / IMMUTABLE /
IMMUTABLE_WITH_EXCLUSION / MUTABLE_WITH_EXCLUSION) + wildcard exclusion
filters (max 5, ≤2 wildcards, CEL-coupled to the exclusion modes),
`encryption_type` gains KMS_DSSE (whole block create-time), structured
`lifecycle_rules` (priority uniqueness, single-`any`/single-`untagged`,
tagged-requires-exactly-one-selector — all AWS policy rules as CELs),
folded `repository_policy` Struct. `repository_name` keeps its explicit
field (a slash-namespaced registry path is registry structure, not the graph
node's name) and gains the provider's regex. TF pin `= 5.82.0` →
`>= 6.12.0`; both live TF defects fixed; outputs.tf aligned to the four
proto outputs; the stale VPC-template Pulumi README and stale audit doc
removed.

### AwsRoute53Zone (breaking rebuild)

Inline `records` REMOVED — each record is its own AwsRoute53DnsRecord
composing onto `zone_id` (verified: no chart or consumer used the inline
list). New surface: `comment`, `force_destroy`, `delegation_set_id`
(public-only CEL), `enable_accelerated_recovery`, private-zone
`vpc_associations` (vpc_region defaults to the zone's region), `dnssec`
fold (KSK from a referenced KMS key + signing toggle; us-east-1 /
ECC_NIST_P256 / service-key-policy requirements documented), and
`query_logging` as a CloudWatch log group ARN reference (us-east-1 +
account-level resource-policy prerequisites documented; the account-scoped
policy is deliberately NOT created per zone). The TF module was rewritten
(it ignored half its old spec); the Pulumi module converged off
`pulumi-aws-native` onto classic pulumi-aws — the last AWS module using
aws-native. Outputs: `zone_id` frozen (4 proto consumers + 5 charts),
`primary_name_server` + `zone_arn` added, the dishonest `caller_reference`
(TF hardcoded "", Pulumi never exported it) dropped.

### AwsRoute53DnsRecord (breaking rebuild)

`type` converted from a 9-value nested enum to the full provider RRType
vocabulary as a validated string (A…TLSA, 17 types); routing policies grown
from four to all seven (geoproximity with the exactly-one-location CEL and
bias dial, CIDR by collection/location, multivalue answer with the
no-alias CEL); `allow_overwrite` added; TTL contract completed
(alias-forbids-ttl, values-require-ttl); `health_check_id` promoted to a
`StringValueOrRef` → the new health check kind. Contract enrolled in the
variables.tf drift guard (its metadata block previously carried only
`name`); floor lifted `>= 5.0` → `>= 6.0.0`; registry prerequisites
`[AwsRoute53Zone]`.

### AwsRoute53HealthCheck (forged, enum 354, id_prefix r53hc)

All eight check types with the per-type contract as CELs: endpoint probes
(HTTP/HTTPS/string-match/TCP — target required, TCP has no default port,
search string exactly for the STR_MATCH pair), CALCULATED aggregation over
child health-check references (max 256 + threshold), CLOUDWATCH_METRIC
alarm mirroring (the private-resource pattern), RECOVERY_CONTROL. Probe
tuning (10/30s interval, threshold, ≥3 checker regions, latency, SNI) and
state shaping (invert, administrative disable). `reference_name` skipped
with recorded reason (an idempotency token; the console name is the Name
tag). Full anatomy: 4 protos, both engines, docs, presets, catalog page,
33-case spec suite, E2E.

### E2E (first-ever for all four kinds)

ECR + Route 53 SDK clients added to the harness; four verifiers
(DescribeRepositories; GetHostedZone; ListResourceRecordSets via the
OutputsVerifier path — a record has no standalone describe API, its identity
is zone+name+type; GetHealthCheck); zone install profile as the record
chain's prerequisite; the private-VPC zone scenario composes the shared
AwsVpc fixture via the `e2e-prerequisites` annotation (the registry stays
honest — public zones are leaves); the health-check scenario ships DISABLED
so the checker fleet never probes the placeholder endpoint; eight test
entrypoints; four outputs-conformance cases.

## Validation

- **Offline gate green**: spec tests ×4 (~106 cases), `tofu init &&
  validate` ×4, Pulumi builds ×4, drift guard (record + health check newly
  enrolled), outputs conformance (4 new cases), `validate-refs`,
  `secret-coverage`, `e2e discover` (4 GREEN rows), `make build-go`,
  kind-map + gazelle regenerated, all manifests/scenarios/presets
  CLI-validated, site catalog mirror regenerated.
- **Live: 8/8 dual-engine lanes green** — ECR 1m56s P / 52s T; zone
  public 1m23s P / 2m03s T and private-VPC chain 5m35s P / 7m42s T;
  record chain 2m50s P / 3m22s T; health check 26s P / 1m09s T.
  Zero-orphan sweep clean (Route 53 zones/health checks, ECR, VPC,
  tagged-resource query).
- **Live-caught infrastructure defect fixed at the rules layer**: every
  Terraform lane initially failed with `Unrecognized remote plugin message`
  — not a module bug but a >104-byte unix-socket path from a nested-mktemp
  `$TMPDIR` (macOS caps socket paths; go-plugin sockets live under TMPDIR).
  Pulumi lanes unaffected — that split is the signature. Re-run with a
  short private TMPDIR: all green.

## Key Decisions

- The zone/record split is now fully honest: the zone is the container
  (own lifecycle, one per domain), records are contents (constant churn,
  many per zone, own routing surface). The SNS topic/subscription
  precedent applied to DNS.
- ECR's registry-level resources (registry policy, registry scanning
  config, replication configuration, pull-through cache rules, creation
  templates, account settings) and ECR Public are account-scoped or
  separate product surfaces — deliberately outside the repository kind.
- Route 53's cross-account association plane, delegation-set management,
  traffic policies, CIDR collections, `records_exclusive`, and the
  route53domains/profiles/resolver/recovery services are separate
  lifecycle owners — records/zones compose onto them by ID where needed.

## Workflow Uplift

- `forge-catalog-kind.mdc`: the private E2E `$TMPDIR` must be a SHORT
  path — nested temp dirs overflow macOS's 104-byte unix-socket limit and
  fail every Terraform lane at provider-schema time with the misleading
  `Unrecognized remote plugin message`.
- `update-catalog-kind.mdc`: IaC updates now grep the TF module for
  `.value` reads on generator-flattened fields — `try(x.value, null)` on a
  flat string silently drops the field on every apply and no offline gate
  flags it (the class behind the ECR KMS defect). A sweep of the AWS
  surface found the only other `.value` readers (`awsstepfunction`,
  `awshttpapigateway`) still carry wrapper/`type = any` contracts, where
  the reads are correct — they convert when those kinds reach their
  depth-pass sessions.

## Chart Note

Four charts set the retired `imageImmutable` on ECR
(ecs-environment, microservices-backend, container-app, ml-workbench) —
the charts wave migrates them to `imageTagMutability`. Zone chart
manifests are unaffected (they only pass `region`), and static-website's
standalone record manifest survives unchanged (the enum→string conversion
keeps `type: A` valid).

## Files

Branch `refactor/aws/bring-kinds-to-90-10-coverage-contd-2`: the four
kind trees (`awsecrrepo`, `awsroute53zone`, `awsroute53dnsrecord`,
`awsroute53healthcheck`), `aa_e2e/verify/` (4 new verifiers + registry),
`catalog_kind.proto`/pb.go (enum 354 + record prerequisites),
`e2e/aws/aws_test.go`, `pkg/outputs/conformance_test.go`,
`pkg/iac/tofu/generators/` (drift enrollment + test fixture repoint),
`go.mod`/`go.sum` (ecr + route53 SDK clients), site catalog mirrors, and
the two workflow-rule uplifts.
