# AWS CI/CD Pair: CodeBuild Project and CodePipeline at the Full Provider Surface

**Date**: July 9, 2026
**Type**: Feature
**Components**: API Definitions, Terraform Modules, Pulumi Modules, E2E Harness

## Summary

The AWS CI/CD pair reaches the full provider surface. **`AwsCodeBuildProject`** grows from its 15-field spec to the complete project model — secondary sources/artifacts/version pins, source auth (CodeConnections/Secrets Manager/OAuth), commit-status configuration, the persistent Docker server, reserved-fleet membership, EFS mounts, batch builds, public visibility, build badges, automatic retry, a folded resource policy, and the completed webhook fold (org/group scope, manual creation, fork-PR comment-approval gating). **`AwsCodePipeline`** gains the entire V2 stage-condition surface — entry gates, post-success checks, and failure handling with automatic rollback/retry, each backed by AWS's managed rules — plus an honest trigger contract. Two live cross-engine defects are fixed, both Terraform contracts move to generator-owned `variables.tf` under the drift guard, and both kinds ship first-ever E2E artifacts with all four live dual-engine lanes green.

## What Was Built

### `AwsCodeBuildProject` — the complete build-project surface

- **Secondary sources, artifacts, and version pins** (12 each, identifier-addressed): multi-repository builds ($CODEBUILD_SRC_DIR_<id> checkouts) and multi-destination outputs, with identifier requiredness enforced by CEL (primary must NOT carry one; every secondary MUST).
- **Source depth**: `auth` (CODECONNECTIONS as the modern token-less path; SECRETS_MANAGER; legacy OAUTH), `build_status_config` (custom commit-status context + target URL), `insecure_ssl`, `git_submodules_config` — with source-type applicability rules (submodules, status reporting) promoted to CELs covering primary AND secondary sources.
- **Environment depth**: the full 12-value environment type set (EC2 types, `MAC_ARM`, `WINDOWS_CONTAINER`), `ATTRIBUTE_BASED_COMPUTE` + `CUSTOM_INSTANCE_TYPE` compute, `certificate` (private-CA bundles), `docker_server` (persistent dedicated Docker daemon — layer state survives across builds), and `fleet_arn` (reserved-capacity fleet membership by reference).
- **Project level**: `auto_retry_limit`, `badge_enabled` (+ `badge_url` output), `project_visibility` + `resource_access_role` (PUBLIC_READ coupling as CEL), `build_batch_config` (batch fan-out with compute/build-count restrictions), `file_system_locations` (EFS build caches), and the folded `resource_policy` (cross-account project sharing — one document keyed by the project ARN).
- **Webhook fold completed**: the five missing filter types (WORKFLOW_NAME/TAG_NAME/RELEASE_NAME/REPOSITORY_NAME/ORGANIZATION_NAME), `manual_creation` (AWS mints the payload URL + HMAC secret without registering — the operator wires the repository by hand from the new `webhook_payload_url` + `webhook_secret` outputs), `scope_configuration` (organization/group/global webhooks for runner projects), `pull_request_build_policy` (comment-approval gating for fork PRs — the open-source CI safety posture), and the `RUNNER_BUILDKITE_BUILD` build type.
- **Honesty CELs**: Lambda environment types reject build/queued timeouts and privileged mode (AWS ignores/forbids them); registry credentials require SERVICE_ROLE pulls; badges require a repository source.
- Exclusion recorded: `host_kernel` merged into the provider two days ago (post-v6.53.0, unreleased) and is absent from the pinned pulumi-aws — deliberately not modeled until both engines can deploy it.

### `AwsCodePipeline` — the complete V2 orchestration surface

- **Stage conditions**: `before_entry` (entry gates — deployment windows, alarm checks), `on_success` (post-deploy verification that can fail a green stage), and `on_failure` with `result` ROLLBACK/RETRY/FAIL, `retry_configuration` (FAILED_ACTIONS/ALL_ACTIONS), and optional rule gating. Rules carry the full contract: managed `rule_type_id` (DeploymentWindow/CloudWatchAlarm/LambdaInvoke/VariableCheck/Commands), configuration, shell commands, input artifacts, region, role ref, timeout.
- **Trigger contract honesty**: PR `events` validated to the provider enum (`OPEN`/`UPDATED`/`CLOSED` — the old comment said UPDATE, and nothing was validated), at-least-one-filter CELs on git configuration and on every filter block, pattern length bounds.
- **Action timeout honesty**: AWS accepts `timeout_in_minutes` overrides ONLY on Manual Approval actions and rejects everything else at CreatePipeline — now a spec CEL, failing at validation instead of apply.
- V2-requiredness CELs extended to stage conditions; single-vs-cross-region artifact-store shape promoted to a CEL.

