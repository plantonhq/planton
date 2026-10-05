# AwsLambda

AWS Lambda runs code without managing servers. This resource models a Lambda function: code source (S3 zip or ECR container image), execution environment, VPC networking, integrations, and function-scoped satellites (aliases, function URL, invoke permissions, async invoke config, recursion detection, runtime management).

The function name is `metadata.name` (create-time immutable in AWS). There is no separate `function_name` field and no `code_source_type` enum — provide exactly one of `spec.s3` or `spec.image_uri`.

## Spec fields (summary)

**Identity and role**

- `region` — AWS region (required)
- `description` — console description (max 256 characters)
- `role_arn` — execution role (`StringValueOrRef`, defaults to `AwsIamRole` → `status.outputs.role_arn`)

**Code source (exactly one)**

- `s3` — zip in S3 (`bucket`, `key`, optional `object_version`); requires `runtime` and `handler`
- `image_uri` — ECR container URI; `runtime` and `handler` must be empty
- `source_code_hash` — base64 SHA256 for declarative zip updates
- `source_kms_key_arn` — KMS key encrypting the S3 package (zip only)
- `code_signing_config_arn` — code signing config (zip only)

**Execution environment**

- `runtime`, `handler` — required for zip, forbidden for container
- `architecture` — `x86_64` or `arm64`
- `memory_size_mb` — 128–10240 (default 128)
- `timeout_seconds` — 1–900 (default 3)
- `ephemeral_storage_mb` — `/tmp` size 512–10240
- `environment` — plain config map (not for secrets)
- `kms_key_arn` — encrypts environment variables at rest
- `layer_arns` — up to five layer version ARNs (zip only)

**VPC**

- `subnet_ids`, `security_group_ids` — both required together for VPC attachment
- `ipv6_allowed_for_dual_stack` — outbound IPv6 on dual-stack subnets

**Integrations**

- `dead_letter_target_arn` — async failure destination (SQS/SNS)
- `tracing_mode` — `Active` or `PassThrough`
- `file_system_config` — EFS access point mount (requires VPC)
- `image_config` — container entrypoint overrides
- `publish` — publish version on each change
- `publish_to` — maintain the `$LATEST.PUBLISHED` head pointer (`LATEST_PUBLISHED`; still rolling out — regions/accounts without the feature reject it at create)
- `reserved_concurrent_executions` — concurrency cap/reservation
- `snap_start` — SnapStart (requires `publish`)
- `logging_config` — log format, levels, optional `log_group` ref
- `recursive_loop` — `Allow` or `Terminate`
- `code_sha256` — deploy-side digest for out-of-band code-change detection

**Execution platform**

- `managed_instances` — run on a Lambda Managed Instances capacity provider (ARN, per-vCPU memory sizing, per-environment concurrency)
- `durable_config` — durable execution (checkpointed progress, up to 366-day budget; adding/removing REPLACES the function)
- `tenant_isolation_mode` — `PER_TENANT` execution-environment isolation (create-time immutable)

**Satellites (folded into spec)**

- `aliases` — named version pointers, canary weights, provisioned concurrency
- `function_url` — HTTPS endpoint with CORS, optionally alias-qualified (`qualifier`)
- `invoke_permissions` — resource-policy invoke grants (qualifier scoping, Alexa `event_source_token`, URL-path restriction)
- `async_invoke_config` — async retry, event age, destinations; optional `qualifier` scope
- `runtime_management` — runtime patch rollout policy; optional `qualifier` scope
- `scaling_configs` — per-qualifier execution-environment bounds (Managed Instances; version numbers or `$LATEST.PUBLISHED`)

Event sources are **not** on this spec — use `AwsLambdaEventSourceMapping`.

## Outputs

- `function_arn` — join key for mappings and policies
- `function_name` — SDK/CLI name (matches `metadata.name`)
- `invoke_arn` — API Gateway integration ARN
- `qualified_arn`, `version` — latest published version (when `publish` enabled)
- `function_url` — HTTPS URL (when configured)
- `alias_arns` — map of alias name → alias ARN
- `log_group_name` — CloudWatch log group in use

## How it works

The Planton CLI validates the manifest, generates IaC inputs, and invokes IaC backends:

- Pulumi (Go modules under `iac/pulumi`)
- Terraform (modules under `iac/tf`)

Credentials and region live in IaC input (`provider_credential`), not in the spec.

Planton does **not** auto-create a CloudWatch log group, execution IAM role, or invoke permissions. Reference `AwsIamRole` for the role; set `logging_config.log_group` for a managed log group, or let AWS create the default on first invoke.

## References

- [AWS Lambda](https://docs.aws.amazon.com/lambda/latest/dg/welcome.html)
- [Runtimes](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html)
- [Container images](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html)
- [VPC configuration](https://docs.aws.amazon.com/lambda/latest/dg/configuration-vpc.html)
- [Catalog page](./catalog.md)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
