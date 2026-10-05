# AzureNetworkWatcherFlowLog Terraform Module

## Overview

Creates a Network Watcher flow log -- the recorder that writes network traffic metadata for one target (virtual network, subnet, or network interface) into a storage account, optionally enriched by Traffic Analytics in a Log Analytics workspace.

## Resources Created

- `azurerm_network_watcher_flow_log` -- the flow log (a child of the region's Network Watcher, which Azure auto-creates -- this module never creates a watcher)

## Variables

The generated `variables.tf` mirrors the proto contract:

- `metadata` -- Planton resource metadata (name, org, env, labels, tags)
- `spec` -- the AzureNetworkWatcherFlowLogSpec fields; the target, storage account, and workspace references arrive as resolved literals

## Outputs

- `flow_log_id` -- the flow log's ARM resource ID
- `flow_log_name` -- the flow log's name
- `network_watcher_name` -- the Network Watcher the flow log attached to

## Usage

The module is executed by the Planton platform with a tfvars file converted from the manifest. To run it standalone, provide `metadata` and `spec` variables matching the generated `variables.tf`.

## Behavior Notes

- **The Network Watcher is referenced, never created**: unset watcher fields resolve in `locals.tf` to the auto-created regional singleton (`NetworkWatcher_{region}` in `NetworkWatcherRG`) -- Azure creates it the moment the region hosts a virtual network and allows exactly one per region per subscription.
- **The flow log's region must be the TARGET's region** (flow logging is regional) -- ARM rejects a cross-region pairing at apply time.
- **Creating a flow log writes a storage lifecycle-management rule that OVERWRITES existing rules** on the target storage account -- point flow logs at an account without hand-managed lifecycle policy.
- **NSG targets are rejected by validation**: Azure stopped accepting new NSG flow logs on 2025-06-30 (the resource class retires fully 2027-09-30) -- target the virtual network, subnet, or network interface instead.
- **`enabled` and schema `version` always send explicit values** (platform defaults true and 1); prefer version 2 for anything new -- it adds flow state and byte/packet counters and is what Traffic Analytics consumes best.
- **When `enabled` is false the provider suppresses retention read-back diffs** (Azure returns defaults for a paused flow log) -- expected, not drift.

## Required Permissions

The deploying principal's least-privilege action set lives in the kind's permissions manifest, `../permissions.yaml`.