### Cross-engine defects fixed (both live)

- **CodePipeline Terraform silently dropped `timeout_in_minutes`** — declared in the contract, never wired into the action block; Pulumi honored it. A user's approval timeout applied on one engine and not the other.
- **CodeBuild webhook outputs diverged**: Terraform always exported `webhook_url`/`webhook_payload_url` (empty when absent), Pulumi exported them only when a webhook existed. Converged on the always-exported shape in both engines.

### Both kinds to the settled engineering bar

- Generator-owned `variables.tf` contracts (the legacy hand-written `{value=string}` FK-wrapper shapes retired), enrolled in the drift guard; `outputs.tf` created for both (outputs previously lived inline in main.tf).
- Terraform floors lifted from `~> 5.0` to `>= 6.16.0` (CodeBuild — `auto_retry_limit` lands there; docker_server 6.2.0, pull_request_build_policy 6.13.0 beneath it) and `>= 6.0.0` (CodePipeline — the full V2 surface predates 6.0).
- Naming basis converged to `metadata.name` in both engines (was `metadata.id`); identity tags converged from the non-standard `planton.org/*` prefix to the shared `planton.ai/*` set.
- `iac/pulumi/iac-input.yaml` created for both; hack manifests rebuilt to exercise the FULL surface — including every arm excluded from live lanes — so the offline plan proofs cover the whole contract.
- Outputs enriched: CodeBuild adds `badge_url`, `public_project_alias`, and `webhook_secret` (sensitive — a provider-minted credential, required for manual webhook creation; marked `sensitive` in the Terraform output and bridged as a secret by Pulumi).
- Richly-commented modules in both engines; zero PARITY-EXCEPTIONs.

## E2E

- Two new verifiers: CodeBuild `BatchGetProjects` (no typed NotFound — existence decided by whether the name resolves) and CodePipeline `GetPipeline` (typed `PipelineNotFoundException`); codebuild + codepipeline SDK modules added.
- Registry prerequisites: CodeBuild `[AwsIamRole]`; CodePipeline `[AwsIamRole, AwsS3Bucket]`. The shared IAM fixture grew CodeBuild-service and CodePipeline role documents; a CodeBuild fixture project serves the pipeline chain.
- Scenarios: CodeBuild full-surface (NO_SOURCE + inline buildspec — standalone, no repository or connection needed) and the CodePipeline s3-codebuild-chain (a two-stage V2 pipeline composed against the versioned bucket fixture, the pipeline role, and the CodeBuild fixture project via the e2e-prerequisites annotation; QUEUED execution, a pipeline variable, a namespaced action, ROLLBACK failure handling).
- **Live dual-engine E2E 4/4 green** (2026-07-09): CodeBuild Pulumi 15m59s total (first lane — 14m12s was IAM-fixture warm-up incl. fresh plugin download; project deploy 14s) / Terraform 5m20s (deploy 10s); CodePipeline chain Pulumi 5m47s / Terraform 5m58s (pipeline deploy 13s/10s). Zero-orphan sweep clean across projects, pipelines, e2e roles, and buckets.
- Live-lane exclusions recorded in the profiles with reasons: webhook + source auth (need a real source-provider connection — the manual OAuth handshake class), badge/public visibility, VPC/EFS/batch/docker-server pass-through arms, resource policy (cross-account principal validation), triggers, cross-region stores, condition rules (act only during executions) — all proven by spec tests and the offline plan gate.

## Validation

- Full offline gate: stub regen, spec/CEL suites for both kinds, targeted Go builds + both Pulumi entrypoints + the Bazel repo build (`make build-go`), foreign-key guard, secret-coverage gate, drift guard + outputs conformance (two new cases), `tofu init`/`validate` + full-surface `plan` proofs for both modules (CodeBuild renders all three resources — project, webhook, resource policy; CodePipeline renders every V2 block), all 8 presets + 3 scenarios + fixtures + hack manifests CLI-validated, site catalog mirror regenerated.
- Guard-tooling note folded into the forge rule: registry-driven guards (`secret-coverage`, `validate-refs`) must run through a CLI built from the working tree — a stale installed binary reads its own compiled-in protos and reports false gaps and false greens alike.
