# AzurePrivateEndpoint Deployment Component (R08)

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework, Azure Provider

## Summary

Added AzurePrivateEndpoint (enum 414, id_prefix `azpe`) as a deployment component in the Azure provider, enabling private connectivity to Azure PaaS services through Azure Private Link. This is the 9th resource in the Azure expansion queue (R08) and a critical building block for the database-stack infra chart.

## Problem Statement / Motivation

Enterprise Azure architectures require private connectivity to PaaS services like PostgreSQL, MySQL, Key Vault, and Storage. Without Private Endpoints, traffic to these services traverses the public internet, even when both the client and service are in Azure. This creates security gaps and compliance issues in regulated industries.

The database-stack infra chart needs AzurePrivateEndpoint to complete its dependency chain:

```
VPC -> Subnet -> PrivateDnsZone -> Database Server -> PrivateEndpoint
```

### Pain Points

- Azure PaaS services expose public endpoints by default, creating unnecessary attack surface
- Private Link requires careful coordination between subnets, endpoints, DNS zones, and service connections
- The T02 spec design had 6 issues discovered during deep provider research that would have caused architectural problems if shipped uncorrected

## Solution / What's New

### Corrected Spec Design (6 Corrections from T02)

Deep research of the Terraform azurerm provider (`private_endpoint_resource.go`) and Pulumi Azure SDK (`privatelink.Endpoint`) revealed 6 corrections needed:

1. **Added `resource_group`** (StringValueOrRef) -- missing from T02 spec, required per DD05
2. **Added `region`** (string) -- private endpoints are regional, unlike Private DNS Zones which are global
3. **Changed `private_connection_resource_id` to StringValueOrRef** -- critical for infra-chart composability. Uses polymorphic pattern (no `default_kind`) since PE can connect to any Azure resource type
4. **Fixed output typing** -- replaced `string custom_dns_configs` with `string network_interface_id` (structured, useful)
5. **Hardcoded `is_manual_connection = false`** -- auto-approved is the 80/20 case
6. **Auto-derived internal names** -- connection name and DNS zone group name derived from `metadata.name`

### Component Structure

```
azureprivateendpoint/v1/
  spec.proto            -- 7 fields (3 StringValueOrRef, including polymorphic)
  stack_outputs.proto   -- 3 outputs
  api.proto             -- KRM wiring
  stack_input.proto     -- IaC module input
  spec_test.go          -- 18 tests (8 valid, 10 invalid)
  README.md             -- User-facing documentation
  examples.md           -- 6 YAML examples
  docs/README.md        -- Research documentation
  iac/
    hack/manifest.yaml  -- Test manifest
    pulumi/
      module/main.go    -- Endpoint + conditional DNS zone group
      module/locals.go  -- Tags and resource group extraction
      module/outputs.go -- Output constants
      main.go           -- Entrypoint
    tf/
      main.tf           -- azurerm_private_endpoint with dynamic dns_zone_group
      variables.tf      -- Typed variables
      outputs.tf        -- 3 outputs
      locals.tf         -- Tags and derived names
      provider.tf       -- azurerm ~> 4.0
```

## Implementation Details

### Polymorphic StringValueOrRef (Correction 3)

The `private_connection_resource_id` field is the second polymorphic StringValueOrRef in the Azure provider (after AzureUserAssignedIdentity's `scope`). It has no `default_kind` because a private endpoint can connect to any Azure service:

```protobuf
dev.planton.shared.foreignkey.v1.StringValueOrRef private_connection_resource_id = 5 [
  (buf.validate.field).required = true
];
```

This enables the database-stack infra chart:

```yaml
spec:
  privateConnectionResourceId:
    valueFrom:
      kind: AzurePostgresqlFlexibleServer
      name: prod-postgresql
      fieldPath: status.outputs.server_id
```

### Conditional DNS Zone Group

The DNS zone group is only created when `private_dns_zone_id` is provided:

```go
if spec.PrivateDnsZoneId != nil {
    dnsZoneGroupName := fmt.Sprintf("%s-dns-zone-group", locals.AzurePrivateEndpoint.Metadata.Name)
    endpointArgs.PrivateDnsZoneGroup = &privatelink.EndpointPrivateDnsZoneGroupArgs{...}
}
```

### Output Extraction

Private IP and NIC ID are computed by Azure and extracted using Pulumi's `ApplyT`:

```go
privateIpAddress := endpoint.PrivateServiceConnection.ApplyT(
    func(conn privatelink.EndpointPrivateServiceConnection) string {
        if conn.PrivateIpAddress != nil { return *conn.PrivateIpAddress }
        return ""
    }).(pulumi.StringOutput)
```

## Benefits

- **Infra chart ready**: All cross-resource references use StringValueOrRef, enabling database-stack composition
- **Correct from the start**: 6 spec corrections prevent technical debt that would require migration later
- **Production quality**: 18 validation tests covering all field combinations and edge cases
- **Feature parity**: Both Pulumi (Go) and Terraform (HCL) implementations with identical behavior
- **Polymorphic pattern**: Establishes reusable pattern for any-target-resource references

## Impact

- **Database stack**: AzurePrivateEndpoint completes the private connectivity chain for databases
- **Existing resources**: No changes to existing resources; this is purely additive
- **Enum registry**: AzurePrivateEndpoint = 414 fills the gap between AzurePublicIp (413) and AzurePrivateDnsZone (415)

## Related Work

- **DD03** (Composite Bundling Rules): Endpoint + DNS zone group bundled per established rules
- **DD05** (AzureResourceGroup First-Class): `resource_group` field follows the pattern
- **R07** (AzurePrivateDnsZone): Upstream dependency for DNS zone group registration
- **R05** (AzureSubnet): Upstream dependency for subnet allocation

---

**Status**: Production Ready
**Build**: `go build` passed
**Tests**: 18/18 passed
**Azure Provider Version**: ~> 4.0
**Pulumi Provider Version**: v6
