# AlicloudNetworkLoadBalancer Pulumi Module

This Pulumi module provisions an Alibaba Cloud Network Load Balancer (NLB) with server groups and listeners.

## Usage

This module is invoked by the OpenMCF Pulumi runner. It reads a `AlicloudNetworkLoadBalancerStackInput` from the Pulumi stack config and creates all resources.

## Outputs

| Output | Description |
|---|---|
| `load_balancer_id` | The NLB instance ID |
| `dns_name` | DNS name assigned to the NLB |
| `server_group_ids` | Map of server group names to IDs |
