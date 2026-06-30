# AWS Client VPN: From Manual Clicks to Production Remote Access

## Introduction

For years, the conventional wisdom held that remote access to private cloud resources required bastion hosts—dedicated "jump boxes" that developers SSH'd into before connecting to internal services. These bastions became security bottlenecks, operational headaches, and single points of failure. They required patching, monitoring, user management, and careful firewall configuration. They represented a persistent administrative burden.

AWS Client VPN represents a paradigm shift: a fully-managed, elastic, OpenVPN-based service that eliminates the need for bastion hosts entirely. It provides secure TLS connections from end-user devices directly to resources within AWS VPCs and interconnected on-premises networks. The service provisions and manages all server-side infrastructure, scaling automatically to meet demand without any operational intervention.

But here's the critical insight that many platform teams discover too late: **an AWS Client VPN endpoint is not a single resource**. It is a composite service construct requiring at least three distinct, interdependent AWS resources to be functional: the Endpoint itself, Target Network Associations (linking the endpoint to VPC subnets), and Authorization Rules (defining access control policies). An endpoint created in isolation exists in a `pending-associate` state and cannot accept a single client connection.

This composite nature is the central challenge for any infrastructure automation effort. A high-level API—like the one Planton provides—must manage this collection of underlying resources as a single, atomic, declarative unit to be truly effective.

This document explores the landscape of AWS Client VPN deployment methods, from manual console operations to production-grade automation. It examines the critical architectural decisions—authentication methods, network topology, and availability models—and explains how Planton abstracts these complexities into a developer-friendly API that balances simplicity with production readiness.

## Understanding AWS Client VPN: Architecture and Strategic Fit

Before automating deployment, it's essential to understand what AWS Client VPN is, how it integrates with your VPC, and—most importantly—when *not* to use it.

### The Four Remote Access Solutions

A common failure point in platform design is choosing the wrong tool. AWS provides four distinct remote access solutions, each with a specific, non-overlapping primary use case. The choice depends on two factors: the *actor* (human vs. network) and the *access model* (network-level vs. resource-level).

| Service | Primary Use Case | Access Model | Connectivity | Key Differentiator |
|---------|------------------|--------------|--------------|-------------------|
| **AWS Client VPN** | Remote workforce, developers, contractors | User-to-Network | OpenVPN over Internet | Managed "dial-in" for roaming users |
| **AWS Site-to-Site VPN** | Connect branch office to VPC | Network-to-Network | IPsec over Internet/DX | Static, persistent tunnel for a fixed location |
| **AWS Direct Connect** | High-bandwidth hybrid cloud | Network-to-Network | Private, dedicated fiber | Bypasses the public internet; a physical link |
| **SSM Session Manager** | Operator/Admin access to EC2/RDS | User-to-Resource (via IAM) | HTTPS (via SSM Agent) | No network access needed; an IAM-based proxy |

**Critical distinction**: If your *only* use case is providing developers with shell access to EC2 instances or port forwarding to an RDS database for a local GUI client, **AWS Systems Manager Session Manager is the superior choice**. It's free, more secure (no network-level access, fine-grained IAM policies, full CloudTrail logging), and requires no VPN infrastructure. Use Client VPN only when you need true network-level access—for example, accessing multiple internal web dashboards, running network discovery tools, or connecting to applications using complex dynamic port ranges.

### Common Use Cases for Client VPN

The strategic fit of AWS Client VPN leads to three primary production scenarios:

1. **Developer Access to Private Resources**: The most common use case. Engineers connect from their laptops to access non-public resources like RDS databases, ElastiCache clusters, internal microservices on EKS, and internal admin dashboards—all without exposing these services to the internet.

2. **Remote Workforce and Contractor Connectivity**: The service elastically scales to support a distributed workforce accessing internal corporate applications. It's ideal for providing secure, granular, and time-bound access to third-party contractors or partners.

3. **Hybrid Cloud and Application Migration**: During a migration, Client VPN provides a unified access plane. Users connect once and, via proper routing, securely access applications that exist *both* in the AWS VPC and in on-premises networks (via a Site-to-Site VPN or Direct Connect link).

### VPC Integration: Subnets, ENIs, and Security Groups

The integration with a VPC is managed through **Target Network Associations**, not configured directly on the endpoint resource.

When you first create a Client VPN Endpoint, it enters a `pending-associate` state and is completely non-functional. To activate it, you must associate the endpoint with at least one subnet in the target VPC. This association performs two critical functions:

1. **Provisions Elastic Network Interfaces (ENIs)** within the specified subnet. These ENIs are managed by the Client VPN service.
2. **Adds the VPC's local route** to the endpoint's route table automatically, enabling routing to the VPC.

All traffic from VPN clients destined for the VPC egresses from these ENIs. Due to Source NAT, the source IP address as seen by resources inside the VPC will be the private IP of the Client VPN ENI, *not* the client's IP from the `client_cidr_block`.

**Security Group Integration**: When the first target network is associated, AWS automatically applies the VPC's *default* security group to the Client VPN ENIs. This is a common operational pitfall—the default security group is often not configured for VPN access patterns and may be overly permissive or overly restrictive. For any production deployment, you must create a **dedicated security group** for the Client VPN endpoint. This security group's egress rules define which protocols and ports clients can access. Correspondingly, target resources (RDS, EKS, etc.) must have ingress rules allowing traffic *from* the Client VPN's security group.

**High Availability Integration**: HA is not automatic—it's a user-managed configuration. To achieve resilience against an Availability Zone failure, associate the endpoint with multiple subnets, each in a different AZ. The Client VPN service will then manage ENIs across those AZs, automatically routing connections to healthy ENIs if one AZ fails.

## The Deployment Methods Spectrum

Deploying AWS Client VPN spans a spectrum from manual console operations to fully declarative platform abstractions.

### Level 0: Manual Console Deployment (The Learning Phase)

The AWS Management Console provides a step-by-step wizard suitable for initial testing but fundamentally not repeatable or scalable.

**The workflow**:
1. **Prerequisite**: Generate a server certificate and CA chain using `easy-rsa`, then import them into AWS Certificate Manager (ACM)
2. **Create Endpoint**: Specify the Client IPv4 CIDR (e.g., `10.0.0.0/22`), server certificate ARN from ACM, and authentication options
3. **Associate Target Network**: Link the endpoint to at least one VPC subnet (state transitions from `pending-associate` to `available`)
4. **Add Authorization Rule**: Without this, clients can connect but cannot access any resources—the #1 user-reported issue
5. **Download Configuration**: Export the `.ovpn` file for distribution to clients

**Common pitfalls in manual setup**:
- **Overlapping CIDR Blocks**: The `client_cidr_block` cannot overlap with the VPC's CIDR—a frequent and fatal configuration error
- **Missing Authorization Rule**: Clients authenticate successfully but can ping nothing because authorization fails
- **Security Group Blocking**: The default security group blocks egress traffic to desired resources
- **Certificate Errors**: Using expired or non-signed client certificates, or allowing the Certificate Revocation List (CRL) to expire

**Verdict**: Acceptable for learning and experimentation. Unacceptable for production environments requiring reproducibility, security, or multi-environment consistency.

### Level 1: Scripting (AWS CLI & Boto3)

Scripting with the AWS CLI or SDKs is the first step toward automation. This approach directly exposes the composite-resource nature of the service.

**AWS CLI sequence**:
```bash
# 1. Create the endpoint resource
aws ec2 create-client-vpn-endpoint \
  --client-cidr-block 10.0.0.0/22 \
  --server-certificate-arn arn:aws:acm:... \
  --authentication-options Type=certificate-authentication,MutualAuthentication={...}

# 2. Associate with a subnet
aws ec2 associate-client-vpn-target-network \
  --client-vpn-endpoint-id cvpn-endpoint-... \
  --subnet-id subnet-...

# 3. Add authorization rule
aws ec2 authorize-client-vpn-ingress \
  --client-vpn-endpoint-id cvpn-endpoint-... \
  --target-network-cidr 10.10.0.0/16 \
  --authorize-all-groups

# 4. (Optional) Add routes for full-tunnel
aws ec2 create-client-vpn-route \
  --client-vpn-endpoint-id cvpn-endpoint-... \
  --destination-cidr-block 0.0.0.0/0 \
  --target-vpc-subnet-id subnet-...
```

This multi-step, imperative process is the foundation that all declarative IaC tools must orchestrate behind the scenes.

**Boto3 (Python)**: Uses the same underlying API calls (`client.create_client_vpn_endpoint()`, `client.associate_client_vpn_target_network()`, etc.). Ideal for building advanced, event-driven automation, such as a Lambda function that performs custom logic on every client connection.

