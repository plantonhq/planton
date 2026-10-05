# Kubernetes: batched EKS real-cluster E2E lanes — thirteen deferred proofs retired live

## Summary

The Kubernetes E2E harness gained a first-class real-cluster batch story,
and the deferred proofs that no local kind cluster can make were executed
against a batch-provisioned EKS cluster on BOTH engines: a real Karpenter
node launch (pending pod → EC2 node → consolidation reclaim), real
ClusterAutoscaler ASG scale-up/down, a CSI-snapshot disaster-recovery loop
into real S3 through keyless IRSA, real Route 53 record writes, a real
Secrets Manager read, SQS-depth-driven scaling, real NLB/Gateway addresses,
CSI clone and snapshot-restore data proofs, and live scheduler preemption.

## Framework

- **Cluster-profile matching in the external lane**: the run declares what
  its external cluster IS (`PLANTON_E2E_CLUSTER_PROFILE`); profiled
  scenarios run only on a matching declaration and skip with the reason
  otherwise — a profiled scenario can never run onto the wrong real
  cluster.
- **`aws-eks` real-cluster scenario profile** (no local constructor; local
  runs skip with the reason) and the **`real_cluster` kind profile
  status** (kinds whose every lane needs a real cluster — the
  Karpenter family; excluded from kind CI matrices by construction).
- **`${E2E_ENV:PLANTON_E2E_*}` manifest tokens** (prefix-fenced, loud on
  unset or non-prefixed names) so committed scenarios reference
  batch-specific identifiers without hardcoding an account.
- **`planton.dev/e2e-engines` scenario annotation**: documented
  PARITY-EXCEPTION arms skip the inexpressible engine's lane with the
  reason (first user: PVC data sources, Terraform-inexpressible).
- **Temp-copy identity contract**: both manifest-rewriting sites (token
  expansion and foreign-key resolution) now preserve the scenario
  basename — verifier dispatch keys behavioral variants off the scenario
  name, and a randomly named copy silently demoted behavioral lanes to
  install-grade passes. Unit tests lock both sites.
- **`aa_e2e/realcluster/aws-eks/` runbook** committed: eksctl cluster
  config, vendored upstream Karpenter prerequisite CloudFormation,
  bootstrap (IRSA roles trust-scoped to chart-derived ServiceAccount
  names, discovery tags, batch test resources, storage scaffolding),
  teardown handling the non-obvious residues, and an AWS-side audit
  script that fails on any surviving resource.

## Defects caught live and fixed

- external-dns Terraform: `try(x, "") != ""` passes a present-but-null
  `optional` scalar into a string template — fixed with the
  `try(coalesce(x), "")` idiom; lesson added to the update rule.
- KarpenterNodePool Terraform: the module passed the proto's deliberately
  flattened template through verbatim, rendering keys at the wrong CRD
  nesting level (server-side apply rejects; no offline gate can see it) —
  locals now rebuild the CRD's template.metadata/template.spec nesting as
  the Pulumi builder's twin.
- Karpenter scenario moved off kube-system to a dedicated namespace:
  keep-on-uninstall CRDs pin their Helm release namespace, and the E2E
  cleanup contract asserts namespace absence; forge rule now teaches both
  (plus sweep-by-release-annotation).
- Velero verifier: Backup/Restore names made run-unique — backup names are
  object-store keys that outlive the CR on persistent buckets.
- Forge-rule additions: saturation fixtures must skip await; verifier-owned
  driver resources keyed in persistent external stores need run-unique
  names; a behavioral lane's evidence is its verifier's own behavioral
  output, never the PASS alone.

## Validation

Smoke lane first (credential chain proof), then per-kind lanes on both
engines with full six-phase runs; import round-trips for Karpenter (scoped
multi-release), Ec2NodeClass, and NodePool; runner unit tests, targeted
builds, `make build-go`, manifest CLI validation, and offline tofu proofs
for both fixed modules. Full batch teardown verified by the committed audit
script.
