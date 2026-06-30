# AWS EC2 Instance Deployment: From ClickOps to Production-Ready IaC

## Introduction: The Infrastructure Evolution

There's a persistent myth in cloud computing: "Just launch an EC2 instance from the console—it's easy!" And it is easy. Click, click, click, and you've got a virtual machine running. But that's precisely the problem.

What's easy for a learning exercise becomes a liability in production. Manual console launches lack repeatability, version control, and peer review. They encourage security anti-patterns: default VPCs, overly permissive security groups, SSH keys scattered across laptops, and instances launched in public subnets when they should be private.

The modern cloud paradigm has shifted decisively toward Infrastructure-as-Code (IaC), where infrastructure is defined as declarative data structures, reviewed like application code, and deployed through automated pipelines. This shift wasn't arbitrary—it was driven by the hard lessons of managing complex systems at scale.

This document explores the landscape of EC2 instance deployment methods, from manual console workflows to sophisticated IaC abstractions. We'll examine why certain approaches have become best practices, what makes a production-grade EC2 configuration, and how Planton's AwsEc2Instance API embodies these principles in a clean, opinionated abstraction.

## The Deployment Landscape: A Maturity Spectrum

### Level 0: The Manual Console (The Learning Tool)

The AWS Management Console's "Launch an instance" wizard is the entry point for most newcomers. It's a multi-step form that guides you through selecting an Amazon Machine Image (AMI), choosing an instance type, configuring networking, adding storage, and setting security groups.

**What it solves:** It makes AWS approachable. New users can see their first EC2 instance running within minutes.

**What it doesn't solve:** Everything that matters in production.

- **No repeatability:** Launch ten instances, and you'll configure them ten different ways.
- **No version control:** There's no Git commit showing what changed and why.
- **No peer review:** No pull request approval before modifying production infrastructure.
- **Error-prone:** Common mistakes include using the default VPC, opening SSH (port 22) to 0.0.0.0/0, forgetting to attach an IAM role, or using invalid device names that cause launch failures.

**Verdict:** Acceptable for learning and one-off experimentation. Unsuitable for anything that will persist longer than a coffee break.

Interestingly, even AWS's own "simplified" tools like AWS Launch Wizard don't truly operate at the console level—they generate CloudFormation stacks under the hood. When these wizards fail, you see CloudFormation errors. This reveals a fundamental truth: CloudFormation (and IaC generally) is the *real* control plane for AWS. Manual deployment is just a convenience wrapper that sacrifices the core benefits of code-driven infrastructure.

### Level 1: Imperative Scripting (The Automation Trap)

This level includes the AWS CLI (`aws ec2 run-instances`) and SDKs like Boto3 (Python), the AWS SDK for Go, Java, or Node.js.

**What it solves:** You can now automate launches with scripts. You can put those scripts in version control.

**What it doesn't solve:** The dependency graph problem.

To launch an EC2 instance in a custom VPC, you don't just run one command. You must:

1. Create (or look up) the VPC ID
2. Create (or look up) subnet IDs
3. Create (or look up) security group IDs
4. Create (or look up) a key pair
5. *Finally*, call `run-instances` with all of those IDs

You, the engineer, must manually orchestrate this multi-step, ordered process. If a security group already exists but a subnet doesn't, your script needs conditional logic to handle that. This is the imperative programming model: you specify *how* to achieve the desired state, step by step.

**Example (Boto3):**

```python
import boto3

ec2 = boto3.resource('ec2')

# Assumes VPC, subnet, and security group already exist
instance = ec2.create_instances(
    ImageId='ami-0c55b159cbfafe1f0',
    InstanceType='t3.micro',
    MinCount=1,
    MaxCount=1,
    SubnetId='subnet-0123456789abcdef',
    SecurityGroupIds=['sg-0abcdef1234567890'],
    KeyName='my-key-pair'
)[0]

print(f"Launched instance: {instance.id}")
```

This code is imperative. It issues commands. It doesn't describe a desired end state; it executes a sequence of API calls.

