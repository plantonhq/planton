---
title: Worked Example — A Minimal but Real Chart
description: A complete three-resource chart (VPC, internet gateway, public subnet) showing Chart.yaml, values.yaml, and a wired template with valueFrom references. Read when you want the full shape of a small chart in one place before writing your first files, or to check your file layout against a known-good one.
---

# Worked Example — A Minimal but Real Chart

Every chart has this shape; the example below fills it in:

```
my-chart/
├── Chart.yaml      # identity + description (an InfraChart manifest)
├── values.yaml     # the parameters users can set, with defaults
└── templates/      # YAML manifests with Jinja placeholders, any nesting
    ├── network.yaml            # multiple resources per file, separated by ---
    └── kubernetes/addons/…     # subdirectories are fine
```

`Chart.yaml`:

```yaml
apiVersion: infra-hub.planton.ai/v1alpha1
kind: InfraChart
metadata:
  name: VPC with Public Subnet
spec:
  selector:
    kind: organization
  description: A VPC with one public subnet and internet egress.
```

`values.yaml`:

```yaml
params:
  - name: aws_region
    description: AWS region for every resource
    value: us-east-1
  - name: vpc_cidr
    description: Primary IPv4 CIDR block for the VPC
    value: 10.0.0.0/16
  - name: subnet_cidr
    description: CIDR for the public subnet (must be inside vpc_cidr)
    value: 10.0.0.0/24
  - name: availability_zone
    description: Availability zone for the subnet (must belong to aws_region)
    value: us-east-1a
```

`templates/network.yaml`:

```yaml
---
apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpc
metadata:
  name: "{{ values.env }}-vpc"
spec:
  region: "{{ values.aws_region }}"
  cidrBlock: "{{ values.vpc_cidr }}"
  enableDnsSupport: true
  enableDnsHostnames: true
---
apiVersion: aws.planton.dev/v1alpha1
kind: AwsInternetGateway
metadata:
  name: "{{ values.env }}-igw"
spec:
  region: "{{ values.aws_region }}"
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: "{{ values.env }}-vpc"
      fieldPath: status.outputs.vpc_id
---
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSubnet
metadata:
  name: "{{ values.env }}-public-1"
spec:
  region: "{{ values.aws_region }}"
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: "{{ values.env }}-vpc"
      fieldPath: status.outputs.vpc_id
  availabilityZone: "{{ values.availability_zone }}"
  cidrBlock: "{{ values.subnet_cidr }}"
  mapPublicIpOnLaunch: true
  routes:
    - destinationCidrBlock: 0.0.0.0/0
      targetType: internet_gateway
      targetId:
        valueFrom:
          kind: AwsInternetGateway
          name: "{{ values.env }}-igw"
          fieldPath: status.outputs.internet_gateway_id
```

Then: `planton chart build . -o json` → fix until exit 0 → confirm the resources array lists the VPC, gateway, and subnet.
