# OpenStack Project Landing Zone

The **OpenStack Project Landing Zone** InfraChart bootstraps a new multi-tenant project on OpenStack --
a Keystone project with a role assignment, baseline networking (private network, subnet, router with
external gateway), a conservative security group, and a default SSH keypair.

Deploy this chart to onboard new teams. Each team gets a clean project with networking ready to go,
and can immediately start deploying compute, storage, and other resources inside their project.

---

## Included Cloud Resources

| Resource                             | Kind                          | Always Created | Notes                                          |
|--------------------------------------|-------------------------------|----------------|-------------------------------------------------|
| **Keystone Project**                 | `OpenStackProject`            | Yes            | New tenant project                              |
| **Role Assignment**                  | `OpenStackRoleAssignment`     | Yes            | Grants a role to the admin user on the project  |
| **SSH Keypair**                      | `OpenStackKeypair`            | Yes            | Default keypair for the project                 |
| **Private Network**                  | `OpenStackNetwork`            | Yes            | Baseline tenant network                         |
| **Subnet**                           | `OpenStackSubnet`             | Yes            | IP allocation for project resources             |
| **Router** (external gateway)        | `OpenStackRouter`             | Yes            | Outbound connectivity via external network      |
| **Security Group** (egress only)     | `OpenStackSecurityGroup`      | Yes            | Conservative baseline -- egress only, no ingress|
| **Router Interface**                 | `OpenStackRouterInterface`    | Yes            | Connects subnet to router                       |

### Security group philosophy

The baseline security group allows **all outbound traffic** but **no inbound traffic**. This is
intentionally conservative -- teams add their own ingress rules (SSH, HTTP, application ports) based
on their specific workload requirements. This prevents accidental exposure of services before the
team has explicitly configured access.

---

## Dependency Graph

Resources are deployed in topological order by the Infra Pipeline:

```
Layer 0 (parallel):  Project, Keypair, Network
Layer 1 (parallel):  RoleAssignment, Subnet
Layer 2 (parallel):  Router, SecurityGroup
Layer 3:             RouterInterface
```

The Role Assignment waits for the Project. The Subnet waits for the Network. The Router Interface
waits for both the Router and Subnet.

---

## Chart Input Values

| Parameter                 | Description                                          | Default              |
|---------------------------|------------------------------------------------------|----------------------|
| **project_description**   | Description for the new Keystone project             | `Managed by Planton` |
| **admin_role_id**         | Keystone role ID to assign                           | *(required)*         |
| **admin_user_id**         | Keystone user ID to grant the role to                | *(required)*         |
| **external_network_name** | Pre-existing external (provider) network name        | `public`             |
| **network_cidr**          | CIDR block for the project baseline subnet           | `10.0.0.0/24`        |
| **dns_nameservers**       | Comma-separated DNS server IPs                       | `8.8.8.8,8.8.4.4`   |
| **ssh_public_key**        | Default SSH public key for the project               | *(required)*         |

### Prerequisites

* An OpenStack environment with **Keystone** identity service and **Neutron** networking.
* **Admin or domain-scoped credentials** capable of creating projects and assigning roles.
* A pre-existing **external network** accessible to the new project.
* The **role ID** and **user ID** for the initial role assignment. You can find these using:

```bash
openstack role list        # find the role ID (e.g., "member" or "admin")
openstack user list        # find the user ID
```

---

## Customization

* **Add ingress rules**: Teams should add their own `OpenStackSecurityGroupRule` resources or create
  additional security groups for specific workloads (SSH, HTTP, application ports).
* **Multiple role assignments**: Duplicate the `OpenStackRoleAssignment` resource in
  `templates/identity.yaml` to grant roles to additional users or groups.
* **Custom network CIDR**: Override `network_cidr` to avoid collisions with other projects' networks
  if VPC peering or shared networking is planned.
* **Domain scoping**: Add `domainId` to the `OpenStackProject` spec if your OpenStack deployment
  uses multiple Keystone domains.

---

## Important Notes

* The `admin_role_id` and `admin_user_id` parameters are required. Without them, the role assignment
  resource will fail validation.
* The `ssh_public_key` is registered as a keypair in the new project, making it available for any
  compute resources created later.
* The external network must already exist and be accessible. It is not created by this chart.
* The baseline security group has **no ingress rules** by default. Teams must explicitly open ports
  for their workloads.
* This chart creates the project and networking foundation. Compute, storage, and application
  resources should be deployed separately (e.g., using the Developer Environment or Kubernetes
  Environment InfraCharts).

---

(c) 2026 Planton. All rights reserved.
