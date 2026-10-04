# AWS Certificate Manager and CloudFront Rebuild — The Edge Pair

**Date:** 2026-07-05  
**Scope:** Kinds #24–#25 — `AwsCertManagerCert`, `AwsCloudFront`; plus the deferred Lambda/KMS/event-source-mapping live-E2E catch-up

## Summary

The two thinnest specs in the AWS catalog — the 59-line certificate and the
76-line CDN — rebuilt to the full provider surface with dual-engine parity.
ACM now models all three ways a certificate enters ACM (requested, imported,
private-CA issued); CloudFront covers the complete `aws_cloudfront_distribution`
surface including modern Origin Access Control folded into the origin spec.
Both kinds passed live dual-engine E2E, and the three lanes deferred by the
previous session (Lambda, KMS, event source mapping) were all run and flipped
green — surfacing and fixing four real defects along the way.

## Product impact

- **Certificates gain the full lifecycle** — bring-your-own PEM material
  (`imported`), private-CA issuance (`certificate_authority_arn`), per-domain
  validation routing, CT-logging/export options, and key algorithms. The
  Route53 zone is now OPTIONAL: a cert without one exports its validation
  records (`domain_validation_records`) for external DNS — the
  external-DNS workflow gets its own preset. `cert_arn` is unchanged for all
  five downstream consumer kinds.
- **CloudFront manifests express the real CDN surface** — multiple origins
  (custom, S3 with one-toggle OAC creation, VPC origins), origin groups with
  failover, ordered behaviors with modern cache/origin-request/response-header
  policy IDs (XOR the legacy `forwarded_values`), custom error responses, geo
  restrictions, access logging, and WAF attachment by reference. The
  `domain_name` and `hosted_zone_id` outputs consumed by DNS alias wiring are
  unchanged; `distribution_arn` and `status` are new.
- **Serverless E2E confidence** — Lambda, KMS, and event-source-mapping kinds
  are now proven by live deploy → verify → destroy runs on both engines, not
  just the offline gate.

## Technical highlights

- **OAC fold:** `s3_origin.create_origin_access_control` provisions a
  SigV4-signing Origin Access Control per origin; a shared external OAC or a
  legacy OAI can be referenced instead — all three arms of private-S3 access
  in one message.
- **Referential CEL:** every behavior's `target_origin_id` must name a
  declared origin OR origin group — broken wiring is unrepresentable.
- **New harness verifiers:** `awss3objectset` (per-key HeadObject via a new
  `OutputsVerifier` interface — the kind also gained a `bucket_id`
  output) and `awssqsqueue` (GetQueueAttributes). These closed the gap that
  made the Lambda/ESM prerequisite chains unverifiable.
- **E2E scenario honesty (ACM):** a requested DNS-validated certificate on an
  unowned domain goes `ValidationStatus: FAILED` at the AWS API — no
  validation records ever populate, deadlocking the provider's create waiter
  on both engines. The live leaf is an imported self-signed certificate
  (10-year validity, throwaway material committed in the scenario; deploys in
  seconds, deletes immediately).

## Live defects found and fixed

- **HCL `&&` does not short-circuit** (ACM TF): a conditional dereferenced
  the absent `imported` block; rewritten with `try()`.
- **`can()` is not a null guard** (CloudFront TF): `can(o.s3_origin)` returns
  true when `s3_origin` is null, so the second `&&` operand still
  dereferences null. All nested-optional guards now use `try(obj.attr,
  default)` — folded into the Terraform authoring rule.
- **`options: binary: main` in Pulumi.yaml** (event source mapping): the
  engine execs a prebuilt binary instead of compiling — every `pulumi up`
  from a clean checkout fails. Removed; the Pulumi entrypoint rule now
  forbids it.
- **Lambda execution role lacked SQS consume permissions**: AWS validates at
  CreateEventSourceMapping that the function role can
  ReceiveMessage/DeleteMessage/GetQueueAttributes; the shared E2E role
  fixture gained the inline policy.

## Breaking changes

| Kind | Change |
|------|--------|
| AwsCertManagerCert | Route53 zone optional; three exclusive creation modes; outputs enriched (spec shape changes) |
| AwsCloudFront | Full spec rebuild — origins require `origin_id`, behaviors target origins by id, viewer certificate is a block (top-level `certificateArn` gone) |
| AwsS3ObjectSet | New `bucket_id` output (additive) |
| charts/aws/static-website | Breaks on new CloudFront shape (`isDefault`, top-level `certificateArn`) — charts wave, not this session |

ACM's chart consumers (`ecs-environment`, `microservices-backend`) were
checked field-by-field and are unaffected.

## Validation

- Offline gate green: Ginkgo spec tests (both kinds), variables.tf drift
  guard, outputs conformance, `tofu validate` (both modules), Go + Bazel
  builds (modules, harness, verify package), `validate-refs --check`,
  `secret-coverage --check` (new sensitive `private_key` covered).
- Live dual-engine E2E, six lanes green: ACM 19s/39s, CloudFront
  5m40s/5m39s, Lambda 3m37s/3m53s, KMS 29s/44s, event source mapping
  7m11s/7m45s. All six profiles now `status: green`.
- Zero-orphan sweep verified clean across ACM, CloudFront, IAM, S3, SQS,
  Lambda, and event source mappings.
