# AWS ALB Examples

Below are several examples demonstrating how to define an AWS ALB (Application Load Balancer) resource in OpenMCF. After creating one of these YAML manifests, apply it with your preferred IaC engine (Pulumi or Terraform) using the OpenMCF CLI:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

or

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Internet-Facing ALB

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: basic-alb
spec:
  region: us-west-2
  subnets:
    - subnet-12345abc
    - subnet-67890def
  securityGroups:
    - sg-abcdef1234567890
```

This example:
- Creates an internet-facing ALB (default behavior when `internal` is not set).
- Deploys across two subnets for high availability.
- Attaches one security group to control traffic.

---

## Internal ALB

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: internal-alb
spec:
  region: us-west-2
  subnets:
    - subnet-internal1
    - subnet-internal2
  securityGroups:
    - sg-internal1234
  internal: true
```

This ALB is internal-facing:
- `internal: true` makes the ALB accessible only within the VPC.
- Ideal for private microservices communication.

---

## ALB with SSL/TLS Termination

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: alb-with-ssl
spec:
  region: us-west-2
  subnets:
    - subnet-1111aaaa
    - subnet-2222bbbb
  securityGroups:
    - sg-zzzzxxxxccccdddd
  ssl:
    enabled: true
    certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abcd1234-5678-efgh-ijkl-123456abcdef
```

This configuration:
- Enables SSL/TLS termination on the ALB.
- Uses an AWS Certificate Manager certificate for HTTPS.

---

## ALB with Custom Idle Timeout and Deletion Protection

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: alb-protected
spec:
  region: us-west-2
  subnets:
    - subnet-custom1
    - subnet-custom2
  securityGroups:
    - sg-custom
  deleteProtectionEnabled: true
  idleTimeoutSeconds: 120
```

This configuration:
- Sets a custom idle timeout of 120 seconds (default is 60).
- Protects the ALB from accidental deletion with `deleteProtectionEnabled: true`.

---

## ALB with Route53 DNS Management

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: alb-with-dns
spec:
  region: us-west-2
  subnets:
    - subnet-abc123
    - subnet-def456
  securityGroups:
    - sg-dns123
  dns:
    enabled: true
    route53ZoneId: Z1234567890ABC
    hostnames:
      - app.example.com
      - api.example.com
```

This example:
- Automatically creates Route53 DNS records for the ALB.
- Maps both `app.example.com` and `api.example.com` to the ALB's DNS name.
- Requires a Route53 Hosted Zone ID.

---

## Complete Example with All Features

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: full-featured-alb
spec:
  region: us-west-2
  subnets:
    - subnet-prod1
    - subnet-prod2
    - subnet-prod3
  securityGroups:
    - sg-prod-alb
  internal: false
  deleteProtectionEnabled: true
  idleTimeoutSeconds: 90
  ssl:
    enabled: true
    certificateArn: arn:aws:acm:us-east-1:987654321098:certificate/prod-cert-12345
  dns:
    enabled: true
    route53ZoneId: Z9876543210DEF
    hostnames:
      - www.production.com
      - api.production.com
```

This comprehensive example:
- Creates an internet-facing ALB across three subnets.
- Enables SSL with a production certificate.
- Configures Route53 DNS for multiple hostnames.
- Enables deletion protection and custom idle timeout.

---

## Using Foreign Key References

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: alb-with-refs
spec:
  region: us-west-2
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.subnet_ids[0]
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.subnet_ids[1]
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: alb-sg
        field: status.outputs.security_group_id
  ssl:
    enabled: true
    certificateArn:
      valueFrom:
        kind: AwsCertManagerCert
        name: my-cert
        field: status.outputs.cert_arn
  dns:
    enabled: true
    route53ZoneId:
      valueFrom:
        kind: AwsRoute53Zone
        name: my-zone
        field: status.outputs.zone_id
    hostnames:
      - app.mydomain.com
```

This example demonstrates:
- Using foreign key references to other OpenMCF resources.
- Dynamically referencing subnet IDs from an AwsVpc resource.
- Referencing security group IDs, certificate ARNs, and Route53 zone IDs.
- This pattern allows for composable infrastructure definitions.

---

## After Deploying

Once you've applied your manifest with OpenMCF, you can verify the ALB creation:

```shell
aws elbv2 describe-load-balancers --names <your-load-balancer-name>
```

You should see details about your Application Load Balancer, including its DNS name, ARN, and configuration. The ALB DNS name can be used to access your applications or services behind the load balancer.
