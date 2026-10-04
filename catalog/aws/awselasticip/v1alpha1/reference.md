# AwsElasticIp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsElasticIpSpec defines the desired configuration for an AWS Elastic IP address.

An Elastic IP (EIP) is a static, public IPv4 address allocated from Amazon's
address pool (or optionally from a Bring-Your-Own-IP or IPAM pool). Unlike
ephemeral public IPs that change when an instance stops, an EIP persists until
explicitly released — making it the foundation for stable public endpoints.

Common consumers of Elastic IPs:
- Network Load Balancers (static IP per subnet via allocation_id)
- NAT Gateways (stable outbound IP for private subnets)
- EC2 instances (persistent public IP across stop/start cycles — attach with
  the `instance` or `network_interface` field below)

For the 95%+ use case, no spec fields are required — simply create the resource
to allocate a VPC EIP from Amazon's pool. The optional fields below support
attachment, Bring-Your-Own-IP (BYOIP), IPAM-managed pools, AWS Local/Wavelength
zone deployments, and reverse DNS.

Mutability: the allocation-shaping fields (pools, address, border group) are
ForceNew — changing any of them replaces the EIP (a NEW public address). The
association fields (instance, network_interface, associate_with_private_ip)
and reverse_dns_domain_name update in place without re-allocating.

AWS's legacy EC2-Classic ("standard" domain) addresses are retired — the
provider itself refuses them — so every EIP this component manages is a VPC
EIP; the modules pin domain = "vpc". Outposts customer-owned IP pools
(customer_owned_ipv4_pool) are excluded with the catalog's recorded Outposts
exclusion class.

Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticIp
metadata:
  name: static-egress-ip-demo
spec:
  region: us-west-2
  # Allocate from an IPAM public pool (omit to allocate from Amazon's pool).
  ipamPoolId: ipam-pool-07ccc86aa41bef7ce
  # Attach the address to an instance (at most one of instance /
  # networkInterface; AWS associates with exactly one target).
  instance:
    value: i-0123456789abcdef0
  # Reverse DNS (PTR) for the address -- AWS grants it only after a forward
  # A record for this domain already resolves to the EIP, so point DNS at
  # the address first, then set this on a follow-up apply.
  reverseDnsDomainName: mail.example.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.publicIpv4Pool` | `string` |  |  |  |
| `spec.address` | `string` |  |  |  |
| `spec.networkBorderGroup` | `string` |  |  |  |
| `spec.ipamPoolId` | `string` |  |  |  |
| `spec.instance` | `string \| valueFrom` |  |  | AwsEc2Instance (`status.outputs.instance_id`) |
| `spec.networkInterface` | `string \| valueFrom` |  |  | AwsEc2Instance (`status.outputs.primary_network_interface_id`) |
| `spec.associateWithPrivateIp` | `string` |  |  |  |
| `spec.reverseDnsDomainName` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.publicIpv4Pool

`string`

EC2 IPv4 address pool identifier to allocate from. Set this to a BYOIP pool
ID when you want the EIP to come from your own registered IP address range
instead of Amazon's pool. When omitted, Amazon allocates from its own pool.

This field is ForceNew: changing it requires replacing the EIP.

### spec.address

`string`

Request a specific IP address from the BYOIP pool identified by
`public_ipv4_pool`. The address must belong to the specified pool.
Only meaningful when `public_ipv4_pool` is set.

This field is ForceNew: changing it requires replacing the EIP.

### spec.networkBorderGroup

`string`

Network border group that controls the location scope of this EIP. Used for
allocating EIPs in AWS Local Zones or Wavelength zones. When omitted, the
EIP is scoped to the Region.

Example values: "us-east-1", "us-east-1-wl1-bos-wlz-1" (Wavelength),
"us-west-2-lax-1a" (Local Zone).

This field is ForceNew: changing it requires replacing the EIP.

### spec.ipamPoolId

`string`

Amazon VPC IP Address Manager (IPAM) pool to allocate this EIP from, e.g.
"ipam-pool-07ccc86aa41bef7ce". IPAM pools let network teams plan, track,
and audit public IPv4 usage centrally; an EIP allocated from a pool is
recorded against that pool's allocations. The pool must be a public-scope
pool provisioned for Elastic IP allocation in this region. May be combined
with `address` to recover a specific address the pool holds.

Takes a literal pool id today; when the platform's IPAM component lands,
reference its pool output instead.

This field is ForceNew: changing it requires replacing the EIP.

### spec.instance

`string | valueFrom`

EC2 instance to associate this EIP with. Reference an AwsEc2Instance's
instance_id output or pass a literal "i-..." id. The address follows the
instance across stop/start cycles until disassociated. At most one of
`instance` and `network_interface` may be set — associating with an
instance targets its primary network interface's primary private IP.
Updates in place: changing the target re-associates the same address
(the allocation is never replaced).

- references: AwsEc2Instance (`status.outputs.instance_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEc2Instance, name: <that resource's name>, fieldPath: status.outputs.instance_id}} -- a bare string does not parse

