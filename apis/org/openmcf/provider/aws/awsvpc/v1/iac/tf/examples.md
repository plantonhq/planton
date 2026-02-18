# AWS VPC Examples

Below are several examples demonstrating how to define an AWS VPC component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-basic-vpc
spec:
  region: "us-west-2"
  vpcCidr: "10.0.0.0/16"
  availabilityZones:
    - "us-west-2a"
    - "us-west-2b"
  subnetsPerAvailabilityZone: 1
  subnetSize: 256
  isNatGatewayEnabled: false
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a basic VPC:
• CIDR block 10.0.0.0/16 for IP address range.
• Two availability zones for high availability.
• One subnet per AZ with 256 hosts each.
• DNS support and hostnames enabled.
• NAT gateway disabled for cost optimization.
• Suitable for simple applications and development.

---

## Production VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-production-vpc
spec:
  region: "us-east-1"
  vpcCidr: "10.1.0.0/16"
  availabilityZones:
    - "us-east-1a"
    - "us-east-1b"
    - "us-east-1c"
  subnetsPerAvailabilityZone: 2
  subnetSize: 512
  isNatGatewayEnabled: true
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a production VPC:
• Three availability zones for maximum availability.
• Two subnets per AZ (public and private).
• Larger subnet size (512 hosts) for scalability.
• NAT gateway enabled for private subnet internet access.
• DNS support and hostnames enabled.
• Suitable for production workloads with high availability.

---

## Development VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-development-vpc
spec:
  region: "us-west-2"
  vpcCidr: "10.2.0.0/16"
  availabilityZones:
    - "us-west-2a"
  subnetsPerAvailabilityZone: 1
  subnetSize: 128
  isNatGatewayEnabled: false
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a development VPC:
• Single availability zone for cost optimization.
• One subnet per AZ for simplicity.
• Smaller subnet size (128 hosts) for development.
• NAT gateway disabled to reduce costs.
• DNS support and hostnames enabled.
• Suitable for development and testing environments.

---

## Multi-Tier VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-multi-tier-vpc
spec:
  region: "us-east-1"
  vpcCidr: "10.3.0.0/16"
  availabilityZones:
    - "us-east-1a"
    - "us-east-1b"
    - "us-east-1c"
  subnetsPerAvailabilityZone: 3
  subnetSize: 256
  isNatGatewayEnabled: true
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a multi-tier VPC:
• Three availability zones for high availability.
• Three subnets per AZ (web, application, database tiers).
• NAT gateway enabled for private tier internet access.
• DNS support and hostnames enabled.
• Suitable for complex multi-tier architectures.

---

## High-Availability VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-ha-vpc
spec:
  region: "us-west-2"
  vpcCidr: "10.4.0.0/16"
  availabilityZones:
    - "us-west-2a"
    - "us-west-2b"
    - "us-west-2c"
  subnetsPerAvailabilityZone: 2
  subnetSize: 1024
  isNatGatewayEnabled: true
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a high-availability VPC:
• Three availability zones for maximum redundancy.
• Two subnets per AZ (public and private).
• Large subnet size (1024 hosts) for scalability.
• NAT gateway enabled for private subnet access.
• DNS support and hostnames enabled.
• Suitable for mission-critical applications.

---

## Microservices VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-microservices-vpc
spec:
  region: "us-east-1"
  vpcCidr: "10.5.0.0/16"
  availabilityZones:
    - "us-east-1a"
    - "us-east-1b"
  subnetsPerAvailabilityZone: 4
  subnetSize: 128
  isNatGatewayEnabled: true
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a microservices VPC:
• Two availability zones for high availability.
• Four subnets per AZ for service isolation.
• Smaller subnet size for service granularity.
• NAT gateway enabled for service communication.
• DNS support and hostnames enabled.
• Suitable for microservices architectures.

---

## Database VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-database-vpc
spec:
  region: "us-east-1"
  vpcCidr: "10.6.0.0/16"
  availabilityZones:
    - "us-east-1a"
    - "us-east-1b"
    - "us-east-1c"
  subnetsPerAvailabilityZone: 1
  subnetSize: 512
  isNatGatewayEnabled: false
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a database VPC:
• Three availability zones for database redundancy.
• One subnet per AZ for database isolation.
• Larger subnet size for database clusters.
• NAT gateway disabled for security.
• DNS support and hostnames enabled.
• Suitable for database-focused workloads.

---

## Container VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-container-vpc
spec:
  region: "us-west-2"
  vpcCidr: "10.7.0.0/16"
  availabilityZones:
    - "us-west-2a"
    - "us-west-2b"
  subnetsPerAvailabilityZone: 3
  subnetSize: 256
  isNatGatewayEnabled: true
  isDnsHostnamesEnabled: true
  isDnsSupportEnabled: true
```

This example creates a container VPC:
• Two availability zones for container orchestration.
• Three subnets per AZ (public, private, data).
• NAT gateway enabled for container internet access.
• DNS support and hostnames enabled.
• Suitable for Kubernetes and container workloads.

---

## Minimal VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-minimal-vpc
spec:
  region: "us-east-1"
  vpcCidr: "10.8.0.0/16"
  availabilityZones:
    - "us-east-1a"
  subnetsPerAvailabilityZone: 1
  subnetSize: 64
  isNatGatewayEnabled: false
  isDnsHostnamesEnabled: false
  isDnsSupportEnabled: true
```

This example creates a minimal VPC:
• Single availability zone for minimal cost.
• One subnet for simple deployments.
• Small subnet size (64 hosts) for minimal resources.
• NAT gateway and DNS hostnames disabled.
• DNS support enabled for basic functionality.
• Suitable for simple testing and minimal deployments.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the VPC is active via the AWS console or by
using the AWS CLI:

```shell
aws ec2 describe-vpcs --vpc-ids <your-vpc-id>
```

For subnet information:

```shell
aws ec2 describe-subnets --filters Name=vpc-id,Values=<your-vpc-id>
```

To check route tables:

```shell
aws ec2 describe-route-tables --filters Name=vpc-id,Values=<your-vpc-id>
```

This will show the VPC details including CIDR blocks, subnets, route tables, and network configuration information.

