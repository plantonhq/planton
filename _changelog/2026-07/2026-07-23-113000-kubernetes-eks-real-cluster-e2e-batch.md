# Kubernetes real-cluster E2E: the batched EKS lane framework, twelve cloud-fabric behavioral proofs, and three live-caught module fixes

**Date**: 2026-07-23
**Scope**: `apis/dev/planton/qa/catalogkinde2eprofile` (new `real_cluster` profile status), `e2e` + `e2e/framework` (external-lane cluster-profile matching, `${E2E_ENV:...}` manifest tokens, engine-restriction annotation, expanded-manifest basename preservation), `apis/dev/planton/provider/kubernetes/aa_e2e` (aws-eks cluster profile, `realcluster/aws-eks` batch runbook, seven new/extended behavioral verifiers), twelve new/updated E2E scenarios across ten kinds, `kubernetesexternaldns` (TF optional-scalar null-guard fix), `kuberneteskarpenternodepool` (TF template-nesting faithful-projection fix), Karpenter-family E2E profiles (deferred → real_cluster, both provisioners validated), `internal/cli/ui/e2ediscover` (REAL-CLUSTER status rendering), `_rules/catalog-kind` (four timeless lessons)

## What changed

### The real-cluster lane framework

E2E assertions that need cloud fabric no kind cluster can provide — cloud
load balancers, IRSA identity hops, snapshot-capable CSI storage, real node
autoscaling — now have a first-class home:

- **External-lane profile matching.** The external-cluster lane declares
  what its cluster IS (`PLANTON_E2E_CLUSTER_PROFILE`); profile-annotated
  scenarios run only on a matching declaration and skip with the reason
  otherwise. Previously every profiled scenario was assumed suitable for
  the external cluster — a sweep would have installed the `cilium-cni`
  profile's primary CNI onto a shared real cluster and destroyed it for
  every later lane.
- **The `aws-eks` cluster profile**: a real-cluster profile with no local
  constructor. Scenarios carrying it skip on local runs and execute in
  batched EKS runs. Used by the cloud-LB, IRSA, CSI-storage, preemption,
  and node-autoscaling scenarios of otherwise-locally-green kinds.
- **The `real_cluster` kind profile status** for kinds whose EVERY
  lane needs a real cluster (the Karpenter family): skipped wherever no
  external cluster is supplied, excluded from kind CI matrices (which build
  from `green`), runnable in every EKS batch.
- **`${E2E_ENV:PLANTON_E2E_*}` manifest tokens**: batch-specific
  identifiers (IRSA role ARNs, bucket/queue/zone ids) expand from the
  bootstrap-exported environment, prefix-fenced so a manifest can never
  read arbitrary process env; unset variables and out-of-prefix names fail
  loudly. Committed scenarios stay honest — no test account's identifiers
  are ever hardcoded.
- **`planton.dev/e2e-engines` scenario annotation**: scenarios exercising a
  surface one engine rejects by documented PARITY-EXCEPTION design (the PVC
  data-source arms, which the Terraform provider cannot express) restrict
  themselves to the expressing engine; the other lane skips with the reason
  instead of failing on its own designed rejection.
- **Expanded-manifest basename preservation**: token expansion writes its
  temp copy under the scenario's own filename — verifier dispatch keys
  behavioral variants off the scenario name in the manifest path, and the
  previous random-only temp name silently demoted every token-carrying
  behavioral scenario to its plain install verifier.
- **The `realcluster/aws-eks` batch runbook** beside the harness: eksctl
  cluster config (system + tainted min-0 `ca-scale` node groups, OIDC, EBS
  CSI + snapshot-controller addons, Karpenter node-role identity mapping),
  the vendored upstream Karpenter prerequisite CloudFormation, bootstrap
  (discovery tags, six IRSA roles trust-scoped to the charts' rendered
  ServiceAccount names, hosted zone, Secrets Manager secret, S3 bucket,
  SQS queue, AWS Load Balancer Controller scaffolding, encrypted-gp3
  StorageClass + VolumeSnapshotClass), teardown (zone records, bucket
  objects, tagged snapshots — the residues a naive cluster delete leaves),
  and a tag-driven audit script that fails on any surviving resource.

