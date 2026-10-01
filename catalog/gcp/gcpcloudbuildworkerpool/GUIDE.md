# GcpCloudBuildWorkerPool Guide

The judgment this guide protects: a private pool is shared build infrastructure -- many triggers and Cloud Deploy targets point at one pool -- and its network posture is decided once, at creation.

## Choosing the network

A pool with no network block runs its workers on Google's service producer network with public egress: dedicated machines, nothing private. `networkConfig` peers the workers into one of your VPC networks; Google requires a Service Networking (private services access) connection on that network first, and the peered network's path must carry the project NUMBER, which both modules resolve from a `GcpVpcNetwork` reference. `privateServiceConnect` instead attaches each worker to a network attachment in the pool's region; `routeAllTraffic` sends everything through it (give the attachment's subnet Cloud NAT if builds still need the internet), otherwise only private ranges go through it. Either way, `noExternalIp` removes public addresses from the workers. All of this is immutable: changing the arm replaces the pool.

## Using the pool

Builds pick a pool by its full name, the `name` output. A trigger's inline build names it in `build.options.workerPool`; a build driven by `cloudbuild.yaml` names it in that file's `options.pool.name`, which is Google's current field; a Cloud Deploy target names it in `executionConfigs[].workerPool` so its render and deploy jobs run there. The caller needs `cloudbuild.workerpools.use` on the pool's project, and the trigger or target must be in the pool's region.

## Lifecycle

`workerConfig` (machine type, disk size, nested virtualization, public IPs), `displayName`, and `annotations` update in place. The pool has no standing charge, so an idle pool is free; a busy pool bills build minutes at its machine type's rate.
