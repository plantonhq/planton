# AlicloudNetworkLoadBalancer

**Date**: 2026-02-19
**Kind**: AlicloudNetworkLoadBalancer
**Enum**: 3026
**ID Prefix**: acnlb

## Summary

Added Alibaba Cloud Network Load Balancer (NLB) as a new cloud resource kind. NLB provides ultra-high performance Layer 4 (TCP/UDP/TCPSSL) load balancing for workloads that require low latency and high throughput.

## What's Included

- **Proto API**: spec, api, stack_input, stack_outputs with comprehensive buf.validate rules
- **Pulumi module**: Go implementation using `nlb.LoadBalancer`, `nlb.ServerGroup`, `nlb.Listener`
- **Terraform module**: HCL implementation using `alicloud_nlb_load_balancer`, `alicloud_nlb_server_group`, `alicloud_nlb_listener`
- **Validation tests**: 40 Ginkgo/Gomega tests covering all fields, ranges, and enums
- **Documentation**: README, examples (3 patterns), research docs, presets
- **Presets**: internet-tcp, internal-tcp-drain, tcpssl-production

## Key Design Choices

- Bundles NLB + Server Groups + Listeners (same pattern as ALB)
- Zone mappings support optional `allocation_id` for fixed EIP binding (unique to NLB vs ALB)
- Connection draining support on server groups (10-900s timeout)
- TCPSSL listeners with mutual TLS support
- Six scheduling algorithms (Wrr, Rr, Sch, Tch, Qch, Wlc)
- Proxy Protocol support for real client IP forwarding
- Cross-zone load balancing toggle
- Billing hardcoded to PayAsYouGo (matching ALB convention)
