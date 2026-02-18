# Examples

## Minimal: Single VPC Attachment

The simplest Transit Gateway with one VPC. Useful for development or as a starting point for expansion.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsTransitGateway
metadata:
  name: dev-tgw
spec:
  region: us-east-1
  description: Development Transit Gateway
  dnsSupport: true
  vpcAttachments:
    - name: dev-vpc
      vpcId:
        value: vpc-0a1b2c3d4e5f00001
      subnetIds:
        - value: subnet-0a1b2c3d4e5f00001
```

## Production: Multi-VPC Full Mesh

Two VPCs with multi-AZ subnets, full-mesh routing, and DNS support. This is the standard production pattern.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsTransitGateway
metadata:
  name: prod-tgw
spec:
  region: us-east-1
  description: Production multi-VPC connectivity hub
  amazonSideAsn: 64512
  defaultRouteTableAssociation: true
  defaultRouteTablePropagation: true
  dnsSupport: true
  vpnEcmpSupport: true
  vpcAttachments:
    - name: app-vpc
      vpcId:
        value: vpc-app001
      subnetIds:
        - value: subnet-app-az1
        - value: subnet-app-az2
      dnsSupport: true
    - name: data-vpc
      vpcId:
        value: vpc-data001
      subnetIds:
        - value: subnet-data-az1
        - value: subnet-data-az2
      dnsSupport: true
    - name: shared-services-vpc
      vpcId:
        value: vpc-shared001
      subnetIds:
        - value: subnet-shared-az1
        - value: subnet-shared-az2
      dnsSupport: true
```

## Cross-Resource References: valueFrom Pattern

Using `valueFrom` to reference VPC and subnet outputs from AwsVpc resources.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsTransitGateway
metadata:
  name: connected-tgw
spec:
  region: us-east-1
  description: TGW with cross-resource references
  dnsSupport: true
  vpcAttachments:
    - name: app-vpc
      vpcId:
        valueFrom:
          kind: AwsVpc
          name: my-app-vpc
          fieldPath: status.outputs.vpc_id
      subnetIds:
        - valueFrom:
            kind: AwsVpc
            name: my-app-vpc
            fieldPath: status.outputs.private_subnets.0.id
        - valueFrom:
            kind: AwsVpc
            name: my-app-vpc
            fieldPath: status.outputs.private_subnets.1.id
    - name: shared-vpc
      vpcId:
        valueFrom:
          kind: AwsVpc
          name: my-shared-vpc
          fieldPath: status.outputs.vpc_id
      subnetIds:
        - valueFrom:
            kind: AwsVpc
            name: my-shared-vpc
            fieldPath: status.outputs.private_subnets.0.id
```

## Centralized Firewall Inspection

Transit Gateway with a dedicated inspection VPC running a virtual firewall. Appliance mode ensures symmetric routing for stateful inspection.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsTransitGateway
metadata:
  name: inspection-tgw
spec:
  region: us-east-1
  description: Hub-and-spoke with centralized firewall
  dnsSupport: true
  vpcAttachments:
    - name: inspection-vpc
      vpcId:
        value: vpc-fw001
      subnetIds:
        - value: subnet-fw-az1
        - value: subnet-fw-az2
      applianceModeSupport: true
      dnsSupport: true
    - name: workload-vpc-a
      vpcId:
        value: vpc-work-a
      subnetIds:
        - value: subnet-work-a-az1
        - value: subnet-work-a-az2
    - name: workload-vpc-b
      vpcId:
        value: vpc-work-b
      subnetIds:
        - value: subnet-work-b-az1
        - value: subnet-work-b-az2
```
