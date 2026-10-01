# GcpGkeFleetMembership Guide

The judgment this guide protects: a cluster joins a fleet exactly once, either at creation through `GcpGkeCluster.fleetProject` or explicitly through this block -- never both.

## Which path

Prefer `fleetProject` for clusters the catalog creates: Google registers the cluster as part of creating it, and the cluster exports the membership's name as `fleet_membership`. Use this block for everything else: a cluster created without `fleetProject`, one created outside Planton, or one in another project joining a central fleet. Declaring this block for a cluster that already set `fleetProject` registers it twice and fails.

## Fleet Workload Identity

`issuer` makes Google trust the cluster's OIDC tokens within the fleet's workload identity pool (`{project}.hub.id.goog`), so workloads use the same identities across the fleet. For a GKE cluster the issuer is `https://container.googleapis.com/v1/` followed by the cluster's `cluster_id` output. Google requires the `locations/` form, which is why the cluster's `self_link` (`zones/` for zonal clusters) is not offered as a reference. Google refuses an issuer change in place, so changing it replaces the membership.

## Lifecycle

The ID, location, cluster, and issuer are create-time decisions; labels update in place. Replacing the membership drops the scope bindings and per-cluster feature settings that target it. Destroy unregisters the cluster and leaves the cluster itself untouched.
