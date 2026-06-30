# OpenStackNetworkPort Research Documentation

## Terraform Resource Analysis

**Resource**: `openstack_networking_port_v2`
**Provider**: `terraform-provider-openstack/openstack` v3.x

### Schema Analysis (26 attributes total, 10 selected for 80/20)

| Attribute | Selected | Rationale |
|-----------|----------|-----------|
| `network_id` | Yes | Core: every port belongs to exactly one network |
| `fixed_ip` | Yes | Core: IP address assignments from subnets |
| `security_group_ids` | Yes | Core: security policy enforcement |
| `no_security_groups` | Yes | Important complement to security_group_ids |
| `admin_state_up` | Yes | Administrative state, consistent pattern |
| `mac_address` | Yes | Specific MAC for bonding/licensing |
| `port_security_enabled` | Yes | Controls SG enforcement, consistent with Network |
| `description` | Yes | Human-readable, consistent with all components |
| `tags` | Yes | Filtering/organization, consistent pattern |
| `region` | Yes | Standard region override |
| `name` | Implicit | Derived from metadata.name |
| `allowed_address_pairs` | No | VRRP/HA-specific, niche |
| `extra_dhcp_option` | No | Very niche DHCP customization |
| `binding` | No | Admin/SR-IOV/advanced networking |
| `dns_name` | No | Requires DNS extension |
| `device_owner` | No | Set by OpenStack services, not users |
| `device_id` | No | Set by OpenStack services, not users |
| `no_fixed_ip` | No | Trunk ports not in scope |
| `tenant_id` | No | Admin-only (consistent exclusion) |
| `value_specs` | No | Low-level escape hatch (consistent exclusion) |
| `qos_policy_id` | No | QoS policy, niche |

### New Patterns Introduced

#### 1. Repeated StringValueOrRef (`security_group_ids`)

This is the first Planton component to use `repeated StringValueOrRef`. Each element in the list independently resolves as either a literal UUID or a `value_from` reference. This enables InfraChart DAG wiring to multiple security groups created in the same chart.

**Proto pattern:**
```protobuf
repeated dev.planton.shared.foreignkey.v1.StringValueOrRef security_group_ids = 3 [
  (dev.planton.shared.foreignkey.v1.default_kind) = OpenStackSecurityGroup,
  (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.security_group_id"
];
```

**IaC handling (Pulumi):** Loop over the slice and call `GetValue()` on each element.
**IaC handling (Terraform):** List comprehension: `[for sg in var.spec.security_group_ids : sg.value]`

#### 2. StringValueOrRef Inside Nested Message (`FixedIp.subnet_id`)

This is the first component with FK annotations inside a nested/repeated message. The `subnet_id` field within the `FixedIp` message carries `default_kind` and `default_kind_field_path` annotations, enabling `value_from` references to subnets.

**Why not plain string?** Unlike SecurityGroup inline rules (where a standalone SecurityGroupRule component exists), there is no standalone "fixed IP assignment" component. This is the only path for subnet FK resolution on ports.

### Pulumi SDK Mapping

| Spec Field | Pulumi Type | Notes |
|------------|-------------|-------|
| `network_id` | `pulumi.StringInput` | Required |
| `fixed_ips` | `PortFixedIpArrayInput` | Nested with SubnetId + IpAddress |
| `security_group_ids` | `pulumi.StringArrayInput` | Resolved from repeated FK |
| `no_security_groups` | `pulumi.BoolPtrInput` | |
| `admin_state_up` | `pulumi.BoolPtrInput` | |
| `mac_address` | `pulumi.StringPtrInput` | ForceNew |
| `port_security_enabled` | `pulumi.BoolPtrInput` | |
| `description` | `pulumi.StringPtrInput` | |
| `tags` | `pulumi.StringArrayInput` | |
| `region` | `pulumi.StringPtrInput` | |

### Terraform Resource Mapping

| Spec Field | TF Attribute | Notes |
|------------|-------------|-------|
| `network_id` | `network_id` | Direct mapping via locals |
| `fixed_ips` | `dynamic "fixed_ip"` | Each entry becomes a block |
| `security_group_ids` | `security_group_ids` | Resolved to set via locals |
| `no_security_groups` | `no_security_groups` | Direct passthrough |
| `admin_state_up` | `admin_state_up` | Direct |
| `mac_address` | `mac_address` | Conditional null |
| `port_security_enabled` | `port_security_enabled` | Direct, nullable |
| `description` | `description` | Conditional null |
| `tags` | `tags` | Converted to set |
| `region` | `region` | Conditional null |

### OpenStack Behavior Notes

- **Default security group**: When neither `security_group_ids` nor `no_security_groups` is set, OpenStack applies the project's default security group automatically. This is Neutron's built-in behavior.
- **Port security inheritance**: If `port_security_enabled` is not set on the port, it inherits from the network's `port_security_enabled` setting.
- **MAC address**: Auto-generated if not specified. Once set (either auto or explicit), it cannot be changed without recreating the port (ForceNew).
- **Fixed IPs and DHCP**: Fixed IPs are assigned from the subnet's allocation pool. If DHCP is enabled on the subnet, the fixed IP is served via DHCP to the attached instance.