**Verdict:** Useful for scripting and automation frameworks, but fundamentally limited. You're responsible for managing state, handling failures, and orchestrating dependencies. This is precisely what declarative IaC tools were invented to solve.

### Level 2: Configuration Management (The Post-Provisioning Layer)

Tools like Ansible, Chef, Puppet, and SaltStack are traditionally used for *configuring* operating systems—installing software, managing files, starting services. However, they also have modules to *provision* infrastructure.

For example, Ansible's `amazon.aws.ec2_instance` module can launch EC2 instances. This creates a workflow where you use a single tool to both provision the instance and then (in a subsequent task) configure the OS.

**What it solves:** Unified tooling for provisioning and configuration.

**What it doesn't solve:** The architectural tension between mutable and immutable infrastructure.

Modern best practice advocates for *immutable infrastructure*: servers should never be modified in place. Instead, configuration should be "baked" into a custom AMI using a tool like Packer (which can use Ansible as a provisioner). The IaC tool then deploys this pre-configured AMI. Changes aren't applied in-place; you deploy a new, updated AMI and replace the old instance.

This separation of concerns—IaC for provisioning, image builders for configuration—is cleaner and more aligned with modern cloud-native patterns.

**Verdict:** Ansible and similar tools have value for specific workflows, but mixing provisioning and post-launch configuration runs counter to the immutable infrastructure paradigm that dominates cloud-native thinking.

### Level 3: Declarative IaC (The Production Standard)

This is where the industry has converged. You define the *desired state* of your infrastructure, and the tool's engine calculates and executes the necessary actions (create, update, delete) to achieve that state.

The key tools in this space are:

- **Terraform (HCL):** Uses a proprietary, domain-specific language (HCL) to define resources like `aws_instance`. It manages state in an external state file, enabling drift detection and complex updates.
- **Pulumi (General-Purpose Languages):** Uses familiar languages like Python, TypeScript, or Go to define resources like `aws.ec2.Instance`. This allows programmatic logic (loops, classes, functions) while maintaining a declarative model.
- **AWS CloudFormation (YAML/JSON):** AWS's native IaC service. You define a "stack" of resources in YAML or JSON. The state is managed by AWS. The `AWS::EC2::Instance` resource is the low-level, canonical definition.
- **AWS CDK (L1 and L2 Constructs):** A framework that *generates* CloudFormation templates using familiar programming languages. It introduces two critical abstraction levels:
  - **L1 Constructs** (e.g., `CfnInstance`): 1:1 mappings to raw CloudFormation resources. Verbose and comprehensive.
  - **L2 Constructs** (e.g., `ec2.Instance`): High-level, opinionated abstractions that provide smart defaults, helper methods, and can implicitly create related resources (like IAM roles).

**Example (Terraform):**

```hcl
resource "aws_instance" "app_server" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.micro"
  subnet_id     = "subnet-0123456789abcdef"
  vpc_security_group_ids = [aws_security_group.web.id]

  tags = {
    Name = "AppServer"
  }
}
```

**Example (Pulumi - Python):**

```python
import pulumi
import pulumi_aws as aws

instance = aws.ec2.Instance("app-server",
    ami="ami-0c55b159cbfafe1f0",
    instance_type="t3.micro",
    subnet_id="subnet-0123456789abcdef",
    vpc_security_group_ids=["sg-0abcdef1234567890"],
    tags={"Name": "AppServer"}
)
```

**Example (AWS CDK - Python, L2):**

```python
from aws_cdk import aws_ec2 as ec2

instance = ec2.Instance(self, "AppServer",
    vpc=vpc,
    instance_type=ec2.InstanceType("t3.micro"),
    machine_image=ec2.MachineImage.latest_amazon_linux2023(),
    vpc_subnets=ec2.SubnetSelection(subnet_type=ec2.SubnetType.PUBLIC)
)

# CDK automatically creates and attaches a security group
instance.connections.allow_from_any_ipv4(ec2.Port.tcp(80))
```

Notice how the CDK L2 construct provides helper methods like `machine_image.latest_amazon_linux2023()` and `connections.allow_from_any_ipv4()`. These abstractions eliminate boilerplate and enforce best practices.

