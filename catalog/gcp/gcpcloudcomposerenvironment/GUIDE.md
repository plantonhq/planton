# GcpCloudComposerEnvironment Guide

The judgment this guide protects: a Composer environment is a
25-45-minute build whose networking is immutable — every CIDR, network
attachment, and encryption decision must be right BEFORE the first
apply, because the fix for any of them is a full recreate of the
slowest resource in the catalog.

## Choose the generation through the image version

`softwareConfig.imageVersion` decides which fields mean anything:
Composer 2 environments network through `nodeConfig.network`/
`subnetwork` + `privateEnvironmentConfig` (VPC peering or PSC);
Composer 3 environments use `composerNetworkAttachment` +
`enablePrivateEnvironment` and unlock the DAG processor sizing,
`webServerPluginsMode`, and metadata retention. Mixing generations'
fields is the most common first-apply rejection — pick the generation,
then configure only its surface.

## Plan CIDRs like they are permanent, because they are

The master range, Cloud SQL range, Composer-internal /20, and the GKE
pod/services ranges must not overlap anything in the wider network —
and none of them can be changed later. Reserve ranges with the network
team before the first apply. The per-range XOR in
`ipAllocationPolicy` (named secondary range XOR CIDR to carve) is
enforced pre-deploy.

## Size with environment_size first, workloads_config second

`environmentSize` sets the managed infrastructure class (GKE + metadata
database); `workloadsConfig` then tunes per-component CPU/memory/storage
within it — scheduler, workers (min/max for autoscaling), web server,
triggerer (all three fields or none), and on Composer 3 the DAG
processor. Both update in place: start SMALL, measure DAG parse and
task latency, and grow the component that is actually starved.

## The service account is a create-time contract

Composer 3 REJECTS creation without an explicit workloads service
account holding `roles/composer.worker` — the fall-back-to-default
behavior is legacy Composer 2 only. Wire a dedicated
`GcpServiceAccount` reference; granting its roles after a failed
create does not resume the create.

## Secrets are pointers, not values

`envVariables` is plain text that every Airflow component and anyone reading the environment can see. Put secrets in `secretEnvVariables`. The module stores each value in its own Secret Manager secret in `region`, grants `nodeConfig.serviceAccount` (or the default compute account when it is empty) `secretmanager.secretAccessor` on that secret alone, and sets an env var of the same name to the version's resource name (`projects/<p>/secrets/<id>/versions/<n>`). DAGs read the value from Secret Manager with that name. With secrets, `nodeConfig.serviceAccount` must be an email, and the two maps cannot share a key.

## Operational surfaces worth setting on day one

A `maintenanceWindow` (12h+ recurring) keeps Google's maintenance out
of your DAG peak; `recoveryConfig` scheduled snapshots are the DR
story for the metadata database; `dataRetentionConfig` keeps task logs
and Airflow metadata from growing unbounded on long-lived
environments. All three update in place but default to off.

## Destroy semantics

Deletion takes 10-15 minutes and the auto-created DAG bucket SURVIVES
— sweep it explicitly when the environment is truly gone.
`deletionPolicy: PREVENT` fits the environment a data platform runs
on; `ABANDON` keeps a meaningful idle bill running outside management,
so use it only for deliberate handovers.

## What is deliberately absent

Composer 1.x fields (node sizing, zones, oauth scopes, python version,
web-server/database machine types) — the generation is deprecated by
Google; the provider documents each with a composer-1-only constraint.
DAG delivery is not IaC: DAGs land in the bucket
(`dag_gcs_prefix` output), and their credentials/config ride
`GcpCloudComposerUserWorkloadsSecret`/`ConfigMap`.
