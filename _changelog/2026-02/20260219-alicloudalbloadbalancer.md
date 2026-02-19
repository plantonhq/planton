# AlicloudAlbLoadBalancer

**Date**: 2026-02-19
**Type**: New Component
**Resource Kind**: AlicloudAlbLoadBalancer (enum 3025, id_prefix: acalb)

## Summary

Added AlicloudAlbLoadBalancer component that provisions an Alibaba Cloud Application Load Balancer (ALB) with bundled server groups and listeners. This is a composite component (per DD07) that creates a fully functional L7 load balancer as a single deployable unit.

## What's Included

- **Proto API**: spec.proto with 7 message types, stack_outputs.proto with 3 outputs, api.proto, stack_input.proto
- **Validations**: CEL validations for address_type, load_balancer_edition, protocol, scheduler, health_check_protocol, health_check_method, listener_protocol, sticky_session_type; range constraints on ports, thresholds, timeouts; min_items=2 on zone_mappings
- **Tests**: spec_test.go with 11 valid-input and 17 invalid-input test cases (39 total)
- **Pulumi Module**: main.go, locals.go, outputs.go, server_groups.go, listeners.go -- server group name-to-ID resolution for listener default actions
- **Terraform Module**: main.tf, server_groups.tf, listeners.tf, variables.tf, outputs.tf, locals.tf, provider.tf -- for_each on server groups and listeners
- **Documentation**: catalog-page.md, examples.md, README.md, docs/README.md, Pulumi overview.md, TF README.md
- **Presets**: 3 presets (internet-http, https-production, internal-grpc)
- **Registration**: Enum 3025 in cloud_resource_kind.proto

## Design Decisions

- **ALB Rules excluded**: 9 action types x 9 condition types would triple the proto spec; listeners' default_actions cover the 80% case
- **Server groups created empty**: Backend membership managed externally (ACK ingress, SAE, manual), matching Azure LoadBalancer pattern
- **Billing hardcoded**: ALB only supports PayAsYouGo; not exposed in spec
- **Zone mappings min 2**: ALB requires multi-AZ for HA; enforced via proto validation
- **Server group name as listener reference**: Listeners reference server groups by name, resolved to IDs in IaC modules

## Dependencies

- AlicloudVpc (vpc_id)
- AlicloudVswitch (vswitch_id in zone_mappings)