**What it solves:** The declarative model solves the dependency graph problem. You describe the desired end state, and the tool figures out the correct order of operations. State management enables drift detection, where the tool compares the code to the real-world infrastructure and identifies discrepancies.

**Verdict:** This is the production standard. The choice between Terraform, Pulumi, CloudFormation, and CDK comes down to language preference, team expertise, and multi-cloud requirements. All are production-proven.

## The Critical Insight: L1 vs. L2 Abstractions

The AWS Cloud Development Kit (CDK) introduced an explicit two-tier model that's profoundly important for understanding how to design IaC APIs:

- **L1 Constructs:** These are auto-generated, 1:1 mappings to CloudFormation resources. They expose *every* parameter of the underlying API. For EC2, that's hundreds of fields. They're comprehensive, verbose, and require you to explicitly define dependencies (like creating an IAM role, attaching policies, creating an instance profile, and then referencing it).

- **L2 Constructs:** These are hand-crafted, high-level abstractions. They provide:
  - **Smart defaults:** You don't have to specify every field.
  - **Helper methods:** Like `ssmSessionPermissions=true`, which automatically generates and attaches the required IAM role and policies.
  - **Implicit resource creation:** The construct can create and wire up related resources on your behalf.

This L1/L2 split is AWS's explicit admission: *the raw API is too complex for most users*.

**Planton's AwsEc2Instance follows the L2 philosophy.** It doesn't expose every EC2 parameter. It focuses on the 20% of configuration that covers 80% of use cases, provides secure-by-default settings, and abstracts complex multi-resource patterns into simple flags.

## Production Essentials: The Non-Negotiables

Deploying an EC2 instance "securely" isn't optional—it's the baseline. Here's what modern production standards demand:

### Secure-by-Default Networking

**The Rule:** EC2 instances in production *must* be placed in **private subnets**.

A well-architected AWS VPC follows a three-tier pattern:

- **Public subnets:** For internet-facing resources like load balancers and NAT Gateways.
- **Private subnets:** For application servers and databases. These instances cannot be reached directly from the internet. They access the internet (for software updates) through a NAT Gateway.
- **Isolated subnets:** For resources that should have no internet access at all.

**Anti-pattern:** Placing an EC2 instance in a public subnet with `associate_public_ip_address = true`. This exposes the instance directly to the internet, dramatically increasing the attack surface.

**Security Groups:**

Security groups are stateful, instance-level firewalls. The principle of least privilege is paramount:

- **Never** use `0.0.0.0/0` (all traffic) for sensitive ports like SSH (22) or RDP (3389).
- **Always** use source-specific rules. For example, a database security group should allow inbound traffic on port 3306 *only* from the security group ID of the application servers, not from an IP range.

### Identity and Access: IAM Instance Profiles

**The Rule:** Static IAM access keys (AKIA...) must *never* be stored on an EC2 instance.

The only secure mechanism for granting an instance AWS API permissions is an **IAM Instance Profile**, which attaches an IAM Role to the instance. This provides temporary, auto-rotated credentials via the instance metadata service.

This IAM role is the central hub for two critical functions:

1. **Management:** The instance needs the `AmazonSSMManagedInstanceCore` policy to be manageable via AWS Systems Manager Session Manager (see below).
2. **Application:** If your application needs to read from S3, fetch a database password from Secrets Manager, or publish metrics to CloudWatch, it uses this role with least-privilege permissions.

In modern cloud architecture, `iam_instance_profile` is not an "advanced" or "optional" field—it's an *essential, core component* of a secure instance.

### Modern Access Patterns: The Bastion-less Revolution

The most significant shift in cloud security over the past five years has been the elimination of bastion hosts and open SSH ports.

#### AWS Systems Manager Session Manager (Recommended)

**How it works:** The SSM Agent (pre-installed on most AWS AMIs) makes an *outbound* HTTPS connection to the SSM control plane. A user with the correct IAM permissions can then start a secure, browser-based or CLI-based shell session *through* this control plane tunnel.

**Why it's superior:**

