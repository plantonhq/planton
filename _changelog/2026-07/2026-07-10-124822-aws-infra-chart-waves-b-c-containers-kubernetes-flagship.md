# AWS Infra-Chart Waves B + C: Containers, Kubernetes, and the Flagship

**Date**: July 10, 2026
**Type**: Feature
**Components**: Infra Charts, AWS Provider, Chart Authoring Rules

## Summary

Five AWS infra-charts forged from first principles against the rebuilt
90/10 kind surface: `fargate-web-service`, `app-runner-service`, and
`ci-cd-pipeline` (the container/delivery tier), plus `eks-platform` and the
`production-web-stack` flagship. The AWS chart catalog now stands at 10.
Every chart passed the full offline gate — structure guard, working-tree
CLI `chart validate` across defaults plus every bool-toggle variant
(28 variants total across the five), live icon URL checks — and the chart
forge rule gained two durable renderer-mechanics lessons proven during the
build.

## The Charts

### fargate-web-service

The production container service: ECS Fargate behind an internet-facing
ALB, request-count autoscaling composed from the ALB's and target group's
`arn_suffix` outputs, a private ECR repository (immutable tags with a
`latest` exclusion, scan-on-push, untagged pruning), split
execution/task IAM roles, and a two-tier security-group contract
(world → ALB → tasks). Toggles: `ecr_repo_enabled`,
`container_insights_enabled`, `https_enabled` (443 + HTTP→HTTPS 301 with a
bring-your-own ACM certificate), `waf_enabled` (REGIONAL managed-rules
ACL attached via the ALB's `webAclArn`), `dns_enabled` (dual-stack A/AAAA
aliases with `evaluateTargetHealth` on), and `assign_public_ip` for
NAT-less networks. Deploys green out of the box on a public httpd sample;
the README walks the push-to-ECR first deploy. 20 params, 7 validation
variants.

### app-runner-service

The simplest production deploy: an App Runner service with its
companions modeled as the first-class, versioned resources they are — a
dedicated `AwsAppRunnerAutoScalingConfiguration` (concurrency trigger,
warm floor), an X-Ray `AwsAppRunnerObservabilityConfiguration` enabled by
reference, and an optional `AwsAppRunnerVpcConnector` with its own
egress-only security group (the group other resources reference to admit
the app). The ECR access role and instance role stay split
(deploy-time vs runtime identity). `auto_deployments_enabled` is guarded
in the template to the private-ECR arm, mirroring the spec's own
ECR_PUBLIC CEL so every independent toggle flip renders valid. Custom
domains associate on the service; the CNAME + certificate-validation
records are a README day-2 recipe because their values are deploy-time
outputs (see the rule uplift below). 6 validation variants.

### ci-cd-pipeline

Push-to-production on AWS-native tooling: a CodePipeline V2 (git push
trigger, QUEUED execution mode) from a CodeConnections source through a
Docker-enabled CodeBuild project into a private ECR repository, with a
hardened artifact bucket (SSE, public-access block, 30-day artifact
expiry) and two least-privilege roles whose inline policies are scoped to
the chart's own resources by render-time name (account segments
wildcarded, with the day-2 tightening taught). The pipeline wires the
build project by rendered name — AWS's action `configuration` is a
`map<string,string>`, so the chart closes that loop at render time — and
the connection ARN is the one literal the user must bring (the
CodeConnections handshake is human-approved by design; the README walks
it). Optional ECS deploy stage pairs with fargate-web-service via
`imagedefinitions.json`; the README ships the complete buildspec.

### eks-platform