### spec.networkInterface

`string | valueFrom`

Network interface (ENI) to associate this EIP with — the precise form of
association: pick the exact ENI (and, with associate_with_private_ip, the
exact private address) the public IP maps to. Reference an
AwsEc2Instance's primary_network_interface_id output or pass a literal
"eni-..." id (standalone ENIs, appliance interfaces). At most one of
`instance` and `network_interface` may be set. Updates in place.

- references: AwsEc2Instance (`status.outputs.primary_network_interface_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEc2Instance, name: <that resource's name>, fieldPath: status.outputs.primary_network_interface_id}} -- a bare string does not parse

### spec.associateWithPrivateIp

`string`

The specific private IPv4 address on the association target that this
EIP maps to. Only meaningful when the target (instance or ENI) carries
multiple private addresses — omitted, AWS uses the primary private IP.
Requires `instance` or `network_interface` to be set. Updates in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipv4":true}}

### spec.reverseDnsDomainName

`string`

Fully qualified domain name to set as this EIP's reverse DNS (PTR)
record, e.g. "mail.example.com" — required by many mail providers before
they accept SMTP traffic from the address. AWS validates SERVER-SIDE that
a forward DNS record (A) for this domain already resolves to the EIP's
address BEFORE granting the PTR record, so create the EIP first, point
the domain at its public_ip output, then set this field on a follow-up
apply. Updates in place; clearing the field resets the PTR record.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\\.)+[a-zA-Z]{2,}\\.?$"}}

## Validation Rules

- `address_requires_pool`: address requires public_ipv4_pool or ipam_pool_id to be set (specific IPs can only be requested from a pool that holds them)
- `at_most_one_association_target`: at most one of instance and network_interface may be set (AWS associates an address with exactly one target)
- `private_ip_requires_association_target`: associate_with_private_ip requires instance or network_interface to be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsElasticIp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.allocation_id` | `string` | The allocation ID of the Elastic IP (e.g., "eipalloc-0123456789abcdef0"). This is the primary identifier used to reference the EIP in other AWS resources such as NLB subnet mappings, NAT Gateways, and EIP associations. |
| `status.outputs.public_ip` | `string` | The public IPv4 address assigned to this Elastic IP. |
| `status.outputs.arn` | `string` | The Amazon Resource Name (ARN) of the Elastic IP. Used for IAM policies and resource-level permissions. |
| `status.outputs.public_dns` | `string` | The public DNS hostname associated with the Elastic IP (e.g., "ec2-1-2-3-4.compute-1.amazonaws.com"). |
| `status.outputs.association_id` | `string` | The association ID (e.g., "eipassoc-...") when the spec attaches this EIP to an instance or network interface; empty for an unattached EIP. |
| `status.outputs.ptr_record` | `string` | The reverse DNS (PTR) record AWS granted for this address when spec.reverse_dns_domain_name is set; empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.instance` | AwsEc2Instance | `status.outputs.instance_id` |
| `spec.networkInterface` | AwsEc2Instance | `status.outputs.primary_network_interface_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsNatGateway | `spec.allocationId` | `status.outputs.allocation_id` |
| AwsNatGateway | `spec.secondaryAllocationIds` | `status.outputs.allocation_id` |
| AwsNatGateway | `spec.availabilityZoneAddresses[].allocationIds` | `status.outputs.allocation_id` |
| AwsNlb | `spec.subnetMappings[].allocationId` | `status.outputs.allocation_id` |
| AwsRedshiftCluster | `spec.elasticIp` | `status.outputs.public_ip` |

## See Also

- [Overview](../README.md)