- **Security:** Requires **no open inbound ports**. Not even port 22. This eliminates the primary attack surface.
- **Access Control:** Access is managed through IAM policies, not SSH keys. You grant `ssm:StartSession` permission to IAM users or roles.
- **Auditability:** Provides full session logging to CloudWatch Logs or S3, creating an immutable audit trail of who accessed what and what commands they ran.

**Prerequisites:**

- SSM Agent running on the instance (default on Amazon Linux, Ubuntu AMIs)
- An IAM instance profile with the `AmazonSSMManagedInstanceCore` managed policy

**Comparison Table:**

| Feature | AWS Systems Manager (SSM) | EC2 Instance Connect (EIC) | SSH + Bastion Host |
|---------|---------------------------|---------------------------|-------------------|
| **Inbound Port Required** | **None** (Outbound 443 only) | **TCP/22** (from EIC service IPs) | **TCP/22** (from user IPs) |
| **Key Management** | **None** (IAM-based) | **Temporary** (via IAM) | **Static** (Manual) |
| **Auditability** | **High** (Keystroke logging) | **Medium** (CloudTrail logs key push) | **Low** (OS-level logs only) |
| **Public IP on Instance** | **No** (Works in private subnets) | **No** (with EIC Endpoint) | **No** (but Bastion needs one) |
| **Core Prerequisite** | SSM Agent + IAM Role | EIC Agent + IAM Permission | key_name + SSHD |

**Verdict:** SSM is the modern standard. EC2 Instance Connect is a middle ground. Bastion hosts are obsolete for most use cases.

### AMI and Bootstrap Strategy

**Pre-Baked (Golden) AMIs:**

For production, don't rely on long-running user data scripts. Instead, use **Packer** to create a custom AMI:

1. Start with a base AMI (e.g., Amazon Linux 2023)
2. Apply all OS patches
3. Install required agents (CloudWatch, SSM)
4. Install application dependencies (e.g., nginx, Java runtime)
5. Save as a new, versioned AMI

This "immutable" approach dramatically speeds up boot times and reduces first-boot failures.

**User Data Scripts:**

Use these only for *lightweight* first-boot customization:

- **Idempotent:** Scripts must be safe to re-run.
- **Logged:** Always redirect output: `exec > /var/log/user-data.log 2>&1`
- **Secure:** Never embed secrets. Instead, fetch them from AWS Secrets Manager using the instance's IAM role:

```bash
#!/bin/bash
secret=$(aws secretsmanager get-secret-value --secret-id prod/db/password --query SecretString --output text)
```

### Storage: The gp3 Standard

**EBS Volume Types:**

- **gp3** (General Purpose SSD v3): This should be your default. It's ~20% cheaper than gp2 and provides 3,000 IOPS and 125 MB/s throughput *for free* at any size. You can provision additional IOPS or throughput independently of volume size.
- **gp2** (General Purpose SSD v2): Legacy. IOPS scaled with volume size (3 IOPS per GiB), which forced users to over-provision storage to get sufficient performance.
- **io2** (Provisioned IOPS SSD): For high-performance, I/O-intensive database workloads requiring sub-millisecond latency.

**Encryption:** All production EBS volumes should have `encrypted: true`. This should be a secure-by-default setting.

### Monitoring and Backup

**CloudWatch Metrics:**

- **Basic (Default):** Free, 5-minute intervals. Acceptable for dev/test.
- **Detailed (Paid):** 1-minute intervals. Essential for production to enable faster alarming and responsive auto-scaling.

**CloudWatch Agent:** Required for in-guest metrics like memory utilization and disk space. Not installed by default.

**Backup Strategy:**

Backups are not an inline property of an EC2 instance. Use **AWS Backup** or Amazon Data Lifecycle Manager (DLM):

- Configure a backup plan (e.g., "daily backups, retain for 30 days")
- Target resources by *tag* (e.g., `Backup: "Daily"`)

This tag-based targeting is why `tags` is an essential field in any EC2 API.

## The 80/20 of EC2 Configuration

Of the hundreds of EC2 parameters, which do users actually set?

