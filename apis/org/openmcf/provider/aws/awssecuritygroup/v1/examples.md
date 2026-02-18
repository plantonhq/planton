# AWS Security Group Examples

## Create Using CLI

Create a YAML manifest using one of the examples below. After the YAML is created, apply it with OpenMCF:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

Or, if using Terraform:

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

(You can also use the shorter form `planton apply -f <yaml-path>` if your environment is configured accordingly.)

---

## Basic Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: my-basic-security-group
  version:
    message: "Initial AWS Security Group"
spec:
  region: us-west-2
  vpcId: "vpc-12345abcde"
  description: "Basic security group allowing inbound HTTP"
  ingress:
    - protocol: "tcp"
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTP inbound from anywhere"
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow all outbound"
```

This creates a simple security group that:
• Attaches to an existing VPC `vpc-12345abcde`.
• Allows inbound HTTP traffic on port 80 from any IPv4 address.
• Permits all outbound traffic.

---

## Example with Multiple Ingress Rules

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: multi-rule-sg
  version:
    message: "Multiple inbound rules"
spec:
  region: us-west-2
  vpcId: "vpc-67890fghij"
  description: "Multiple ingress ports for web and SSH"
  ingress:
    - protocol: "tcp"
      fromPort: 22
      toPort: 22
      ipv4Cidrs:
        - "10.0.0.0/24"
      description: "Allow SSH from internal subnet"
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTPS from anywhere"
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow all outbound"
```

In this example:
• SSH (port 22) is restricted to the `10.0.0.0/24` internal network.
• HTTPS (port 443) traffic is allowed from any IPv4 source.
• Outbound traffic is fully open.

---

## Example with IPv6 Rules

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: ipv6-enabled-sg
  version:
    message: "Inbound rules using IPv6"
spec:
  region: us-west-2
  vpcId: "vpc-13579abcd"
  description: "Allow inbound IPv4 and IPv6 traffic for HTTP"
  ingress:
    - protocol: "tcp"
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "10.1.2.0/24"
      ipv6Cidrs:
        - "::/0"
      description: "Allow HTTP from a specific IPv4 range and all IPv6"
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      ipv6Cidrs:
        - "::/0"
      description: "Allow all outbound IPv4 and IPv6"
```

Here, both IPv4 and IPv6 addresses are configured for inbound and outbound traffic. This is helpful if your application
supports IPv6 in addition to IPv4.

---

## Example with Self-Reference

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: sg-with-self-reference
  version:
    message: "Security group with self-referencing"
spec:
  region: us-west-2
  vpcId: "vpc-abcdef1234"
  description: "Allows traffic from the same security group"
  ingress:
    - protocol: "tcp"
      fromPort: 8080
      toPort: 8080
      selfReference: true
      description: "Allow inbound 8080 traffic from itself"
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      selfReference: true
      description: "Allow all outbound to itself"
```

With `selfReference` set to `true`, resources in the same security group can communicate with each other on the
specified ports. This is often used in cluster or load balancer scenarios where internal traffic needs to be allowed.

---

## Example with Security Group References

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: sg-with-group-refs
  version:
    message: "Ingress from another security group"
spec:
  region: us-west-2
  vpcId: "vpc-ghijklmn12"
  description: "Allows inbound traffic only from specific security group"
  ingress:
    - protocol: "tcp"
      fromPort: 3306
      toPort: 3306
      sourceSecurityGroupIds:
        - "sg-1234567890abcdef0"
      description: "Allow MySQL inbound from a DB client security group"
  egress:
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow outbound HTTPS to anywhere"
```

In this scenario, inbound MySQL traffic is permitted only from `sg-1234567890abcdef0`. This pattern is ideal for
internal service-to-service traffic within AWS.

---

## Minimal Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: minimal-security-group
  version:
    message: "Minimal SG example"
spec:
  region: us-west-2
  vpcId: "vpc-zz99xx88yy"
  description: "Just enough config for a valid SG"
```

This minimal manifest:
• Requires a VPC ID and a description.
• Has no ingress or egress rules, which means all inbound traffic is denied, and outbound defaults to allow (depending
on AWS defaults).

---

## After Deploying

Once you’ve chosen one of these examples, apply it with OpenMCF using Pulumi or Terraform:

```shell
openmcf pulumi up --manifest aws-sg.yaml --stack myorg/dev
```

or

```shell
openmcf terraform apply --manifest aws-sg.yaml --stack myorg/dev
```

When the command completes successfully, your Security Group will be created in AWS. You can confirm by checking the AWS
console or by using the AWS CLI:

```shell
aws ec2 describe-security-groups --group-names <your-security-group-name>
```

You should now have a functional AWS Security Group that fits your specified rules and VPC configuration.

---

Happy securing with OpenMCF!

P.S. If you run into any issues or have questions, feel free to open an issue in our GitHub repository or reach out to
our community forums.