**Verdict**: Suitable for administrative scripts and one-off operations. Not ideal for declarative, state-managed infrastructure.

### Level 2: Configuration Management (Ansible)

Analysis of the `amazon.aws` and `community.aws` Ansible collections reveals a significant gap: **there is no dedicated, first-class module** for `aws_ec2_client_vpn_endpoint` or its associated resources.

While mature modules exist for AWS Site-to-Site VPN (`amazon.aws.ec2_vpc_vpn`) and VPC Endpoints (`amazon.aws.ec2_vpc_endpoint`), Client VPN users must resort to workarounds:
- Using the `amazon.aws.cloudformation` module to deploy a CloudFormation template
- Using `ansible.builtin.command` or `community.aws.aws_cli` to wrap imperative CLI commands
- Writing a custom Ansible module

**Verdict**: Due to this lack of first-class declarative support, Ansible is a poor choice for managing AWS Client VPN compared to modern IaC tools.

### Level 3: AWS-Native IaC (CloudFormation & CDK)

**CloudFormation** provides a complete, declarative model but exposes the granular, composite nature of the service directly to the developer. A template must define:
- `AWS::EC2::ClientVpnEndpoint`: The main endpoint
- `AWS::EC2::ClientVpnTargetNetworkAssociation`: A separate resource linking the endpoint to a subnet
- `AWS::EC2::ClientVpnAuthorizationRule`: A separate resource linking the endpoint to a destination CIDR
- `AWS::EC2::ClientVpnRoute`: A separate resource for adding routes

This is verbose and requires manually wiring all the dependencies, but it is fully declarative and idempotent.

**AWS Cloud Development Kit (CDK)** provides two levels of abstraction:

1. **L1 Construct (CfnClientVpnEndpoint)**: A direct, auto-generated 1:1 mapping to the CloudFormation resource. Using it is identical to writing CloudFormation, requiring manual creation and linking of all associated resources.

2. **L2 Construct (ec2.ClientVpnEndpoint)**: This is a high-level, human-designed abstraction and serves as a **primary model for Planton's API**. This construct's constructor accepts the VPC and target subnets directly, then *automatically creates and manages* the underlying `ClientVpnTargetNetworkAssociation` resources. It provides helper methods like `.addRoute()` and `.addAuthorizationRule()`, which handle the creation of dependent resources. **This L2 construct solves the composite resource problem**, presenting an atomic, logical `ClientVpn` object to the developer.

**Verdict**: CDK L2 is the gold standard for AWS-native teams. CloudFormation is suitable for those already invested in CFN. Neither is ideal for multi-cloud infrastructure management.

### Level 4: Third-Party IaC (Terraform, Pulumi, OpenTofu)

**Terraform** is the most common IaC tool for managing Client VPN. Like CloudFormation, it exposes granular, primitive resources:
- `aws_ec2_client_vpn_endpoint`
- `aws_ec2_client_vpn_network_association`
- `aws_ec2_client_vpn_authorization_rule`
- `aws_ec2_client_vpn_route`

The developer is responsible for linking these resources. Community modules (e.g., `terraform-aws-client-vpn-endpoint`) are widely used to bundle these primitives into a reusable, CDK-L2-style component.

**Production HA pattern in Terraform**:
```hcl
variable "production_subnet_ids" {
  description = "A list of subnet IDs, one in each AZ, for HA"
  type        = list(string)
  default     = ["subnet-0a...", "subnet-0b...", "subnet-0c..."]
}

resource "aws_ec2_client_vpn_network_association" "prod_ha" {
  for_each               = toset(var.production_subnet_ids)
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.main.id
  subnet_id              = each.value
}
```

**Pulumi** follows an identical model, providing granular resources in its `aws.ec2clientvpn` package:
- `aws.ec2clientvpn.Endpoint`
- `aws.ec2clientvpn.NetworkAssociation`
- `aws.ec2clientvpn.AuthorizationRule`
- `aws.ec2clientvpn.Route`

Because Pulumi uses general-purpose programming languages (TypeScript, Python, Go), creating multiple associations or rules often feels more natural than HCL's `for_each` meta-argument—you simply use a standard `for` loop.

**Verdict**: Terraform is the industry standard for multi-cloud, state-managed IaC. Pulumi is an excellent alternative for teams preferring general-purpose languages. Both are production-ready foundations.

