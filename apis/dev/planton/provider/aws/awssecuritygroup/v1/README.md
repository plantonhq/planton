# AwsSecurityGroup API

The **AwsSecurityGroup** API resource provides a standardized and user-friendly way to create and manage AWS EC2
Security Groups. By focusing on essential configurations such as inbound and outbound rules, VPC attachment, and
descriptions, it aims to streamline the process of securing cloud resources under the Planton multi-cloud
deployment framework.

## Purpose

Deploying AWS Security Groups can sometimes be cumbersome—defining numerous rules, ensuring consistency across
environments, and enforcing naming and usage policies. The **AwsSecurityGroup** resource addresses these challenges by:

- **Centralizing Security Configurations**: Offers a concise and consistent way to manage inbound and outbound traffic
  rules.
- **Encouraging Best Practices**: Provides recommended patterns for restricting or allowing traffic, ensuring you follow
  AWS guidelines.
- **Simplifying Multi-Cloud Workflows**: Leverages Planton’s uniform resource model, letting you declare AWS
  Security Groups alongside other provider resources in the same manifest.

## Key Features

### Inbound (Ingress) & Outbound (Egress) Rules

- **Granular Rule Definitions**: Easily specify protocol, port ranges, CIDR blocks, and other Security Groups for
  inbound and outbound traffic.
- **Flexible Protocol Handling**: Support for TCP, UDP, ICMP, or `-1` for “all protocols,” enabling quick setup of
  typical or specialized rules.

### Self-Reference

- **Allow Internal Communication**: Optionally allow traffic to and from the same Security Group, commonly used for
  internal load balancers or cluster communication.

### VPC Integration

- **Mandatory VPC Attachment**: The resource requires you to specify a VPC, ensuring your Security Group is always
  well-scoped within an AWS network boundary.
- **Seamless AWS Networking**: Ties in with subnets, NAT gateways, and internet gateways, consistent with
  Planton’s approach to multi-service deployments.

### Minimal, Opinionated Spec

- **Concise Configuration**: Define only the essential fields—Amazon Resource Names (ARNs) or IDs for references, rule
  sets, and a description.
- **Validations & Conventions**: Planton ensures naming, rule definitions, and descriptions meet AWS constraints
  before attempting deployment.

## Benefits

- **Reduced Setup Complexity**: A single spec handles both inbound and outbound rules without juggling multiple
  AWS-specific templates or parameters.
- **Improved Security Posture**: Encourages the principle of least privilege by providing a clear, easily-auditable
  structure for traffic rules.
- **Multi-Cloud & Multi-Account**: Use the same workflow in Planton to manage Security Groups across various AWS
  accounts, or combine with other cloud resources in a single manifest.

## Example Usage

Below is a minimal YAML snippet demonstrating how to configure and deploy an AWS Security Group using Planton.
All keys are in camel-case to maintain consistency.

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsSecurityGroup
metadata:
  name: my-security-group
  version:
    message: "Initial AWS Security Group deployment"
spec:
  vpcId: "vpc-12345abcde"
  description: "Allows inbound HTTP and SSH for web tier"
  ingress:
    - protocol: "tcp"
      fromPort: 22
      toPort: 22
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow inbound SSH from anywhere"
    - protocol: "tcp"
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow inbound HTTP from anywhere"
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow all outbound traffic"
```

### Deploying with Planton

1. **Validate the Manifest (Optional)**

   ```bash
   planton validate --manifest aws-sg.yaml
   ```

2. **Pulumi Deployment**

   ```bash
   planton pulumi up --manifest aws-sg.yaml --stack org/project/my-stack
   ```

3. **Terraform Deployment**

   ```bash
   planton terraform apply --manifest aws-sg.yaml --stack org/project/my-stack
   ```

Planton will validate your manifest against the Protobuf schema and then provision your Security Group using
either Pulumi or Terraform. You can easily integrate this step into CI/CD pipelines, combine it with other AWS resources
in the same YAML, and maintain consistent security rules across multiple environments.

---

Happy securing! If you have questions or encounter issues, please open an issue on our GitHub repository or connect with
our community for assistance.
