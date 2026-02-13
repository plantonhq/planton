# ScalewayPrivateNetwork Resource Kind (R02)

## Summary

Implemented `ScalewayPrivateNetwork` as the second Scaleway resource kind in OpenMCF. This is the first Scaleway resource to use `StringValueOrRef` for cross-resource dependency wiring, establishing the pattern that all subsequent Scaleway resources will follow.

## What Changed

### New Files (in `apis/org/openmcf/provider/scaleway/scalewayprivatenetwork/v1/`)

**Proto schemas (4):**
- `spec.proto` -- Spec with `StringValueOrRef vpc_id` referencing ScalewayVpc, region, optional ipv4_subnet/ipv6_subnets, enable_default_route_propagation
- `stack_input.proto` -- Standard target + provider_config pattern
- `stack_outputs.proto` -- Exports `private_network_id` (primary cross-resource reference) and `ipv4_subnet_cidr`
- `api.proto` -- Resource envelope with api_version, kind, metadata, spec, status

**Pulumi Go module (7):**
- Entrypoint loads stack input and calls module.Resources
- Module initializes locals (resolving `vpc_id` from `StringValueOrRef` via `GetValue()`), creates Scaleway provider, provisions Private Network
- Supports optional IPv4 subnet, optional IPv6 subnets, and enable_default_route_propagation
- Exports `private_network_id` and `ipv4_subnet_cidr` as stack outputs

**Terraform HCL module (5):**
- Uses `scaleway_vpc_private_network` resource with dynamic blocks for optional ipv4_subnet and ipv6_subnets
- Standard metadata, spec, and credential variables
- Exports `private_network_id` and `ipv4_subnet_cidr`

**Documentation (2):**
- `README.md` -- Covers universal connector role, how Scaleway networking hierarchy works, constraints, use cases
- `examples.md` -- 5 examples covering minimal dev, explicit subnet, valueFrom VPC reference, multi-tier architecture, and dual-stack IPv6

### Modified Files

- `pkg/crkreflect/kind_map_gen.go` -- ScalewayPrivateNetwork registered in the cloud resource kind map

## Why This Matters

Private Network is the "universal connector" in the Scaleway resource graph. 8 of the remaining 17 resource kinds depend on `private_network_id`:
- ScalewayPublicGateway, ScalewayLoadBalancer, ScalewayInstance, ScalewayKapsuleCluster, ScalewayRdbInstance, ScalewayRedisCluster, ScalewayMongodbInstance, ScalewayServerlessContainer

This implementation establishes the `StringValueOrRef` cross-resource reference pattern that all subsequent Scaleway resources will follow.

## Build Verification

- `go build` -- clean
- `go vet` -- clean
- `terraform validate` -- "Success! The configuration is valid."
- `make generate-cloud-resource-kind-map` -- ScalewayPrivateNetwork registered

## Branch

`feat/scaleway-cloud-provider`
