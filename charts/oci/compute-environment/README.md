# Compute Environment InfraChart

This chart provisions a **VM-based workload environment on Oracle Cloud Infrastructure**:

* Custom VCN with internet and NAT gateways
* Public subnet (load balancer) and private subnet (compute instances)
* Network security group allowing SSH and application traffic from the VCN
* Compute instance with configurable flex shape (OCPUs, memory)
* Optional application load balancer with HTTP health checks
* Optional block volume for persistent storage

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Virtual Cloud Network | `OciVcn` | Always |
| Public Subnet | `OciSubnet` | Always |
| Private Subnet | `OciSubnet` | Always |
| Compute NSG | `OciSecurityGroup` | Always |
| Compute Instance | `OciComputeInstance` | Always |
| Application Load Balancer | `OciApplicationLoadBalancer` | `enable_load_balancer` |
| Block Volume | `OciBlockVolume` | `enable_block_volume` |

## Parameters

| Name | Description | Default |
|------|-------------|---------|
| `compartment_ocid` | OCI compartment OCID | — |
| `vcn_cidr` | VCN CIDR block | `10.0.0.0/16` |
| `public_subnet_cidr` | Public subnet CIDR | `10.0.0.0/24` |
| `private_subnet_cidr` | Private subnet CIDR | `10.0.1.0/24` |
| `instance_name` | Instance display name | `app-server` |
| `availability_domain` | OCI availability domain | `US-ASHBURN-AD-1` |
| `shape` | Compute shape | `VM.Standard.E4.Flex` |
| `ocpus` | OCPUs (flex shapes) | `1` |
| `memory_gb` | Memory in GB (flex shapes) | `16` |
| `image_id` | Boot image OCID | — |
| `ssh_public_key` | SSH public key | — |
| `enable_load_balancer` | Create ALB | `false` |
| `lb_listener_port` | LB listener port | `443` |
| `lb_backend_port` | Backend port | `8080` |
| `enable_block_volume` | Create block volume | `false` |
| `block_volume_size_gb` | Volume size (GB) | `100` |
