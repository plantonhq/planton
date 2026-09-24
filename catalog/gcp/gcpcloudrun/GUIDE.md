# GcpCloudRun Guide

The judgment this guide protects: a Cloud Run service is two objects
wearing one name — a stable serving endpoint and a revision template —
and almost every operational mistake comes from conflating them. Endpoint
decisions (ingress, invoker policy, traffic split) take effect in place;
template decisions (containers, scaling, networking) only ever take
effect by stamping out a NEW immutable revision. Know which side of the
line a change lives on and deploys stop being surprising.

## The auth split is one decision, three dials

`allowUnauthenticated` grants public access THROUGH the IAM check (an
`allUsers` invoker grant); `invokerIamDisabled` turns the check OFF
entirely (for org policies that forbid `allUsers` grants); `iapEnabled`
puts Google's login wall in front of everything. The spec refuses the
first two together, and IAP with public access is a contradiction in
intent — behind IAP, let IAP be the front door. Services behind the
composed load balancer should also set ingress to
`INTERNAL_LOAD_BALANCER` and consider `defaultUriDisabled: true`: the
default *.run.app URL is the second front door everyone forgets, and
disabling it is the difference between "we route through the LB" and
"traffic CANNOT bypass the LB".

## Sandboxes: untrusted code beside the service, not inside it

For agent workloads that execute model-generated code, mark one
container `sandboxLauncher: true` (the supervisor) and declare
`sandboxTemplates` — the images and startup shapes the supervisor may
launch on demand as isolated, short-lived containers. Sandboxes share
the service's declared volumes but carry only literal environment
variables: pass secrets through the supervisor, never into a sandbox.
`resourceManagerTags` binds tag values at creation for org-policy, IAM
conditions, and cost reporting; it is immutable, so a tag change is a
redeploy — plan it as one.

## Three ways a variable gets its value -- only two keep a secret

Every `env` entry takes exactly one of `value`, `valueFromSecret`, or
`secretValue`, and the choice decides who can read it:

- `value` is written into the revision template. Anyone who can view the
  service reads it in the console and the Admin API, and every past
  revision keeps it. Configuration only. On Planton a `$secret/...`
  reference here is refused before anything deploys, because the platform
  resolves references to plain values before the module runs -- the
  reference would land in the revision as the secret itself.
- `valueFromSecret` points at a Secret Manager secret you already own
  (cross-project with the full `projects/*/secrets/*` name). Rotation is
  Secret Manager's: with version `latest`, new instances pick up a new
  version without a deploy, and running instances keep the old one -- so
  for a while two values serve at once. The runtime identity's
  `secretAccessor` grant is yours to make.
- `secretValue` is for a secret Planton holds -- on Planton it takes only
  a `$secret/...` reference (outside Planton, the value itself): the
  component creates one Secret Manager secret for the variable, replicated
  only in the service's regions, grants the runtime identity access to
  that secret alone, and pins the variable to the exact version it stored.
  A new value is a new revision -- this is the revision-template line
  again: rotating a `secretValue` is a deploy, visible in the revision
  history, never a silent drift between instances. Changing the Planton
  secret does not redeploy anything by itself; every deploy (a push, a
  promotion, a rollback, a CLI deploy) resolves the reference again and
  carries the secret's current value. Destroying the service removes the
  copy.

Pick `secretValue` for anything referenced as `$secret/...`; pick
`valueFromSecret` when the secret's lifecycle belongs to someone else
(another team's rotation, a secret shared across services). Give the
service a dedicated `serviceAccount` either way: without one the grant
goes to the project's Compute Engine default account, which every other
default-identity workload in the project shares.

## Probes: three types, three jobs, three shapes

Startup gates first traffic (HTTP/TCP/gRPC — its whole window is capped
at 240 seconds: `failureThreshold × periodSeconds`); liveness restarts a
broken instance (HTTP/gRPC only — the spec has no TCP arm because the
API rejects it, and timing may stretch to 3600s for expensive checks);
readiness pulls an instance from serving WITHOUT restarting it and
re-admits it as soon as a probe succeeds again. There is no
success-threshold dial: the GA API's probe message carries no such
field (the Terraform provider models one, but the control plane
silently drops it — re-admission is single-success by design). Reach
for readiness when the failure is recoverable (a saturated downstream,
a cache warm) — restarting those instances (liveness) makes the
incident worse.
Never point liveness at an endpoint that calls your dependencies: one
slow downstream becomes a fleet-wide restart storm.

## The cost model lives in three fields

`resources.cpuIdle` (default true) is request-based billing — CPU only
while serving; set it false and every instance bills for its lifetime,
which is what real background work needs and what idle services waste.
`scaling.minInstanceCount` buys away cold starts at idle cost.
`serviceScaling.maxInstanceCount` is the service-WIDE bill cap across
revisions — during a rollout two revisions serve at once, and
per-revision caps double the worst case exactly when finance is
watching. MANUAL mode pins the fleet regardless of traffic: a load-test
lever and an emergency brake, not an operating mode.

## Deploy from source is a build-system decision

`buildConfig` makes GCP the build system (the Cloud Run functions path):
source in a GCS archive, Cloud Build produces the image, optionally in a
private worker pool under a dedicated build identity. Its quiet payoff
is `enableAutomaticUpdates` + the containers' `baseImageUri`: Google
patches OS/runtime layers of the running image with no redeploy. Teams
with a CI/CD pipeline should keep shipping images — mixing pipeline
images and source deploys on one service makes "what is running" a
research project.

## Multi-region is a different resource wearing the same kind

`multiRegionSettings` with `region: global` deploys one service identity
into several regions at once — but the deeper choice is between that and
the composition this catalog already proves: one service per region
behind a global load balancer (serverless NEGs). The composed path gives
per-region rollout control, per-region traffic evidence, and custom
domains; multi-region gives one object to manage. Prefer composition for
production estates; multi-region for fleets of identical internal
services where per-region ceremony is pure overhead.

## Destroy stance

`deletionProtection` (default true) is the API-side guard — the delete
itself fails until the manifest opts out. `deletionPolicy` layers the
engine-side stance under it: PREVENT fails any destroying plan before it
starts; ABANDON drops the service from state and leaves it serving —
break-glass for state surgery, not an operating mode. Deleting a service
tears down its endpoint AND every revision; the URL is gone the moment
the destroy lands.

## On the diagram

The service consumes `GcpProject`, `GcpServiceAccount` (runtime
identity), `GcpKmsKey` (image CMEK), `GcpCloudSql` (socket volumes),
`GcpGcsBucket` (FUSE volumes), and `GcpVpcNetwork`/`GcpSubnetwork`
(direct VPC egress). Its `service_name` output is the handle
`GcpRegionNetworkEndpointGroup` bridges into the load-balancer family —
the custom-domain path — and `url` is what application configs and
Pub/Sub push subscriptions consume.

## Pairs well with

- `GcpRegionNetworkEndpointGroup` — the bridge into the HTTPS load
  balancer; the production custom-domain path.
- `GcpServiceAccount` — a dedicated least-privilege runtime identity;
  the Compute default SA is for experiments only.
- `GcpCloudSql` — managed Unix-socket volumes, no proxy sidecar needed.
- `GcpServerlessVpcConnector` — only when an org constraint forbids
  direct VPC egress; otherwise skip the connector entirely.
- `GcpCloudRunJob` — the run-to-completion sibling for batch work; do
  not run batch inside a service with `cpuIdle: false`.
