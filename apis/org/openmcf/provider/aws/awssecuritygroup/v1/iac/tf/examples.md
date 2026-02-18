# AWS Security Group Examples

Below are several examples demonstrating how to define an AWS Security Group component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: my-basic-security-group
spec:
  region: us-west-2
  vpcId:
    value: "vpc-12345abcde"
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

This example creates a basic security group:
• Attaches to VPC vpc-12345abcde.
• Allows inbound HTTP traffic on port 80 from any IPv4 address.
• Permits all outbound traffic.
• Suitable for web servers and basic applications.

---

## Web Server Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: web-server-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-67890fghij"
  description: "Security group for web servers with HTTP, HTTPS, and SSH access"
  ingress:
    - protocol: "tcp"
      fromPort: 22
      toPort: 22
      ipv4Cidrs:
        - "10.0.0.0/24"
      description: "Allow SSH from internal subnet"
    - protocol: "tcp"
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTP from anywhere"
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

This example creates a web server security group:
• SSH access restricted to internal subnet (10.0.0.0/24).
• HTTP and HTTPS access from anywhere.
• All outbound traffic permitted.
• Suitable for public-facing web servers.

---

## Database Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: database-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-13579abcd"
  description: "Security group for database servers with restricted access"
  ingress:
    - protocol: "tcp"
      fromPort: 3306
      toPort: 3306
      sourceSecurityGroupIds:
        - "sg-1234567890abcdef0"
      description: "Allow MySQL from application security group"
    - protocol: "tcp"
      fromPort: 5432
      toPort: 5432
      sourceSecurityGroupIds:
        - "sg-1234567890abcdef0"
      description: "Allow PostgreSQL from application security group"
  egress:
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow outbound HTTPS for updates"
```

This example creates a database security group:
• MySQL access from specific application security group.
• PostgreSQL access from application security group.
• Restricted outbound access to HTTPS only.
• Suitable for database servers with strict access control.

---

## Application Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: application-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-abcdef1234"
  description: "Security group for application servers"
  ingress:
    - protocol: "tcp"
      fromPort: 8080
      toPort: 8080
      ipv4Cidrs:
        - "10.0.0.0/16"
      description: "Allow application traffic from VPC"
    - protocol: "tcp"
      fromPort: 22
      toPort: 22
      ipv4Cidrs:
        - "10.0.0.0/24"
      description: "Allow SSH from management subnet"
  egress:
    - protocol: "tcp"
      fromPort: 3306
      toPort: 3306
      destinationSecurityGroupIds:
        - "sg-database123456"
      description: "Allow database access"
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTPS outbound"
```

This example creates an application security group:
• Application traffic on port 8080 from VPC.
• SSH access from management subnet.
• Database access to specific security group.
• HTTPS outbound for external API calls.
• Suitable for application servers with controlled access.

---

## Load Balancer Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: load-balancer-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-ghijklmn12"
  description: "Security group for application load balancer"
  ingress:
    - protocol: "tcp"
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTP from anywhere"
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTPS from anywhere"
  egress:
    - protocol: "tcp"
      fromPort: 8080
      toPort: 8080
      destinationSecurityGroupIds:
        - "sg-application123456"
      description: "Allow traffic to application servers"
```

This example creates a load balancer security group:
• HTTP and HTTPS access from anywhere.
• Traffic forwarding to application security group.
• Suitable for public-facing load balancers.
• Restricted outbound access to application servers only.

---

## Bastion Host Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: bastion-host-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-1234567890"
  description: "Security group for bastion host with restricted SSH access"
  ingress:
    - protocol: "tcp"
      fromPort: 22
      toPort: 22
      ipv4Cidrs:
        - "203.0.113.0/24"
      description: "Allow SSH from office IP range"
  egress:
    - protocol: "tcp"
      fromPort: 22
      toPort: 22
      destinationSecurityGroupIds:
        - "sg-private-servers"
      description: "Allow SSH to private servers"
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTPS for updates"
```

This example creates a bastion host security group:
• SSH access restricted to office IP range.
• SSH forwarding to private server security group.
• HTTPS outbound for system updates.
• Suitable for secure administrative access.

---

## Redis Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: redis-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-redis123456"
  description: "Security group for Redis cache servers"
  ingress:
    - protocol: "tcp"
      fromPort: 6379
      toPort: 6379
      sourceSecurityGroupIds:
        - "sg-application123456"
      description: "Allow Redis access from application servers"
  egress:
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTPS outbound"
```

This example creates a Redis security group:
• Redis access on port 6379 from application servers.
• HTTPS outbound for updates and monitoring.
• Suitable for Redis cache servers.
• Restricted access for security.

---

## Monitoring Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: monitoring-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-monitoring123"
  description: "Security group for monitoring and logging servers"
  ingress:
    - protocol: "tcp"
      fromPort: 9100
      toPort: 9100
      sourceSecurityGroupIds:
        - "sg-application123456"
      description: "Allow Prometheus metrics from applications"
    - protocol: "tcp"
      fromPort: 22
      toPort: 22
      ipv4Cidrs:
        - "10.0.0.0/24"
      description: "Allow SSH from management subnet"
  egress:
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "Allow HTTPS for external monitoring services"
```

This example creates a monitoring security group:
• Prometheus metrics collection from applications.
• SSH access from management subnet.
• HTTPS outbound for external monitoring services.
• Suitable for monitoring and observability infrastructure.

---

## Self-Referencing Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: cluster-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-cluster123"
  description: "Security group allowing internal cluster communication"
  ingress:
    - protocol: "tcp"
      fromPort: 8080
      toPort: 8080
      selfReference: true
      description: "Allow internal communication on port 8080"
    - protocol: "tcp"
      fromPort: 9090
      toPort: 9090
      selfReference: true
      description: "Allow internal communication on port 9090"
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      selfReference: true
      description: "Allow all internal outbound traffic"
```

This example creates a self-referencing security group:
• Internal communication on ports 8080 and 9090.
• All internal outbound traffic permitted.
• Suitable for cluster environments.
• Enables service-to-service communication.

---

## IPv6 Enabled Security Group

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: ipv6-sg
spec:
  region: us-west-2
  vpcId:
    value: "vpc-ipv6-enabled"
  description: "Security group with IPv6 support for dual-stack networking"
  ingress:
    - protocol: "tcp"
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "10.1.2.0/24"
      ipv6Cidrs:
        - "::/0"
      description: "Allow HTTP from IPv4 range and all IPv6"
    - protocol: "tcp"
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      ipv6Cidrs:
        - "::/0"
      description: "Allow HTTPS from anywhere"
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

This example creates an IPv6-enabled security group:
• Dual-stack IPv4 and IPv6 support.
• HTTP access from specific IPv4 range and all IPv6.
• HTTPS access from anywhere on both protocols.
• All outbound traffic on both protocols.
• Suitable for modern dual-stack applications.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the security group is active via the AWS console or by
using the AWS CLI:

```shell
aws ec2 describe-security-groups --group-names <your-security-group-name>
```

For detailed security group rules:

```shell
aws ec2 describe-security-group-rules --filters Name=group-id,Values=<your-security-group-id>
```

To list all security groups in your VPC:

```shell
aws ec2 describe-security-groups --filters Name=vpc-id,Values=<your-vpc-id>
```

This will show the security group details including rules, tags, and configuration information.
