# AWS Infra-Chart Wave D: Serverless REST API, Event Backbone, and Transactional Email

**Date**: July 12, 2026
**Type**: Feature
**Components**: Infra Charts, AWS Provider, API Definitions, Chart Authoring Rules

## Summary

Three AWS infra-charts forged from first principles against the rebuilt
90/10 kind surface: `serverless-rest-api`, `event-driven-backbone`,
and `transactional-email`. The AWS chart catalog now stands at 13. Every
chart passed the full offline gate — structure guard, working-tree CLI
`chart validate` across defaults plus every bool-toggle variant (14
variants total across the three), provider-wide sweep (13/13), and live
icon URL checks. The build also surfaced and fixed a kind defect at
the root — the Route 53 DNS record's name pattern rejected the
underscore-prefixed service records (DMARC, DKIM, SRV) its own field
comments promise to support — and published the missing kind logos
(the SES pair plus ten stale-CDN entries) to the asset CDN.

## The Charts

### serverless-rest-api

A production serverless REST API: one Lambda function (a deliberate
"Lambdalith" — the handler's own router owns the URL space) behind an
HTTP API Gateway `$default` catch-all on an auto-deployed stage, with an
on-demand DynamoDB table (PITR, deletion protection, single-table-design
key shape), a least-privilege execution role (logs + item-level table
actions composed by render-time name; Scan deliberately ungranted), and
error/throttle alarms wired to email. Toggles: `container_image_enabled`
(ECR image XOR S3 zip — the template owns Lambda's exactly-one-code-source
CEL), `sort_key_enabled`, `table_deletion_protection`, `auth_enabled`
(Cognito user pool + SRP app client + JWT authorizer + route auth, one
toggle owning the whole CEL-coupled set), `cors_enabled`, and
`alarms_enabled`. Two honesty calls: no dead-letter queue (Lambda DLQs
only receive asynchronous invocations, and API Gateway invokes
synchronously — a DLQ here would sit forever empty looking like a safety
net), and gateway-level 5xx/latency alarms are a README day-2 recipe
because their `ApiId` dimension only exists after deployment. The
function's invoke permission is account-scoped at render time with the
exact execution-ARN tightening taught as day-2. 20 params, 7 validation
variants.

### event-driven-backbone

The event backbone decoupled services are built on: a custom EventBridge
bus (the domain's own namespace, not the account default), a routing rule
prefix-matching the event `source`, and a durable SQS work queue with a
locked-down DLQ behind it — 14-day retention, `byQueue` redrive
allowlist, and a depth alarm that pages the moment any dead letter
appears. Toggles: `sns_fanout_enabled` (an SNS topic target plus one
subscriber queue with raw delivery — the copy-per-team pattern shipped
end to end) and `lambda_consumer_enabled` (execution role + function +
event source mapping with partial-batch failure reporting). Every
delivery permission is a resource policy composed at render time from
names the chart controls — the rule ARN and topic ARN in `aws:SourceArn`
conditions scope each `SendMessage`/`Publish` grant to exactly this
chart's resources, closing the silently-dropped-delivery trap the SNS
subscription spec documents. One structural subtlety: the DLQ's redrive
allowlist uses literal ARNs deliberately, because referencing the work
queue would create a dependency cycle with the work queue's own
dead-letter reference. 13 params, 4 validation variants.

### transactional-email

Production sending on SES: a DKIM-signed domain identity (Easy DKIM
2048 stated explicitly), a configuration set enforcing the delivery
posture (TLS required by default, bounce/complaint suppression,
reputation metrics on), a bounce/complaint feedback topic scoped to
exactly this configuration set's ARN, and reputation alarms at the
thresholds AWS itself acts on (bounce ≥ 5%, complaint ≥ 0.1%, hourly
windows, per-set `ses:configuration-set` dimension). Toggles:
`mail_from_enabled` (the custom envelope domain that makes SPF align for
DMARC), `dns_records_enabled` (DMARC TXT with a policy-ratchet param +
report address, and — guarded together with MAIL FROM, the API-coupled
toggle pattern — the MX to the regional SES feedback endpoint and the
SPF TXT, all render-stable), `tls_required`,
`feedback_notifications_enabled`, and `alarms_enabled`. The DKIM CNAME
trio is the explicit post-deploy step in the README — the tokens are
generated with the identity, so no template can publish them; the
feedback (machine) and alerts (human) topics are deliberately separate
so a bounce storm pages once. 15 params, 6 validation variants.

## Kind Fix: Route 53 DNS Record Name Pattern

The transactional-email chart's `_dmarc.<domain>` TXT record — a
composition the DNS record spec's own comments and the SES kind's docs
prescribe — failed validation: the `name` field's pattern admitted only
`[A-Za-z0-9\-\.]` labels, rejecting the underscore convention that DMARC,
DKIM (`<token>._domainkey.<domain>`), SRV (`_sip._tcp`), and ACME
challenge records are built on. Fixed at the root rather than worked
around in the chart:

- `apis/dev/planton/provider/aws/awsroute53dnsrecord/v1/spec.proto` —
  the pattern now admits `_` in every arm, and the field comment
  documents the underscore-label convention.
- `spec_test.go` — a regression case validates `_dmarc.example.com`,
  `token._domainkey.example.com`, and `_sip._tcp.example.com`.
- Stubs regenerated (`make protos`); the spec test package is green; the
  working-tree CLI was rebuilt before re-validation.

## Kind Logo Assets Published

The chart icon check surfaced that the SES kinds' logos had never been
published, and a full sweep of the AWS kind-logo URLs found 20 returning
404. Fixed to the extent the sources allow:

- The official AWS SES architecture icon was staged in the platform
  repo's asset tree for `awssesemailidentity` and
  `awssesconfigurationset` and published to the `planton-assets` R2
  bucket (with the CDN cache purge that clears negative-cached 404s).
- Ten kinds whose logos existed in the platform tree but had never been
  synced were published the same way (ECS task definition, EKS access
  entry/addon/Fargate profile, Lambda event source mapping, LB
  listener/rule/target group, NLB, SNS subscription). All verified 200.
- Ten kinds have no logo file in the platform tree at all (ASG,
  ElastiCache user/user group, IAM instance profile/policy, launch
  template, MSK serverless, Redshift serverless pair, VPC endpoint) —
  recorded as platform-side debt; publishing requires sourcing each
  official icon.

The chart forge rule now carries the full icon-publishing pipeline
(asset source of truth, sync/PUT paths, the tree-vs-CDN trap, the
negative-cache purge) and a new guardrail: a chart validation failure
can be the kind's defect — fix the kind at the root, never
bend the chart around it.

## Validation

- `hack/guards/ensure_chart_structure.sh` — pass.
- Working-tree CLI `chart validate` per chart: `serverless-rest-api`
  (7 variants), `event-driven-backbone` (4), `transactional-email` (6)
  — all green.
- Provider-wide `chart validate --all charts/aws` — **13/13**.
- `go test ./apis/dev/planton/provider/aws/awsroute53dnsrecord/v1/...`
  — green after the pattern fix.
- Icon URLs curl-verified 200 for all three charts (and the 12 newly
  published kind logos).
- Zero cloud provisioning — charts are configuration artifacts; the
  offline gate mirrors CI `lint.charts` exactly. Server-side
  `chart build` proof remains platform-integration scope.

## Impact

The AWS catalog's serverless and messaging tiers are now covered by
composable, deploy-green-by-default charts, and the email chart makes
SES's deliverability setup — the part teams most often get wrong — a
parameter form. The DNS record pattern fix unblocks every future
composition that publishes service records (DMARC, DKIM, SRV, ACME)
through `AwsRoute53DnsRecord`, for every consumer, not just charts.

## Related Work

- Waves A–C: `2026-07-10-115240-aws-infra-chart-wave-a-foundations-and-web.md`,
  `2026-07-10-124822-aws-infra-chart-waves-b-c-containers-kubernetes-flagship.md`
- The catalog design and authoring bar:
  `2026-07-10-105115-aws-infra-chart-catalog-clean-slate-and-state-backend-specials.md`

---

**Status**: ✅ Production Ready
