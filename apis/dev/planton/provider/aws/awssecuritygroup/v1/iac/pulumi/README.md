# AWS Security Group

The **AwsSecurityGroup** API resource configures and manages an AWS EC2 Security Group for inbound and outbound traffic
control. It allows you to define ingress and egress rules via a straightforward YAML manifest, making it easier to
enforce least-privilege network policies within your VPC. This component is part of Planton’s multi-cloud
deployment framework, with support for Pulumi and Terraform under the hood.

By specifying `apiVersion`, `kind`, `metadata`, and `spec`, you can quickly validate and provision your AWS Security
Group across multiple environments—dev, staging, and production—while adhering to best practices for naming, rule
definition, and Protobuf-driven validations.

---

## Key Features

- **Simplified Security Group Creation**  
  Define a Security Group with inbound (`ingressRules`) and outbound (`egressRules`) traffic rules in one YAML file.

- **Fine-Grained Rule Control**  
  Configure rules for IPv4 or IPv6 CIDRs, reference other Security Groups for internal traffic, or allow
  self-referencing
  to handle traffic within the same group.

- **Pulumi & Terraform Support**  
  Deploy using either Pulumi or Terraform by leveraging the Planton CLI, which abstracts away the underlying
  differences, ensuring a consistent user experience.

- **Centralized Validation**  
  Validations defined in Protobuf (and enforced by Planton) ensure you catch potential misconfigurations (e.g.,
  missing VPC ID or invalid name formats) before deployment.

- **Customizable & Extensible**  
  Extend the specification with your organization’s default ingress/egress rules, or add more fields in a forked module
  if you have unique requirements.

---

## Installation

1. **Planton CLI**  
   Install the [Planton CLI](https://github.com/plantonhq/planton) to manage manifests, validate
   configurations, and orchestrate deployments with Pulumi or Terraform.

2. **AWS Credentials**  
   Configure AWS credentials so Planton can interact with AWS:
    - Environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, or
    - A Planton credential resource referencing your AWS account.

3. **Pulumi or Terraform**
    - [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) if using Pulumi.
    - [Terraform CLI](https://developer.hashicorp.com/terraform/downloads) if you prefer Terraform.

---

## Usage

1. **Define a Manifest**  
   Create a YAML file (e.g., `security-group.yaml`) that describes your AWS Security Group. For instance:

   ```yaml
   apiVersion: aws.planton.dev/v1
   kind: AwsSecurityGroup
   metadata:
     name: my-sg
     version:
       message: "Initial SG for web tier"
   spec:
     vpcId: vpc-012345abcdef
     description: "Allow inbound SSH and HTTP traffic"
     ingressRules:
       - protocol: "tcp"
         fromPort: 22
         toPort: 22
         ipv4Cidrs:
           - "192.168.0.0/16"
         description: "Allow SSH from internal network"
       - protocol: "tcp"
         fromPort: 80
         toPort: 80
         ipv4Cidrs:
           - "0.0.0.0/0"
         description: "Allow HTTP from anywhere"
     egressRules:
       - protocol: "-1"
         fromPort: 0
         toPort: 0
         ipv4Cidrs:
           - "0.0.0.0/0"
         description: "Allow all outbound"
   ```

2. **Validate the Manifest**  
   (Optional) Check for schema or field errors before deployment:
   ```bash
   planton validate --manifest security-group.yaml
   ```

3. **Deploy**  
   Use the CLI to provision via Pulumi or Terraform:
   ```bash
   # Pulumi
   planton pulumi up --manifest security-group.yaml --stack myorg/dev

   # Terraform
   planton terraform apply --manifest security-group.yaml --stack myorg/dev
   ```

4. **Verify**  
   Once provisioning completes, confirm the Security Group’s creation via AWS CLI or console:
   ```bash
   aws ec2 describe-security-groups --group-ids <group-id-returned>
   ```

---

## API Resource Specification

Below is a summary of the key fields from the **AwsSecurityGroup** resource. For full details, refer to the Protobuf
definitions in the Planton repository.

### `apiVersion`

- **Description**: Must be `"aws.planton.dev/v1"`.
- **Required**: Yes.

### `kind`

- **Description**: Must be `"AwsSecurityGroup"`.
- **Required**: Yes.

### `metadata`

- **Description**: Provides standard resource metadata.
- **Fields**:
    - **name**: Unique resource name (must be 3–63 characters, only alphanumeric plus `-` or `_` in the middle).
    - **version.message**: Mandatory string explaining version/change context.

### `spec.vpcId`

- **Type**: `string`
- **Required**: Yes
- **Description**: ID of the VPC in which this Security Group is created (e.g., `vpc-abcdef123`).

### `spec.description`

- **Type**: `string`
- **Required**: Yes (AWS requires a description)
- **Description**: A short explanation of the Security Group’s purpose. Maximum 255 characters.

### `spec.ingressRules`

- **Type**: `repeated SecurityGroupRule`
- **Description**: Inbound traffic rules. If empty, inbound is fully denied.
- **Fields** within each rule:
    - **protocol** (`string`, required): e.g., `tcp`, `udp`, `icmp`, or `-1` for all.
    - **fromPort** (`int32`): start of port range; `0` if not applicable.
    - **toPort** (`int32`): end of port range; `0` if not applicable.
    - **ipv4Cidrs** (`repeated string`): list of IPv4 CIDRs allowed.
    - **ipv6Cidrs** (`repeated string`): list of IPv6 CIDRs allowed.
    - **sourceSecurityGroupIds** (`repeated string`): reference other SGs as inbound source.
    - **selfReference** (`bool`): whether to allow traffic from this SG itself.
    - **description** (`string`): optional explanation of the rule (<= 255 chars).

### `spec.egressRules`

- **Type**: `repeated SecurityGroupRule`
- **Description**: Outbound traffic rules. If empty, AWS typically defaults to allow all.
- **Fields** within each rule:
    - **protocol** (`string`, required)
    - **fromPort** (`int32`)
    - **toPort** (`int32`)
    - **ipv4Cidrs** (`repeated string`)
    - **ipv6Cidrs** (`repeated string`)
    - **destinationSecurityGroupIds** (`repeated string`)
    - **selfReference** (`bool`)
    - **description** (`string`, <= 255 chars)

---

## Customization and Extensibility

- **Default Rules**  
  You can maintain a library of common Security Group rules (e.g., internal traffic, standard ports) and reuse them
  across multiple manifests.

- **Self Reference**  
  Allow traffic to/from the same group (like internal load balancer or clustered nodes) by setting `selfReference` to
  `true`.

- **Provider Credential**  
  Override AWS credentials at run time using Planton’s credential management or environment variables if needed.

- **Integration with Other Resources**  
  Combine your Security Group with ECS services, RDS instances, or any other Planton resource in a single
  environment stack.

---

## Further Reading

- **[examples.md](./examples.md)**: Multiple sample manifests showcasing common inbound/outbound rules, advanced usage,
  and minimal configurations.
- **[Planton Guide](https://github.com/plantonhq/planton/blob/main/docs/Guide.md)**: Explore
  multi-cloud workflows, advanced CLI usage, and best practices.
- **AWS Documentation
  **: [Amazon EC2 Security Groups](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-security-groups.html)
  for deeper insight into advanced security group configurations and constraints.