### Level 5: Kubernetes-Native Abstraction (Crossplane)

Crossplane is an open-source Kubernetes add-on that enables platform teams to manage cloud infrastructure through the Kubernetes API using Custom Resource Definitions (CRDs).

The `provider-aws` defines granular CRDs for `ClientVpnEndpoint`, `NetworkAssociation`, etc., analogous to Terraform resources.

The key feature for platform engineering is **Composition**: Crossplane allows a platform engineer to define a new, high-level `CompositeResourceDefinition` (XRD) (e.g., `kind: XEC2ClientVpn`). A corresponding `Composition` object then *composes* this high-level abstraction from the granular, primitive resources. This provides a clean, Kubernetes-native API to application teams while hiding the complexity of the composite resource model.

This is the exact architectural pattern Planton implements, but with a **Protobuf API** as the high-level specification instead of a Kubernetes CRD.

**Verdict**: Crossplane is the gold standard for Kubernetes-native platform engineering. Its composition model validates the approach Planton takes with Protobuf APIs.

## Authentication Architectures: The Most Important Decision

The choice of authentication method is the most significant architectural decision, with profound impacts on user experience, administrative overhead, and security. AWS Client VPN supports three mutually exclusive methods.

### Option 1: Certificate-Based (Mutual TLS)

This is the simplest authentication method to set up, requiring no external dependencies. It's often used for developer or test environments.

**How it works**: Authentication is "mutual" (mTLS). The server presents its certificate (from ACM) to the client, and the client presents its unique client certificate to the server. Both must be signed by the same Certificate Authority (CA) chain.

**PKI generation workflow**:
```bash
# 1. Create the root CA
easyrsa build-ca

# 2. Create the server certificate and key
easyrsa build-server-full server

# 3. Create a unique certificate and key for each user
easyrsa build-client-full <username>
```

**Client-side distribution**: The admin downloads the base `.ovpn` file, *manually edits it* to paste the contents of the user's certificate and private key inside `<cert>` and `<key>` blocks, then securely distributes this modified file to the user.

**The critical weakness—lifecycle management**: When an employee leaves, their certificate must be revoked. This is a manual, multi-step, disruptive process:
1. `easyrsa revoke <username>`
2. `easyrsa gen-crl` (generates an updated Certificate Revocation List)
3. `aws ec2 import-client-vpn-client-certificate-revocation-list` (uploads the new CRL)

**Critical caveat**: Importing a new CRL immediately **resets all active client connections**. Additionally, the CRL file itself has an expiration date. If the CRL expires, **all new client connections will fail**—a critical failure mode that must be monitored.

**Use case**: Quick development and testing setups, small teams, temporary environments.

### Option 2: Active Directory (User-Based)

This method allows clients to authenticate using their existing corporate username and password, integrating with AWS Directory Service.

**How it works**: The client (using either the AWS Client or a generic OpenVPN client) is prompted for a username and password, validated against an Active Directory. This method does not use client certificates for authentication.

**Integration options**:
1. **AWS Managed Microsoft AD**: A fully managed AD domain hosted in AWS
2. **AD Connector**: A directory gateway proxy that forwards authentication requests from AWS to an *on-premises* Active Directory—the standard pattern for enterprises with existing on-premises user directories

