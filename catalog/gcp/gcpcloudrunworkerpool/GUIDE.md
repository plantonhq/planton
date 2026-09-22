# GcpCloudRunWorkerPool Guide

The judgment this guide protects: a worker pool is a Cloud Run service
with the front door removed. Everything about building the instance is
the same; everything about reaching it and scaling it is different.
Choose it when the process should be up all the time and pull its own
work.

## Three Cloud Run shapes

`GcpCloudRun` serves requests: it has a URL, a port, ingress, invoker IAM,
and scales with traffic. `GcpCloudRunJob` runs to completion: tasks start,
finish, and exit. This kind is the third shape -- a pool of instances
that never receive a request and never exit on their own: a Pub/Sub
puller, a scheduler loop, a stream processor, a GPU inference worker
reading from a queue. If the code answers HTTP, it is a service; if it
ends, it is a job; if it should run forever and find its own work, it is
a worker pool.

## Nothing reaches a worker pool

There is no port and no URL. A worker pool's containers open outbound
connections -- to a queue, a database, a cache -- and that is the whole
network story. Health probes therefore cannot fall back to "the serving
port": a startup or liveness probe must name the port a health listener
in the container binds, and the container must bind one. `vpcAccess`
gives the pool private reach into a VPC (direct VPC egress through
`networkInterfaces` is the recommended form; `egress: ALL_TRAFFIC` sends
public traffic that way too).

## You drive the scaling

A service scales with traffic; a worker pool has none. `MANUAL`
(Google's default) runs exactly `manualInstanceCount` instances --
including `0`, which parks the pool without deleting it, the cheapest
way to pause a worker. `AUTOMATIC` moves the count between
`minInstanceCount` and `maxInstanceCount` on a signal you supply, for
example a queue-depth metric wired through Cloud Monitoring. Because
CPU is always allocated on a worker pool, an idle instance bills like a
busy one: size the count for the work, and park what is not needed.

## Revisions and instance splits

Every template change -- a new image, an env var, a resource limit --
stamps out a new immutable revision, exactly as on a service. With
`instanceSplits` empty, every instance moves to the latest ready
revision. For a gradual rollout of a new worker image, split instances
by percentage between `INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST` and a named
`INSTANCE_SPLIT_ALLOCATION_TYPE_REVISION`; the percents must sum to 100,
which Google enforces at deploy.

## Two levers wait on the Pulumi SDK

The Pulumi SDK the catalog pins bridges an earlier provider than the
Terraform one, and its worker-pool types lack two things the Terraform
provider has: a container's `sandboxLauncher` flag and a probe's list of
HTTP headers (the SDK carries one header). Both are held out of the spec
-- the flag entirely, the header list capped at one -- so both engines
accept exactly the same manifests. They return the day the SDK carries
them (re-evaluated at pulumi-gcp v10 GA).

## What a destroy does

`deletionProtection` defaults to true: a destroy fails until the manifest
flips it, and both engines send the value explicitly. `deletionPolicy:
PREVENT` is a second, independent guard; `ABANDON` walks away from a
running (and billing) pool. `region`, `workerPoolName`, and a container's
`dependsOn` are the only immutables -- everything else rolls a new
revision in place.
