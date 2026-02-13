# AzureNetworkSecurityGroup Pulumi Module -- Architecture Overview

## Resource Flow

```
Stack Input (AzureNetworkSecurityGroupStackInput)
  │
  ├── target: AzureNetworkSecurityGroup (api + spec + metadata)
  └── provider_config: AzureProviderConfig (credentials)
        │
        ▼
  initializeLocals()
  ├── Extracts resource group name via .GetValue()
  ├── Builds Azure tags from metadata
  └── Returns Locals struct
        │
        ▼
  Resources()
  ├── Creates Azure provider (auth via service principal)
  ├── Creates network.NetworkSecurityGroup (NSG shell)
  │   ├── Name, Location, ResourceGroupName, Tags
  │   └── No inline rules (rules created as separate resources)
  ├── For each spec.SecurityRules[i]:
  │   └── Creates network.NetworkSecurityRule
  │       ├── Name, Priority, Direction, Access, Protocol
  │       ├── SourcePortRange (default "*")
  │       ├── DestinationPortRange (required)
  │       ├── Source address: plural prefixes override singular
  │       ├── Destination address: plural prefixes override singular
  │       ├── Description (optional)
  │       └── DependsOn: NSG resource
  └── Exports outputs:
      ├── nsg_id (ARM resource ID)
      └── nsg_name
```

## Design Decisions

### Separate Rules (Not Inline)

Security rules are created as separate `network.NetworkSecurityRule` resources
rather than inline `SecurityRules` on the NSG. This approach:

1. Provides per-rule error messages during deployment
2. Gives each rule its own lifecycle in Pulumi state
3. Follows the pattern established by AzureUserAssignedIdentity (separate role assignments)
4. Avoids the Terraform "inline vs separate" conflict issue

### Address Prefix Precedence

When both singular (`source_address_prefix`) and plural (`source_address_prefixes`)
are available, the plural field takes precedence if non-empty. This avoids
complex CEL cross-field validation in the proto spec while supporting both
simple (single CIDR) and advanced (multi-CIDR) use cases.

### Default Handling

Fields with OpenMCF defaults (`source_port_range` = "*", `source_address_prefix` = "*",
`destination_address_prefix` = "*") are resolved by OpenMCF middleware before
the Pulumi module runs. The module uses `rule.GetXxx()` to access the resolved values.
