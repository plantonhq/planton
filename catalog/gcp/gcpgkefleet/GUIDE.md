# GcpGkeFleet Guide

The judgment this guide protects: the fleet must exist before anything joins it, it is one per project, and its defaults reach every cluster in it -- including clusters other teams register later.

## Declare it first

Google creates a project's fleet implicitly the moment the first cluster registers, whether through `GcpGkeCluster.fleetProject` or a `GcpGkeFleetMembership`. If this block is declared after that, its create collides with the fleet Google already made. So declare the fleet first, and point every scope, feature, and membership's `projectId` at its `project_id` output: the chart then orders them after it, and the E2E harness deploys it first because each child names it as a prerequisite.

## What the defaults reach

`defaultClusterConfig` applies to every cluster in the fleet, existing and future. `BASIC` security posture and vulnerability scanning are included with GKE; the `ENTERPRISE` modes belong to GKE's paid security capabilities. Binary Authorization `POLICY_BINDINGS` audits workloads against GKE platform policies, which are created through the Binary Authorization API -- distinct from the project policy `GcpBinaryAuthorizationPolicy` declares.

## Lifecycle

Everything but the project updates in place. Destroy deletes the fleet, and Google refuses while memberships or scopes remain, so tear the children down first (a chart does this in reverse order). `deletionPolicy: PREVENT` guards a fleet teams depend on.

## What is not offered yet

Fleet labels and the compliance posture default arrived in Google's provider 8.3, after the Pulumi SDK this catalog pins (pulumi-gcp v9.37.0). An argument one engine cannot send is never a one-engine field, so both wait for pulumi-gcp v10.