**The primary benefit—granular access**: Authorization rules can be assigned based on AD group membership (using the group's SID). For example:
- Grant `Engineering_Group` (S-xxxxx14) access to `10.10.10.0/24` (Dev VPC)
- Grant `Finance_Group` (S-xxxxx15) access to `10.10.20.0/24` (Finance VPC)
- Grant `Admin_Group` (S-xxxxx16) access to `0.0.0.0/0` (Full Access)

**Significant limitation**: Client VPN authorization rules **do not support nested or recursive AD groups**. If User A is in Group-Devs, and Group-Devs is in Group-VPC-Access, an authorization rule for Group-VPC-Access will *not* apply to User A. This forces a "flat" group management strategy for VPN access.

**Use case**: Hybrid enterprises with existing on-premises Active Directory infrastructure.

### Option 3: SAML 2.0-Based (Federated Authentication)

This is the modern, enterprise-grade solution for authentication, enabling Single Sign-On (SSO) and strong Multi-Factor Authentication (MFA).

**How it works**: This method *requires* the use of the AWS-provided VPN Client (not generic OpenVPN clients):
1. The user initiates a connection
2. The AWS Client opens the system's default web browser, redirecting to the organization's Identity Provider (IdP) login page
3. The user authenticates (username, password, and MFA token)
4. The IdP federates with AWS via a pre-configured SAML assertion and redirects back to the client
5. The client receives the assertion and completes the VPN connection

**Integration options** (any SAML 2.0-compliant IdP):
- **AWS IAM Identity Center (formerly AWS SSO)**: The simplest, AWS-native IdP for managing users and groups
- **Amazon Cognito**: Can be used as a SAML IdP broker
- **Third-Party IdPs**: Okta, Azure AD, Ping Identity, Google Workspace, etc.

**The critical advantage—ops-free lifecycle management**: User provisioning and de-provisioning are handled entirely within the corporate IdP. When a user is deactivated in Okta or Azure AD, their VPN access is *instantly* revoked. There are no client certificates or CRLs to manage, no `.ovpn` files to distribute, and no manual revocation workflows.

**Use case**: Production environments for modern, cloud-native organizations. **This is the clear best practice for any enterprise deployment.**

### Recommendation Summary

| Method | Use Case | Client-Side Burden | Admin Burden (Lifecycle) | Best For... |
|--------|----------|-------------------|-------------------------|-------------|
| **Mutual TLS (Certs)** | Dev teams, test setups | High (installing certs) | **Very High** (PKI, CRLs, distribution) | Quick dev/test, simple setups |
| **Active Directory** | Enterprise w/ on-prem AD | Low (username/pass) | Low (uses existing AD) | Hybrid enterprises |
| **SAML (Federated)** | Modern, cloud-native orgs | Low (browser SSO) | **None** (uses existing IdP) | **Production/Enterprise (Default Choice)** |

For any production system, **SAML-based federated authentication is the clear best practice**. The operational simplicity of centralized IdP-based lifecycle management far outweighs the one-time setup complexity. Mutual TLS should only be used for temporary, non-production, or small-team test environments.

## Production Essentials: Beyond Basic Deployment

A functional dev endpoint and a resilient prod endpoint are two very different things. Production deployments require careful consideration of "day 2" operations: lifecycle, routing, availability, logging, and security.

### Split-Tunnel vs. Full-Tunnel: A Fundamental Topology Decision

The `split_tunnel` boolean parameter (which defaults to `false`) defines the entire network topology for the client.

**Full-Tunnel (split_tunnel = false)**:
- When a client connects, the Client VPN endpoint pushes a default route (`0.0.0.0/0`) to the client's device
- *All* traffic from the client—for the VPC, for on-prem, and for the public internet (e.g., Google, Slack)—is routed through the VPN tunnel
- **Use case**: High-security or compliance-driven environments where all client internet traffic must be routed through the VPC and inspected by AWS Network Firewall or security group rules
- **Costs & cons**: This is the most expensive and highest-latency option. All internet traffic incurs AWS Data Transfer Out costs and requires the VPC to have a NAT Gateway with a `0.0.0.0/0` route

**Split-Tunnel (split_tunnel = true)**:
- The endpoint pushes *only* the specific routes from its own route table (e.g., `10.10.0.0/16` for the VPC) to the client
- All other traffic (e.g., to the public internet) uses the client's local network interface
- **Use case**: **This is the 80% use case** and the recommended configuration for most deployments, especially for developers
- **Benefits**:
  - **Cost**: Dramatically reduces Data Transfer Out costs, as only private traffic traverses the VPN
  - **Performance**: Provides low-latency access to public internet resources by using the client's local connection

**Recommendation**: Default to `split_tunnel = true` for development and most production scenarios. Reserve full-tunnel for compliance-driven environments requiring deep packet inspection of all user traffic.

### High Availability: The Cost-Resiliency Trade-off

High availability is a user-managed configuration. The Client VPN *service* is managed and scalable, but the *endpoint* itself is only as resilient as its associations.

**The HA pattern**: A production endpoint *must* be associated with subnets in at least two, and preferably all, Availability Zones in its region.

**The cost-resiliency conflict**: This pattern creates a direct conflict with the pricing model. AWS Client VPN pricing has a fixed, 24/7 component for *each subnet association* (e.g., $0.10/hour/association):
- **Dev (Single-AZ)**: 1 association × $0.10/hr × 720 hrs/mo = **$72/month (fixed cost)**
- **Prod (3-AZ HA)**: 3 associations × $0.10/hr × 720 hrs/mo = **$216/month (fixed cost)**

This fixed cost is incurred **regardless of whether any users are connected**. Platform teams must budget for this fixed HA cost as the baseline for any production endpoint.

**Cost optimization strategy for non-prod**: For development and testing environments, associate the endpoint with *only one subnet*. This cuts the fixed association cost by 2/3 and is a reasonable trade-off for non-production workloads. An automation script (e.g., a scheduled Lambda) can disassociate all subnets at 7 PM and re-associate them at 7 AM to eliminate the fixed cost during off-hours.

### Observability: CloudWatch Logs

Connection logging is **disabled by default**. For any production environment, it is an essential, non-negotiable configuration.

**Configuration**: The `connection_log_options` block must be enabled on the endpoint, specifying a CloudWatch Logs Log Group and Log Stream.

**Value**: This is the *only* source of truth for troubleshooting connection failures. The logs capture:
- Successful and failed connection attempts
- The reason for failures (e.g., "TLS error," "user name and password errors," "client certificate expired")
- Connection initiation and termination events
- Bytes ingress/egress per connection

Without these logs, an administrator is blind to *why* a user's connection is failing (e.g., bad password vs. expired certificate vs. missing authorization rule).

### DNS Resolution for VPN Clients

By default, a client connecting to the VPN will continue to use its locally-configured DNS servers. This will fail when trying to resolve private hostnames (e.g., `db.internal.my-vpc`).

**VPC DNS Resolver**: The standard solution is to specify the VPC's built-in DNS resolver as a custom DNS server for the endpoint. The IP for this resolver is *always* the `.2` address of the VPC's primary CIDR block (e.g., `10.10.0.2` for a `10.10.0.0/16` VPC). This must be configured in the `dns_servers` parameter of the endpoint.

**Custom DNS**: The endpoint can also be pointed to other DNS servers, such as an on-premises resolver (accessible via an AD Connector or S2S VPN) or a Route 53 Resolver Inbound Endpoint.

**Split-Tunnel DNS caveat**: When custom DNS servers are set, the VPN client will send *all* DNS queries (for both private `*.internal` and public `youtube.com`) to the VPC's DNS resolver. The VPC resolver will resolve the public domain, but it may return an IP that is geographically suboptimal for the user. This is a known trade-off.

### Common Anti-Patterns (Summary)

- **CIDR Overlap**: The `client_cidr_block` overlaps with the VPC's CIDR or another target network. This is a fatal misconfiguration.
- **Missing Authorization**: The most common user error. The client connects and authenticates, but all traffic is dropped because no authorization rule exists.
- **Expired CRL**: Using Mutual TLS and allowing the Certificate Revocation List to expire. This blocks *all* new connections, effectively causing an outage.
- **Default Security Group**: Relying on the VPC's default security group, which is not designed for VPN access and violates the principle of least privilege.
- **Single-AZ Production**: Associating a production endpoint with only one subnet to save costs. This creates a single point of failure for all remote access if that AZ has an issue.

## What Planton Supports

Planton provides a unified, Protobuf-defined API for deploying AWS Client VPN endpoints that abstracts the composite resource model into a single, declarative specification.

### Design Philosophy: Hiding Complexity, Not Capability

The primary challenge with AWS Client VPN is its composite nature—requiring the orchestration of Endpoint, NetworkAssociation, and AuthorizationRule resources. The official Terraform provider exposes this complexity directly to the user, requiring manual linking of separate resources.

**Planton's approach**: Abstract the composite resource model entirely. The `AwsClientVpnSpec` message accepts parameters for the endpoint, the subnet associations, and the authorization rules in a single, high-level message. The framework's Pulumi controller is responsible for fanning out this single specification into the 3+ underlying AWS API calls, managing dependencies and ensuring atomic deployment.

This is the same pattern used by the AWS CDK L2 construct (`ec2.ClientVpnEndpoint`)—the gold standard for AWS-native abstractions.

### Current Implementation (v1)

The Planton AWS Client VPN API (see `spec.proto`) provides:

**Essential Fields (80% Case)**:
- `vpc_id`: The target VPC to attach the endpoint to
- `subnets`: A list of subnet IDs to associate (at least one required; multiple for HA)
- `client_cidr_block`: The IPv4 pool for connected clients (e.g., `10.0.0.0/22`)
- `server_certificate_arn`: The ARN of the ACM server certificate
- `authentication_type`: Currently supports `certificate` (Mutual TLS) in v1
- `cidr_authorization_rules`: A list of VPC CIDRs clients can access
- `disable_split_tunnel`: Boolean to control tunnel mode (defaults to `false`, enabling split-tunnel)

**Advanced Fields (20% Case)**:
- `vpn_port`: The port number (defaults to `443`)
- `transport_protocol`: `tcp` or `udp` (recommended: `tcp`)
- `log_group_name`: CloudWatch Logs group for connection logging
- `security_groups`: Custom security group IDs (defaults to VPC default if not set)
- `dns_servers`: Custom DNS server IPs (up to 2)

### Supported Authentication: Certificate-Based (v1)

The current v1 API supports **certificate-based (Mutual TLS) authentication only**. This is the simplest method to implement and is suitable for:
- Development and testing environments
- Small teams with manual user management
- Temporary or short-lived VPN setups

**Future roadmap**: Active Directory and SAML-based federated authentication will be added in a future API version (v2), enabling production-grade, enterprise-ready deployments with centralized IdP lifecycle management.

### Network Security and Subnet Associations

**Subnet associations** are specified as a list of subnet IDs in the `subnets` field. The controller creates a `NetworkAssociation` resource for each subnet, enabling multi-AZ HA deployments with a simple list.

**Security groups** can be optionally specified. If not provided, the VPC's default security group is applied. For production, a dedicated security group should be created with appropriate egress rules.

### Authorization Rules: Declarative Access Control

Authorization rules are specified as a list of IPv4 CIDRs in the `cidr_authorization_rules` field. For each CIDR, the controller creates an `AuthorizationRule` resource that grants access to all authenticated users.

**v1 limitation**: The current API creates "authorize all groups" rules. Fine-grained, group-based authorization (e.g., granting specific AD groups access to specific CIDRs) will be added in a future API version alongside Active Directory and SAML authentication support.

### DNS and Logging

**DNS resolution**: The `dns_servers` field accepts up to two IPv4 addresses. For most deployments, this should be set to the VPC's DNS resolver (the `.2` address of the VPC's primary CIDR block) to enable resolution of private VPC hostnames.