A production Kubernetes platform with access-entry authentication
(`authenticationMode: API` — access entries are the only auth path, no
aws-auth ConfigMap), audit + authenticator control-plane logs, cluster
and node IAM roles, and the compute shape behind one toggle:
`auto_mode_enabled` off (default) renders a managed AL2023 node group
plus the four core add-ons as first-class resources (vpc-cni, coredns,
kube-proxy, eks-pod-identity-agent, each with OVERWRITE create-conflict
resolution); on renders the Auto Mode trio and deliberately NO add-ons —
Auto Mode manages networking/DNS itself, and the chart owns this
either/or because the protos leave it un-CEL'd. `AwsIamOidcProvider` on
the cluster's `oidc_issuer_url` output wires IRSA (the README teaches the
per-workload trust-policy loop), and an admin `AwsEksAccessEntry`
prevents the classic API-mode lockout. 6 validation variants.

### production-web-stack (the flagship)

Startup-in-a-box, the catalog's one deliberate kitchen-sink: a two-tier
VPC (public/private subnets per AZ with subnet-owned routes, single
shared EIP-backed NAT), the full Fargate web chain (ECR, split roles,
ALB/TG/listeners, circuit-breaker service with request-count
autoscaling), Aurora PostgreSQL Serverless v2 (managed master password
exporting the secret ARN, storage encryption hardcoded on, final
snapshot + PITR, writer + tier-1 reader), optional two-node HA Redis
(both encryptions on, task-SG-only ingress, deny-all egress), the
custom-domain plane (zone-create or bring-your-own, DNS-validated ACM
cert, HTTPS + 301, dual-stack aliases), a default-on REGIONAL WAF, and
the pager wire: SNS topic + email subscription + four alarms whose
dimensions are all render-stable names (service CPU/memory via
ClusterName/ServiceName, database CPU + ACUUtilization via
DBClusterIdentifier). ALB 5xx/latency alarms are deliberately a README
day-2 recipe: their dimensions are deploy-time `arn_suffix` values, and a
pre-rendered placeholder would silently watch nothing. The task role
ships with a name-pattern-scoped read on the managed DB secret
(`rds!*`), with the exact-ARN tightening taught. 26 resources at full
toggle depth; 7 validation variants.

## Rule Uplifts (learn-once)

Two renderer-mechanics lessons proven against the validator during the
build, folded into `_rules/charts/forge-planton-infra-chart.mdc`:

1. **Outputs only feed `StringValueOrRef` fields.** Plain `string`,
   `repeated string` (a DNS record's `values`), and `map<string,string>`
   (alarm `dimensions`, pipeline action `configuration`) take render-time
   literals only — the typed loader rejects a `valueFrom` object where it
   expects a scalar. Close such loops with render-time naming the chart
   controls on both sides, design around render-stable values, or make
   the wiring an explicit day-2 README recipe; never render a placeholder
   where a real value belongs.
2. **API-coupled toggles are guarded together in templates** (each bool
   is validation-flipped independently, so a pair only valid in
   combination is a design defect), and `{% if not (...) %}` negated arms
   are how a chart owns an either/or shape the protos deliberately leave
   un-CEL'd — including suppressing the neighbors that must not render
   beside the active arm.

## Validation

- `bash hack/guards/ensure_chart_structure.sh` — green.
- Working-tree CLI (`go build -o /tmp/planton .`) `chart validate` per
  chart: fargate-web-service 7/7, app-runner-service 6/6, ci-cd-pipeline
  2/2, eks-platform 6/6, production-web-stack 7/7 variants.
- Provider-wide sweep: `chart validate --all charts/aws` — **10 passed,
  0 failed out of 10**.
- Icon URLs verified live (HTTP 200 ×5: awsecsservice,
  awsapprunnerservice, awscodepipeline, awsekscluster, awsalb).
- No cloud interaction: charts are configuration artifacts; the offline
  gate mirrors CI `lint.charts` exactly. Server-side `chart build` proof
  belongs to platform integration.

## Impact

The AWS chart catalog doubles to 10 of the planned 17, now covering the
highest-demand production paths: containers three ways (Fargate, App
Runner, EKS), native CI/CD, and the complete startup stack. Every chart
composes the rebuilt kind surface through typed references, renders
valid in every toggle variant, and documents its own day-2 evolution —
the catalog a team browses and finds the thing they were about to build
by hand.
