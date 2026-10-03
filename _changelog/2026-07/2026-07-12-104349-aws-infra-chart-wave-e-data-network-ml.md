# AWS Infra-Chart Wave E: Data Lakehouse, Kafka Streaming, Private Network Hub, and ML Workbench

**Date**: July 12, 2026
**Type**: Feature
**Components**: Infra Charts, AWS Provider, Chart Authoring Rules

## Summary

The final four AWS infra-charts, forged from first principles against the
rebuilt 90/10 kind surface: `data-lakehouse`, `kafka-streaming`,
`private-network-hub`, and `ml-workbench`. The AWS chart catalog is now
complete at 17. Every chart passed the full offline gate — structure
guard, working-tree CLI `chart validate` across defaults plus every
bool-toggle variant (11 variants across the four), the provider-wide
sweep (17/17), and live icon URL checks — with zero cloud interaction
(charts are configuration artifacts).

## The Charts

### data-lakehouse

A queryable data lake: an S3 lake bucket whose lifecycle tiers aging data
STANDARD → STANDARD_IA → GLACIER_IR (deliberately stopping at Glacier
Instant Retrieval — Athena can still query it in place; deep-freeze
classes that would strand the lake from its own query engine are a
taught day-2 decision, not a default), a Glue catalog database whose
`locationUri` and the Athena workgroup's `outputLocation` are closed at
render time through chart-controlled bucket names (both are plain-string
S3 URIs by the specs' design), and an Athena workgroup with enforced
configuration and a per-query scan ceiling expressed in GiB and converted
inline (`{{ values.query_scan_limit_gib * 1073741824 }}`). Toggles:
`managed_results_enabled` (AWS-managed 24-hour result storage XOR a
dedicated 30-day-expiry results bucket — one toggle owning both sides of
the spec's `managed_results_excludes_output_location` CEL),
`ingestion_enabled` (a Direct-PUT Firehose landing GZIP,
date-partitioned JSON plus its least-privilege delivery role), and
`force_destroy`. The Glue database name is its own underscore-safe param
because Glue names admit no hyphens. 9 params, 4 validation variants.

### kafka-streaming

Managed Kafka on MSK, serverless-first: `provisioned_enabled` owns the
either/or between `AwsMskServerlessCluster` (per-GB billing, zero broker
management) and `AwsMskCluster` (broker count / instance type / EBS
sizing params, SASL/IAM only). Both arms authenticate with IAM over TLS
on port 9098 behind the same two-tier security-group contract — a client
group applications attach (membership IS the network grant) and a
cluster group admitting 9098 only from client-group members by
reference — so flipping the purchase model never touches client code or
network posture. The network is bring-your-own (`vpc_id` +
`broker_subnet_ids` params with the network-foundation `valueFrom` swap
taught in the README), and the README carries the scoped
`kafka-cluster:*` IAM client policy and the honest serverless/provisioned
crossover math. 9 params, 2 validation variants.

### private-network-hub

A Transit Gateway hub for EXISTING VPCs — the chart deliberately creates
no spoke networks (a hub's spokes are real workload VPCs; the
bring-your-own grain the data-plane charts set). The gateway pins both
default-route-table dials off and each spoke's attachment pins its
default-table memberships off, so routing-domain membership is always
explicit — also dodging AWS's disabled-to-enabled-replaces-the-gateway
trap. Each spoke associates with its own `AwsTransitGatewayRouteTable`,
and one toggle (`spokes_interconnected_enabled`) selects between two
real postures: segmented (default — no cross-domain propagation plus a
blackhole guardrail for the peer spoke's CIDR, so isolation survives
future broad routes by longest-prefix match) and interconnected (each
domain propagates the other's attachment; the guardrails do not render).
The README carries the spoke-side return-route recipe (the subnet
spec's polymorphic `transit_gateway` route target), the
shared-services-domain and inspection/egress day-2 paths, and AWS's
one-association-per-attachment rule. 9 params, 2 validation variants.

### ml-workbench

SageMaker Studio for a team: a domain with idle auto-shutdown on
JupyterLab and Code Editor (the spec's all-three-required trio, domain
defaults every profile inherits), a least-privilege execution role
(`AmazonSageMakerFullAccess` starter plus explicit render-time grants on
the artifacts bucket, which the managed policy's "sagemaker-in-the-name"
S3 fence would otherwise exclude), a versioned artifacts bucket with
noncurrent-version pruning, and a dedicated internet-path-free VPC (two
routeless private subnets, no IGW/NAT — in `PublicInternetOnly` mode
Studio egress rides the service network, so the network costs nothing).
The `vpc_only_enabled` toggle flips `appNetworkAccessType` to `VpcOnly`
AND renders everything that makes the locked-down posture actually
function: a self-referencing inter-app security group (ephemeral ports +
NFS 2049), an endpoints group admitting 443 from Studio apps only, the
kind documentation's full endpoint set (SageMaker API/runtime, STS,
CloudWatch Logs, ECR API/Docker, Service Catalog — all with private DNS,
backed by the VPC's two DNS switches) and a free S3 gateway endpoint
attached to the VPC MAIN route table via an explicit `valueFrom`
(routeless subnets ride the main table; their own `route_table_id`
outputs are deliberately empty). 10 params, 3 validation variants.

## Design decisions

- **A hub attaches real networks.** `private-network-hub` takes spoke
  VPCs as parameters instead of shipping demo spokes: throwaway VPCs on
  a connectivity hub would be catalog filler, and the two-posture toggle
  (segmentation XOR interconnection) makes both branches of validation a
  real architecture rather than one real arm and one degenerate render.
- **A workbench owns its environment.** `ml-workbench` creates its own
  VPC — the catalog's established split: plug-in charts join an
  application network by params (aurora-postgres, kafka-streaming);
  environment charts own theirs (network-foundation,
  production-web-stack). A Studio domain is an isolated environment, and
  its zero-cost routeless VPC makes the chart deployable in a fresh
  account.
- **Tier the lake only as far as the query engine reaches.** The
  lakehouse lifecycle stops at GLACIER_IR; deep-archive transitions are
  taught as a deliberate day-2 rule because Athena cannot read them in
  place.
- **One toggle per API contract.** Athena's managed-results exclusivity,
  MSK's purchase-model either/or, and the hub's posture flip each live
  behind exactly one bool whose branches are independently valid — the
  honest-toggle discipline the validator's per-bool flipping enforces.

## Validation

- `hack/guards/ensure_chart_structure.sh` — pass.
- Working-tree CLI (`go build -o /tmp/planton .`) `chart validate` per
  chart: data-lakehouse 4 variants, kafka-streaming 2, private-network-hub
  2, ml-workbench 3 — all green.
- Provider-wide `chart validate --all charts/aws` — **17 passed, 0 failed
  out of 17**.
- Icon URLs verified live (HTTP 200) for all four charts; kafka-streaming
  carries the MSK service logo via the `awsmskcluster` asset path (the
  serverless kind's logo remains a recorded platform asset gap).
- Leakage grep clean on every commit.

## Related Work

- Waves A–D forged the first 13 charts
  (`2026-07-10-115240`, `2026-07-10-124822`, `2026-07-12-083940`).
- The chart forge rule gained two proven renderer mechanics from this
  build: inline arithmetic for unit-friendly params (GiB params feeding
  byte fields) and template-literal list loops with filters for
  chart-designed resource families (the endpoint set).

---

**Status**: ✅ Production Ready
