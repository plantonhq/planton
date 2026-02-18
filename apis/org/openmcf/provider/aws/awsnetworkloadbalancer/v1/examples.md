# AWS Network Load Balancer Examples

Below are several examples demonstrating how to define an AWS Network Load Balancer (NLB) resource in OpenMCF. After creating one of these YAML manifests, apply it with your preferred IaC engine (Pulumi or Terraform) using the OpenMCF CLI:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

or

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

---

## 1. Minimal TCP NLB (Simplest)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: minimal-tcp-nlb
  labels:
    openmcf.org/provisioner: pulumi
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-0a1b2c3d4e5f00001
    - subnetId: subnet-0a1b2c3d4e5f00002
  listeners:
    - name: tcp-80
      port: 80
      protocol: TCP
      targetGroup:
        port: 8080
        protocol: TCP
```

This example:
- Creates an internet-facing NLB across two subnets.
- Single TCP listener on port 80 forwarding to targets on port 8080.
- No security groups (optional for NLB).
- Minimal configuration for the simplest use case.

---

## 2. Internal NLB for Microservices

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: internal-microservices-nlb
  labels:
    openmcf.org/provisioner: pulumi
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-private-az1
    - subnetId: subnet-private-az2
  internal: true
  listeners:
    - name: tcp-443
      port: 443
      protocol: TCP
      targetGroup:
        port: 8443
        protocol: TCP
        targetType: ip
```

This example:
- Creates an internal NLB accessible only within the VPC.
- Uses private subnets for internal microservice-to-microservice traffic.
- Targets are IP addresses (typical for ECS Fargate or container workloads).

---

## 3. Internet-Facing with Static IPs (Elastic IPs)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: static-ip-nlb
  labels:
    openmcf.org/provisioner: pulumi
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-public-az1
      allocationId: eipalloc-0a1b2c3d4e5f00001
    - subnetId: subnet-public-az2
      allocationId: eipalloc-0a1b2c3d4e5f00002
  listeners:
    - name: tcp-443
      port: 443
      protocol: TCP
      targetGroup:
        port: 443
        protocol: TCP
```

This example:
- Assigns a static Elastic IP per Availability Zone.
- Critical for allowlisting by partners, firewalls, or legacy systems.
- Each AZ has a predictable public IP that does not change across scaling events.

---

## 4. TLS Termination with ACM Certificate

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: tls-termination-nlb
  labels:
    openmcf.org/provisioner: pulumi
spec:
  region: us-east-1
  subnetMappings:
    - subnetId: subnet-public-az1
    - subnetId: subnet-public-az2
  listeners:
    - name: tls-443
      port: 443
      protocol: TLS
      tls:
        certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abcd1234-5678-efgh-ijkl-123456abcdef
        sslPolicy: ELBSecurityPolicy-TLS13-1-2-2021-06
      targetGroup:
        port: 8080
        protocol: TCP
```

This example:
- Uses a TLS listener for TLS termination at the NLB.
- NLB decrypts incoming TLS and forwards plaintext TCP to targets on port 8080.
- ACM certificate and optional TLS 1.3 security policy.

---

## 5. Multiple Listeners (TCP + TLS)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: multi-listener-nlb
  labels:
    openmcf.org/provisioner: pulumi
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-public-az1
    - subnetId: subnet-public-az2
  listeners:
    - name: tcp-80
      port: 80
      protocol: TCP
      targetGroup:
        port: 8080
        protocol: TCP
    - name: tls-443
      port: 443
      protocol: TLS
      tls:
        certificateArn: arn:aws:acm:us-east-1:123456789012:certificate/abc-12345
      targetGroup:
        port: 8443
        protocol: TCP
```

This example:
- Two listeners: TCP on port 80 and TLS on port 443.
- Each listener has its own inline target group.
- Downstream services reference `status.outputs.target_group_arns.tcp-80` or `status.outputs.target_group_arns.tls-443`.

---

## 6. With Health Check and Target Group Options

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: nlb-with-health-and-options
  labels:
    openmcf.org/provisioner: pulumi
spec:
  region: us-west-2
  subnetMappings:
    - subnetId: subnet-az1
    - subnetId: subnet-az2
  listeners:
    - name: tcp-443
      port: 443
      protocol: TCP
      tcpIdleTimeoutSeconds: 300
      targetGroup:
        port: 8080
        protocol: TCP
        targetType: ip
        healthCheck:
          protocol: HTTP
          port: traffic-port
          path: /healthz
          healthyThreshold: 3
          unhealthyThreshold: 3
          intervalSeconds: 30
          timeoutSeconds: 10
          matcher: "200-399"
        deregistrationDelaySeconds: 120
        preserveClientIp: true
        connectionTermination: true
        stickinessEnabled: true
```

This example:
- HTTP health check on `/healthz` instead of default TCP.
- Preserves client IP for application logging and security.
- Connection termination for faster deregistration of long-lived connections.
- Source-IP stickiness for session affinity.
- Shorter deregistration delay (120s) for faster scaling.

---

## 7. Production-Ready with valueFrom References

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: prod-api-nlb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsNetworkLoadBalancer.prod-api-nlb
spec:
  region: us-west-2
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          field: status.outputs.public_subnets[0].id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: prod-nlb-eip-az1
          field: status.outputs.allocation_id
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          field: status.outputs.public_subnets[1].id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: prod-nlb-eip-az2
          field: status.outputs.allocation_id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: prod-nlb-sg
        field: status.outputs.security_group_id
  internal: false
  deleteProtectionEnabled: true
  crossZoneLoadBalancingEnabled: true
  listeners:
    - name: tls-443
      port: 443
      protocol: TLS
      tls:
        certificateArn:
          valueFrom:
            kind: AwsCertManagerCert
            name: prod-api-cert
            field: status.outputs.cert_arn
      targetGroup:
        port: 8080
        protocol: TCP
        targetType: ip
        healthCheck:
          protocol: HTTPS
          path: /api/health
          matcher: "200"
  dns:
    enabled: true
    route53ZoneId:
      valueFrom:
        kind: AwsRoute53Zone
        name: example-com-zone
        field: status.outputs.zone_id
    hostnames:
      - api.example.com
```

This example:
- References VPC subnets, Elastic IPs, security groups, certificate, and Route53 zone via `valueFrom`.
- **Note**: If `AwsElasticIp` is not available in your OpenMCF version, use a direct `allocationId` value (e.g., `eipalloc-xxx`) instead of the `valueFrom` reference.
- Composable infrastructure with declarative cross-resource dependencies.
- Production settings: deletion protection, cross-zone load balancing, HTTPS health check.
- Route53 alias records for `api.example.com`.

---

## After Deploying

Once you've applied your manifest with OpenMCF, you can verify the NLB creation:

```shell
aws elbv2 describe-load-balancers --names <your-load-balancer-name>
```

To register targets with a target group (e.g., from ECS or EKS), reference the target group ARN:

```yaml
# Example: ECS service registering with NLB target group
targetGroupArn:
  valueFrom:
    kind: AwsNetworkLoadBalancer
    name: prod-api-nlb
    field: status.outputs.target_group_arns.tls-443
```
