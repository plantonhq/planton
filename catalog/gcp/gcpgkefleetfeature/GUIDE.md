# GcpGkeFleetFeature Guide

The judgment this guide protects: one block is one feature, its fleet defaults reach every cluster in the fleet including future ones, and creating it adopts whatever was already on.

## Which settings go where

`feature` names the feature, and the spec accepts only that feature's settings: `multiclusteringress`, `fleetobservability`, `clusterupgrade`, `rbacrolebindingactuation`, and `workloadidentity` are fleet-level settings of the features with those names; `fleetDefaultMemberConfig` holds the member defaults of `configmanagement`, `servicemesh` (`mesh`), and `policycontroller`. Google's API holds one spec per feature, so a block on another feature would never take effect -- the spec refuses it.

## Defaults and overrides

Google applies `fleetDefaultMemberConfig` to every cluster in the fleet, including clusters that join later -- the recommended way to configure Config Sync, Policy Controller, and Service Mesh. `membershipConfigs` overrides the default for named clusters; Config Sync's per-cluster form also offers `stopSyncing` (a break-glass pause) and `deploymentOverrides` (reconciler resources). Each override is an entry of the feature's map in Google, so it lives here, keyed by the cluster's membership name; the membership must be in the feature's fleet project.

## Lifecycle

Creating a feature that is already on adopts it: its settings are replaced by this block's. Destroy turns the feature off and removes the overrides, except `rbacrolebindingactuation`, which Google never deletes -- destroy empties its allowlist, and a custom role still used by a scope binding cannot leave the list. Both modules enable the Fleet API and the feature's own API and never disable them.

## What is not offered

Hierarchy Controller and Policy Controller through Config Management are turned down by Google (Config Sync 1.20 and 1.21); configuring them would block Config Sync upgrades or install nothing. Use the `policycontroller` feature for Policy Controller.
