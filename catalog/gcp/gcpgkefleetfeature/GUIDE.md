# GcpGkeFleetFeature Guide

The judgment this guide protects: one block is one feature, its fleet defaults reach every cluster in the fleet including future ones, and creating it adopts whatever was already on.

## Which settings go where

`feature` names the feature, and the spec accepts only that feature's settings: `multiclusteringress`, `fleetobservability`, `clusterupgrade`, `rbacrolebindingactuation`, and `workloadidentity` are fleet-level settings of the features with those names; `fleetDefaultMemberConfig` holds the member defaults of `configmanagement`, `servicemesh` (`mesh`), and `policycontroller`. Google's API holds one spec per feature, so a block on another feature would never take effect -- the spec refuses it.

## Defaults and overrides

Google applies `fleetDefaultMemberConfig` to every cluster in the fleet, including clusters that join later -- the recommended way to configure Config Sync, Policy Controller, and Service Mesh. `membershipConfigs` overrides the default for named clusters; Config Sync's per-cluster form also offers `stopSyncing` (a break-glass pause) and `deploymentOverrides` (reconciler resources). Each override is an entry of the feature's map in Google, so it lives here, keyed by the cluster's membership name; the membership must be in the feature's fleet project.

A per-cluster Config Sync override always says `enabled`: the provider sends an unset value as false, which removes Config Sync from that cluster, so the spec requires `true` or `false`. In the fleet-wide default the opposite holds -- `false` and `0` are never sent to Google, so `enabled: false` cannot keep Config Sync off (a source still installs it; declare no `configSync` instead), `preventDrift: false` leaves Google's setting, an audit interval of `0` leaves Google's interval, and an empty metrics `backends` list cannot turn export off.

## Lifecycle

Creating a feature that is already on adopts it: its settings are replaced by this block's. Settings the block does not declare survive the first apply and are removed on the next. Two blocks naming the same feature in one fleet fight over it, and destroying either turns the feature off for both -- give each feature one block. Destroy turns the feature off and removes the overrides, except `rbacrolebindingactuation`, which Google never deletes -- destroy empties its allowlist, and a custom role still used by a scope binding cannot leave the list. Both modules enable the Fleet API and the feature's own API and never disable them.

## What is not offered

Hierarchy Controller and Policy Controller through Config Management are turned down by Google (Config Sync 1.20 and 1.21); configuring them would block Config Sync upgrades or install nothing. Use the `policycontroller` feature for Policy Controller.
