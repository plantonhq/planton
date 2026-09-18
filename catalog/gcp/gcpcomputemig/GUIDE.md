# GcpComputeMig Guide

Operational judgment for running managed instance groups as code — the
things the spec reference cannot tell you.

## The template rotation is the deploy pipeline

Every `template` change rotates the template and — with `updatePolicy`
`PROACTIVE` — rolls the fleet. That makes this ONE resource your VM
deploy pipeline: bake an image, change `sourceImage`, apply, and the
group replaces instances within the surge/unavailability budget. Two
consequences worth internalizing: (1) size the budget for your traffic —
`maxSurgeFixed` above zero buys zero-unavailability rollouts at
temporary double-capacity cost; (2) `OPPORTUNISTIC` means an applied
change may sit unapplied for days on a stable fleet — pair it with
deliberate refresh operations, not with the expectation of convergence.
The `current_template_self_link` output changes on every rotation —
compare across applies to confirm a rollout actually rolled.

## RECREATE vs SUBSTITUTE is an identity decision

`SUBSTITUTE` (default) gives replaced instances fresh random names —
surge works, rollouts can be zero-downtime, and nothing downstream may
depend on instance names. `RECREATE` preserves names — what stateful
consumers and per-instance configs want — but the group must take
instances DOWN to replace them (the spec enforces the unavailability
budget above zero). Decide per workload, not per rollout: flipping the
method later churns the whole fleet's identity.

## Auto-healing thresholds are not LB thresholds

Wire a DEDICATED GcpHealthCheck into `autoHealing`, tuned conservative
(long interval, high unhealthy threshold), and let the load balancer's
own health check stay aggressive. Sharing one aggressive check turns a
slow garbage-collection pause into a repair storm: the LB should stop
ROUTING to a slow instance in seconds; the group should REPLACE it only
when it is genuinely gone. Size `initialDelaySec` to the real cold start
(app boot + warmup), not to the VM boot.

## Stateful groups are a set of one-way doors

`statefulDisks`, stateful IPs, and `perInstanceConfigs` change the
group's character: repairs preserve state, updates stop rebalancing
(regional groups want `instanceRedistributionType: NONE`), and the
autoscaler cannot scale IN safely below the configured instances. The
per-instance-config destroy knobs matter at teardown:
`removeInstanceOnDestroy` deletes the INSTANCE with its config;
`removeInstanceStateOnDestroy` strips just the state; the preserved
disks' `deleteRule` decides whether data survives permanent instance
deletion. Rehearse the teardown path before trusting it with data.

## Standby pools trade money for cold-start time

`standbyPolicy` mode `SCALE_OUT_POOL` with `targetSuspendedSize` keeps
pre-booted VMs suspended: scale-outs resume in seconds instead of
booting in minutes, while suspended VMs bill for disks and memory only.
The economics work when your scale-out latency SLO is tighter than a VM
boot and the workload's boot is expensive (large images, slow warmup) —
otherwise it is idle spend. Stopped standbys (`targetStoppedSize`) are
cheaper and slower; mixing both tiers is legal.

## Resize requests are asks, not guarantees

A `resizeRequests` entry queues a request for capacity (the Dynamic
Workload Scheduler path for scarce shapes — GPUs, large memory). It may
sit QUEUED indefinitely, and granted instances are RECLAIMED when
`requestedRunDurationSeconds` expires. Watch the request's state through
the API, and treat quota as the first suspect when nothing gets granted:
the ask counts against the same regional quota as ordinary instances.

## Quota is a fleet-multiplied problem

A rollout with surge, a standby pool, and an autoscaler burst can
demand `targetSize + surge + standby + burst` instances of quota
simultaneously — in ONE zone for zonal groups. Regional groups spread
the demand but require quota in EVERY zone the distribution policy
names. Check `compute.googleapis.com/cpus` and per-family quotas against
the worst case, not the steady state.

## Teardown discipline

`deletionPolicy: DELETE` (default) tears down the whole stack — VMs
terminate, per-VM disks follow their template `autoDelete`, stateful
disks follow their `deleteRule`, and an ACTIVE resize request is
cancelled. `ABANDON` leaves the fleet running unmanaged — the escape
hatch when handing a serving tier to another team's IaC. `PREVENT` fails
the destroy outright. One provider asymmetry recorded honestly: the
zonal instance TEMPLATE carries no deletion policy (it is always deleted
on destroy) — the regional one participates like every other resource.

## Managed workload identity and host-error recovery

`template.workloadIdentityConfig` issues every VM in the group a SPIFFE
identity (and, with `identityCertificateEnabled`, rotated X.509
certificates) so services authenticate to each other by identity rather
than by shared secret or network position. It is part of the template,
so changing it rotates the template and rolls the group.
`template.scheduling.hostErrorTimeoutSeconds` (90..330 in steps of 30)
tightens how quickly Compute Engine declares a hung host failed and
starts recovery; leave it unset for the default timing.

## Coverage decisions on record

CSEK raw-key encryption arms are deliberately not modeled (raw key
material does not belong in manifests or state — use CMEK). Legacy
preemptible-only VMs are not modeled: set
`scheduling.provisioningModel: SPOT` and both engines derive the legacy
flag the API still requires.
