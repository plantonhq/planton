# AwsElasticIp Examples

## 1. Minimal — Allocate a VPC Elastic IP

The simplest use case: allocate a static public IP from Amazon's pool.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: my-eip
  org: acme
  env: dev
  id: my-eip-dev
spec: {}
```

**What this creates:** A single VPC Elastic IP from Amazon's default pool. The `allocation_id` and `public_ip` outputs are immediately available for downstream references.

---

## 2. Static IPs for a Network Load Balancer

Allocate dedicated Elastic IPs for each NLB availability zone, then wire them via `valueFrom`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: nlb-eip-us-east-1a
  org: acme
  env: prod
  id: nlb-eip-us-east-1a-prod
spec: {}
---
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: nlb-eip-us-east-1b
  org: acme
  env: prod
  id: nlb-eip-us-east-1b-prod
spec: {}
```

Reference from an AwsNetworkLoadBalancer:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNetworkLoadBalancer
metadata:
  name: api-nlb
  org: acme
  env: prod
  id: api-nlb-prod
spec:
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.public_subnets.[0].id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: nlb-eip-us-east-1a
          fieldPath: status.outputs.allocation_id
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.public_subnets.[1].id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: nlb-eip-us-east-1b
          fieldPath: status.outputs.allocation_id
  listeners:
    - name: https
      port: 443
      protocol: TLS
      tlsConfig:
        certificateArn:
          valueFrom:
            kind: AwsCertManagerCert
            name: api-cert
            fieldPath: status.outputs.certificate_arn
      targetGroup:
        port: 8443
        protocol: TCP
        targetType: ip
```

---

## 3. EIP for a NAT Gateway (Future Pattern)

Allocate an EIP for a NAT Gateway to provide a predictable outbound IP for private subnets:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: nat-eip
  org: acme
  env: prod
  id: nat-eip-prod
spec: {}
```

The NAT Gateway (when available) would reference:

```yaml
spec:
  allocationId:
    valueFrom:
      kind: AwsElasticIp
      name: nat-eip
      fieldPath: status.outputs.allocation_id
```

---

## 4. BYOIP — Allocate from Your Own IP Range

When your organization has registered a BYOIP address range with AWS:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: byoip-eip
  org: acme
  env: prod
  id: byoip-eip-prod
spec:
  publicIpv4Pool: ipv4pool-ec2-0123456789abcdef0
```

---

## 5. BYOIP — Request a Specific IP Address

When you need a specific IP from your BYOIP range (e.g., for DNS records already pointing to it):

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: specific-ip-eip
  org: acme
  env: prod
  id: specific-ip-eip-prod
spec:
  publicIpv4Pool: ipv4pool-ec2-0123456789abcdef0
  address: "198.51.100.10"
```

---

## 6. Wavelength Zone — Edge Deployment

Allocate an EIP in an AWS Wavelength zone for 5G edge applications:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: edge-eip
  org: acme
  env: prod
  id: edge-eip-prod
spec:
  networkBorderGroup: us-east-1-wl1-bos-wlz-1
```

---

## 7. Production — Multi-Purpose EIP with Direct Value

When using a direct allocation ID value (no valueFrom):

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: bastion-eip
  org: acme
  env: prod
  id: bastion-eip-prod
  labels:
    team: platform
    purpose: bastion-host
spec: {}
```

After deployment, the outputs are available:
- `status.outputs.allocation_id` — Use in NLB subnet mappings or NAT Gateways
- `status.outputs.public_ip` — The actual IP address for DNS records or firewall rules
- `status.outputs.arn` — For IAM policy resource conditions
- `status.outputs.public_dns` — The AWS-assigned DNS name