**Connection logging**: Setting the `log_group_name` field enables CloudWatch Logs integration. The controller automatically creates a log stream and configures the endpoint to send connection events to CloudWatch for troubleshooting and audit compliance.

### Multi-Environment Best Practice

Following AWS best practices, Planton encourages separate Client VPN endpoints for each environment:
- `company-dev-vpn` → Development VPC
- `company-staging-vpn` → Staging VPC
- `company-prod-vpn` → Production VPC

Each endpoint is associated with its respective VPC, providing complete isolation for network access, logging, and security policies.

### What Planton Abstracts Away

By using the Planton API, platform teams avoid:
- **Manual resource linking**: No need to manually create and link `NetworkAssociation` and `AuthorizationRule` resources
- **Dependency management**: The controller handles the proper sequencing of resource creation (endpoint → associations → rules)
- **State management**: Pulumi manages all underlying resource state, enabling drift detection and reconciliation
- **Error-prone configuration**: Built-in validation prevents common mistakes like CIDR overlaps, missing authorization rules, and invalid port/protocol pairings

## Conclusion: The Path to Production Remote Access

AWS Client VPN represents a fundamental shift from bastion hosts to managed, elastic remote access. But its composite resource model creates significant complexity for infrastructure automation. Teams that attempt to manually manage the endpoint, associations, and authorization rules with low-level Terraform or CloudFormation resources quickly encounter operational friction, especially when scaling to multi-environment, multi-region deployments.

Planton abstracts this complexity into a single, declarative API that makes the simple case simple (single-region, certificate-based access for developers) while making the advanced case possible (multi-AZ HA, CloudWatch logging, custom security groups).

The paradigm shift is clear: AWS Client VPN eliminates bastion host operational overhead; infrastructure as code eliminates deployment inconsistency. Together, they represent the modern path to secure, scalable remote access infrastructure.

**For production deployments**: Plan to adopt SAML-based federated authentication (coming in v2) to eliminate the administrative burden of certificate lifecycle management. Use split-tunnel mode to control costs. Deploy across multiple AZs for high availability. Enable CloudWatch logging for observability. And most importantly—use Planton's abstraction to manage the composite resource model as a single, atomic unit.

