# AWS Lambda Rebuild, Event Source Mapping Forge, and KMS Key Depth Pass

**Date:** 2026-07-04  
**Scope:** Kinds #21–#23 — `AwsLambda`, `AwsLambdaEventSourceMapping` (enum 352), `AwsKmsKey`

## Summary

Three related serverless and encryption kinds brought to the 90/10 bar in one
session: Lambda rebuilt from a never-deployable legacy module to the full
`aws_lambda_function` surface with folded satellites; event source mappings
split out as a first-class composable kind; KMS keys expanded from a six-field
stub to the complete customer-managed key surface with multi-alias support.

## Product impact

- **Lambda manifests change shape** — `function_name` and `codeSourceType` are
  gone; naming is `metadata.name`; code is exactly one of `s3` or `image_uri`.
  Embedded IAM roles, log groups, and invoke permissions are no longer auto-
  created — compose `AwsIamRole`, `AwsCloudwatchLogGroup`, and explicit
  `invoke_permissions` instead.
- **Event-driven pipelines gain a composable edge** — SQS/Kinesis/DynamoDB/
  MSK/MQ/DocumentDB → Lambda wiring is now `AwsLambdaEventSourceMapping`, not
  buried inside the function spec.
- **KMS keys expose the real provider surface** — `key_spec` is a provider
  string (13 values), `policy` is a JSON document, rotation is honest, aliases
  are a repeated list. Thirty-nine downstream consumers of `key_arn` are unchanged.

## Technical highlights

- Registry-driven E2E test-name discovery in `pkg/e2e/profile/discover.go`
  eliminates per-kind PascalCase table maintenance.
- Complete E2E artifacts for all three kinds (verifiers, profiles, scenarios,
  dual-engine test entrypoints); live lanes deferred pending credential refresh.
- `function_arn` output kept stable — 13 FK references across HTTP API Gateway,
  Cognito triggers, and Firehose verified safe.

## Breaking changes

| Kind | Change |
|------|--------|
| AwsLambda | Drop `function_name`; retire embedded role/log-group/permission creation |
| AwsKmsKey | `key_spec` enum → strings; `alias_name` → `aliases[]`; drop `rotation_enabled` output |
| charts/aws/serverless-api | Breaks on new Lambda shape — charts wave, not this session |

## Validation

Offline gate green: spec tests, drift guard, outputs conformance, tofu validate
(all three TF modules), Pulumi bazel builds. Live E2E not run (`InvalidClientTokenId`).

## Post-review fixes (same session)

A full adversarial review of the implementation surfaced and fixed, before any
live lane runs:

- **KMS verifier lifecycle** — destroyed KMS keys sit in `PendingDeletion` for
  the 7–30 day recovery window and stay describable, so the verifier now treats
  `PendingDeletion`/`PendingReplicaDeletion` as absent (verify-absent could
  never pass otherwise).
- **Harness: install-manifest prerequisites** — the dependency resolver now
  expands the `planton.dev/e2e-prerequisites` annotation on each prerequisite's
  OWN install manifest (recursively, cycle-checked), so a dependency whose
  manifest composes sibling fixtures (the zip-backed Lambda referencing the S3
  object-set) deploys after them. Previously the ESM lane deployed the function
  before its S3 chain and could not succeed. Unit-tested.
- **`AwsS3ObjectSet` registry honesty** — gained its required
  `prerequisites: [AwsS3Bucket]` edge (the bucket ref is required).
- **CI matrix regression avoided** — registry-driven test-name discovery
  returned `KubernetesArgocd` where the green ArgoCD lane's entrypoints are
  `TestKubernetesArgoCD_*`; an explicit override map (verified as the only
  enum/test-name deviation repo-wide) plus tests now guard it. The dead
  60-entry hand table was removed.
- **ESM Pulumi schema registry** — pulumi-aws v7.35.0 (the pinned SDK) does
  carry `SchemaRegistryConfig`; the Kafka schema-registry block is now wired on
  both source families, closing a silent TF/Pulumi divergence. The ESM TF
  provider floor rose to `>= 6.16.0` (where `schema_registry_config` landed).
- **Lambda alias CEL honesty** — four new rules reject configurations AWS
  rejects at deploy time: provisioned concurrency on a weighted (canary) alias
  or on `$LATEST`, more than one additional routing version, and weights
  outside 0.0–1.0. Spec tests cover each.
- Cosmetics: gofmt across touched files, empty-vs-omit `description` idiom
  aligned across engines, leftover Pulumi.yaml scaffolding removed.