### Essential (80% of Use Cases)

- `region`: The AWS region where the instance will be created (e.g., `us-west-2`)
- `ami_id`: The OS image
- `instance_type`: The VM size (e.g., `t3.micro`)
- `subnet_id`: Where to place the instance (forces a conscious networking decision)
- `security_group_ids`: What firewall rules to apply
- `tags`: Essential for cost allocation and automation

### Common (15% of Use Cases)

- `iam_instance_profile`: For SSM access or AWS API permissions
- `key_name`: For traditional SSH access
- `user_data`: For first-boot scripts
- `root_volume_size`: To customize root storage
- `ebs_optimized`: For I/O-intensive workloads
- `disable_api_termination`: Termination protection for critical instances

### Rare (5% of Use Cases)

- Additional EBS volumes (`ebs_block_device`)
- Detailed monitoring
- Burstable instance credit specifications
- Placement groups (for HPC)
- Dedicated hosts/tenancy
- Nitro Enclave options
- Instance metadata options (IMDSv2)

### Example Configurations

**Dev/Test Server:**

```protobuf
aws_ec2_instance {
  instanceName: "dev-server"
  amiId: "ami-0c55b159cbfafe1f0"  // Amazon Linux 2
  instanceType: "t3.small"
  subnetId: "subnet-public-us-east-1a"
  securityGroupIds: ["sg-dev-ssh-wide-open"]
  connectionMethod: BASTION
  keyName: "my-dev-key"
  rootVolumeSizeGb: 20
  tags: {"env": "dev", "owner": "dev-team"}
}
```

**Production App Server:**

```protobuf
aws_ec2_instance {
  instanceName: "prod-app-server"
  amiId: "ami-app-v1.2.3"  // Custom Packer AMI
  instanceType: "m5.large"
  subnetId: "subnet-private-app-1a"
  securityGroupIds: ["sg-app-alb-ingress", "sg-app-egress"]
  connectionMethod: SSM
  iamInstanceProfileArn: "arn:aws:iam::123456789012:instance-profile/app-ssm-role"
  rootVolumeSizeGb: 40
  ebsOptimized: true
  disableApiTermination: true
  tags: {"env": "prod", "app": "billing-api"}
}
```

**Production Database Server:**

```protobuf
aws_ec2_instance {
  instanceName: "prod-db-node-1"
  amiId: "ami-db-v1.0.0"  // Custom DB AMI
  instanceType: "r5.xlarge"  // Memory-optimized
  subnetId: "subnet-private-db-1b"
  securityGroupIds: ["sg-db-internal-access"]
  connectionMethod: SSM
  iamInstanceProfileArn: "arn:aws:iam::123456789012:instance-profile/db-ssm-role"
  rootVolumeSizeGb: 30
  ebsOptimized: true
  disableApiTermination: true
  tags: {"env": "prod", "role": "database", "Backup": "Daily"}
}
```

## Lifecycle Management: Immutable Infrastructure

### The Anti-Pattern: Mutable Infrastructure

The traditional model: an administrator logs into a server and runs `apt upgrade` or manually modifies configuration files. This is "mutable" infrastructure—servers that are modified in place.

**The problem:** The running server no longer matches its code definition. This is called "configuration drift." You can't reliably recreate the server, because the true state exists only in that one running instance, not in version-controlled code.

This anti-pattern contributed to the 2024 CrowdStrike outage, where a faulty, un-versioned, in-place update caused a global catastrophe.

### The Best Practice: Immutable Infrastructure

**The model:** Servers are *never* modified in place. Instead:

1. A change is required (e.g., security patch)
2. A *new* AMI is baked with Packer, containing the patch
3. The IaC code is updated to reference the new `ami_id`
4. The controller deploys a *new* instance with the new AMI
5. Traffic is routed to the new instance (Blue/Green deployment)
6. The *old* instance is terminated

**Benefits:**

- **Eliminates drift:** The running infrastructure always matches the code.
- **Atomic deployments:** Updates are all-or-nothing.
- **Instant rollbacks:** Route traffic back to the old instance.