### Behavioral proofs landed (all dual-engine on a live EKS cluster)

- **KubernetesKarpenter**: live install (both fixed-name releases, IRSA,
  interruption queue) + blind import round-trip.
- **KubernetesKarpenterNodePool / Ec2NodeClass**: object lanes +
  round-trips, and the node-launch proof — a pending driver pod only the
  pool can satisfy produces a REAL EC2 node, the pod schedules onto it,
  and consolidation reclaims the empty node after deletion.
- **KubernetesClusterAutoscaler**: real ASG scale-up AND scale-down of a
  dedicated min-0 tagged node group (tag auto-discovery + IRSA; scale-down
  tuned through the typed scaling block).
- **KubernetesVelero**: the CSI-snapshot DR loop against a real S3 bucket
  via keyless IRSA — a run-unique marker stamped onto an EBS-backed volume
  survives namespace deletion through snapshot → bucket → restore.
- **KubernetesService**: LoadBalancer provisioning — a real NLB address in
  `.status.loadBalancer.ingress` (deploys deliberately never wait for it;
  the address is the lane's assertion).
- **KubernetesIngressNginx**: the spec's documented AWS annotation recipe
  validated as written — a real NLB provisioned for the controller Service.
- **KubernetesGateway**: an istiod-provisioned Gateway reaches Programmed
  with a real cloud address in `.status.addresses` (the half the kind
  lanes deliberately pin away with ClusterIP mesh fixtures).
- **KubernetesPersistentVolumeClaim**: both data-source arms — snapshot
  restore (verifier-cut VolumeSnapshot) and clone — with marker data read
  from the provisioned volume.
- **KubernetesExternalDns**: the full write-and-own loop against Route 53 —
  A + TXT ownership records written via keyless IRSA, then removed with
  their source (sync policy), asserted through the AWS API.
- **KubernetesClusterSecretStore**: a real Secrets Manager read — the
  verifier-owned ExternalSecret materialized a Secret matching the stored
  plaintext, keyless.
- **KubernetesKeda**: a real 0→1 scale off live SQS queue depth through the
  IRSA pod identity.
- **KubernetesPriorityClass**: preemption under genuine scheduling pressure
  — the high-priority pod schedules only by evicting a low-priority pad,
  with the scheduler's Preempted event as testimony.

### Live-caught fixes (the reason these lanes exist)

- **kubernetesexternaldns (Terraform)**: `try(x, "") != ""` does not guard
  an `optional` scalar inside a present block — unset optionals arrive as
  null and reached a string template ("Invalid template interpolation
  value"). Fixed with the null-safe `try(coalesce(x), "")` read; lesson in
  the update rule.
- **kuberneteskarpenternodepool (Terraform)**: the module passed the proto
  spec through verbatim, but the spec deliberately flattens the CRD's
  `template.metadata`/`template.spec` nesting — server-side apply rejected
  every rendered CR (`.spec.template.expireAfter: field not declared in
  schema`). The module now rebuilds the nesting exactly like its Pulumi
  twin.
- **Velero DR verifier**: backup names are keys in the object store, not
  just CR names — fixed names collide on persistent buckets (the kind
  lanes' ephemeral MinIO masked this). Run-unique names; lesson in the
  forge rule.

## Why

"E2E passed" is a customer-grade promise. The kind lanes prove everything a
local cluster can honestly prove; this batch retires the accumulated rows
that needed real cloud fabric — and it caught three defects (two in shipped
modules, one in the verifier machinery) that no offline gate or kind lane
could see, which is precisely the argument for running them.