**For Planton:** The controller's update strategy should follow the immutable pattern. A change to `ami_id` should trigger a "create-before-destroy" replacement.

### Vertical Scaling

Changing an instance's size (e.g., from `t3.micro` to `m5.large`) is a *disruptive operation*. The instance must be stopped, modified, and restarted. The Planton controller can automate this, but users should be aware this causes downtime.

## When to Use EC2 vs. Alternatives

### EC2 vs. Containers (ECS/EKS)

- **Use EC2 when:** You have a monolithic legacy application, require deep OS-level control, or are running software that isn't containerized.
- **Use ECS/EKS when:** Your application is microservices-based, containerized, and you need automated deployments, scaling, and service discovery.

### EC2 vs. Lambda

- **Use EC2 when:** You have a stateful application, a 24/7 process, or any task that runs longer than 15 minutes.
- **Use Lambda when:** Your logic is event-driven (e.g., "process this file upload" or "handle this API request") and completes quickly.

### Single EC2 vs. Auto Scaling Groups (ASG)

This is the most important comparison.

**The Problem:** A single EC2 instance is mortal. If it fails, it's gone. There's no automatic recovery.

**The Solution:** Deploy a single instance *inside* an Auto Scaling Group with `min=1, max=1, desired=1`.

**Why:** The ASG monitors the instance's health (via EC2 or ELB health checks). If the instance fails, the ASG automatically terminates the unhealthy instance and launches a healthy replacement. This provides *fault tolerance* for a single instance.

**Planton Opportunity:** The AwsEc2Instance API could include a `high_availability: true` flag. If set, the controller would transparently deploy an ASG of 1, abstracting this powerful best practice from the end user.

### EC2 vs. Lightsail

- **Use EC2 when:** You need full control, VPC integration, IAM roles, granular security, and the full range of instance types. This is the correct choice for all enterprise applications.
- **Use Lightsail when:** You want a simple, all-in-one VPS with a fixed monthly price for hobbyist projects that don't need to integrate with the broader AWS ecosystem.

## Planton's Choice: An L2 Abstraction

The AwsEc2Instance API embodies the lessons learned from the deployment landscape:

1. **Abstraction over raw APIs:** Like CDK L2 constructs, it focuses on the 20% of configuration that covers 80% of use cases, not 1:1 API mapping.

2. **Secure-by-default:** The default connection method is SSM (bastion-less), not SSH. The spec requires conscious choices about networking (subnet_id, security_group_ids), preventing default VPC anti-patterns.

3. **Opinionated guidance:** By requiring fields like `subnet_id` and `security_group_ids`, and providing an enum for `connection_method`, the API guides users toward production-ready patterns.

4. **Composability:** The API uses foreign key references (`StringValueOrRef`) to compose with other resources (VPCs, security groups, IAM roles), enabling clean separation of concerns.

5. **Validation:** CEL validators enforce rules like "SSM requires an IAM instance profile" and "SSH requires a key name," preventing misconfiguration at definition time.

## Conclusion: Infrastructure as First-Class Code

The evolution from manual console clicks to declarative, abstracted IaC isn't just a technical progression—it's a paradigm shift in how we think about infrastructure.

Infrastructure is no longer something you "set up" once and then manually maintain. It's code. It's versioned, reviewed, tested, and deployed through automated pipelines. It's immutable. When you need to change it, you don't modify it in place—you deploy a new version and replace the old one.

The AWS EC2 instance, one of the oldest and most fundamental AWS services, has matured from a "click and SSH" model to a secure, declarative, SSM-managed resource that's deployed, monitored, and replaced like any other piece of software.

Planton's AwsEc2Instance API brings this philosophy full circle: it provides a clean, L2-style abstraction that enforces best practices, minimizes cognitive load, and lets engineers focus on solving business problems rather than wrestling with hundreds of low-level API parameters.

The next time you need to deploy an EC2 instance, you won't be clicking through a console wizard. You'll be defining a protobuf spec, committing it to Git, and letting a controller reconcile reality to match your declared intent. That's the infrastructure-as-code future—and it's already here.

