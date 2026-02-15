## 1. Minimal Single-Node (Public, No VPC)

A single-node development domain with gp3 storage, encryption enabled, and no VPC (publicly accessible). Suitable for development, prototyping, and learning.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: dev-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsOpenSearchDomain.dev-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: t3.small.search
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 10
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
```

## 2. Production VPC with Dedicated Masters and Zone Awareness

A production-grade domain with 3 data nodes across 3 AZs, 3 dedicated master nodes, VPC deployment using `valueFrom` references, FGAC with internal user database, and enforced HTTPS.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: prod-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.prod-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 3
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 3
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 100
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.private_subnets.[0].id
      - valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.private_subnets.[1].id
      - valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.private_subnets.[2].id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: opensearch-sg
          fieldPath: status.outputs.security_group_id
  domainEndpointOptions:
    enforceHttps: true
    tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: true
    masterUserName: admin
    masterUserPassword:
      value: "ChangeMe!Str0ng#2024"
  autoTuneEnabled: true
```

## 3. FGAC with Internal User Database

A domain with fine-grained access control using the internal user database. No VPC — secured via access policies and FGAC. Suitable for small teams or development environments that need role-based access.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: fgac-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsOpenSearchDomain.fgac-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 2
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 50
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  domainEndpointOptions:
    enforceHttps: true
    tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: true
    masterUserName: admin
    masterUserPassword:
      value: "MyS3cure!Pass#2024"
  accessPolicies:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          AWS: "*"
        Action: "es:ESHttp*"
        Resource: "arn:aws:es:us-east-1:123456789012:domain/fgac-search/*"
```

## 4. FGAC with IAM Master User

Fine-grained access control using an IAM role as the master user instead of the internal user database. Ideal for organizations that centralize identity management through IAM.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: iam-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.iam-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.xlarge.search
    instanceCount: 3
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 3
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 100
  encryptAtRestEnabled: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: opensearch-key
      fieldPath: status.outputs.key_arn
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.private_subnets.[0].id
      - valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.private_subnets.[1].id
      - valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.private_subnets.[2].id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: opensearch-sg
          fieldPath: status.outputs.security_group_id
  domainEndpointOptions:
    enforceHttps: true
    tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: false
    masterUserArn:
      valueFrom:
        kind: AwsIamRole
        name: opensearch-admin
        fieldPath: status.outputs.role_arn
  autoTuneEnabled: true
```

## 5. Analytics with Warm + Cold Storage

An analytics-optimized domain with 3 data nodes, 3 UltraWarm nodes for infrequently accessed data, and cold storage enabled for archival data. Ideal for log analytics, time-series data, and SIEM workloads where older data is retained at lower cost.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: analytics-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.analytics-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.xlarge.search
    instanceCount: 3
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 3
    warmEnabled: true
    warmType: ultrawarm1.medium.search
    warmCount: 3
    coldStorageEnabled: true
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 200
    iops: 6000
    throughput: 250
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - value: subnet-0a1b2c3d4e5f00001
      - value: subnet-0a1b2c3d4e5f00002
      - value: subnet-0a1b2c3d4e5f00003
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f00001
  domainEndpointOptions:
    enforceHttps: true
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: true
    masterUserName: admin
    masterUserPassword:
      value: "Analyt1cs!Str0ng#2024"
  logPublishingOptions:
    - logType: INDEX_SLOW_LOGS
      cloudwatchLogGroupArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/opensearch/analytics-search/index-slow-logs"
    - logType: SEARCH_SLOW_LOGS
      cloudwatchLogGroupArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/opensearch/analytics-search/search-slow-logs"
    - logType: ES_APPLICATION_LOGS
      cloudwatchLogGroupArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/opensearch/analytics-search/app-logs"
  autoTuneEnabled: true
```

## 6. Custom Endpoint with ACM Certificate

A domain with a custom domain endpoint (e.g., `search.example.com`) backed by an ACM certificate. Useful for providing a stable, branded endpoint to applications.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: branded-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.branded-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 2
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 2
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 80
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - value: subnet-private-az1
      - value: subnet-private-az2
    securityGroupIds:
      - value: sg-opensearch
  domainEndpointOptions:
    enforceHttps: true
    tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"
    customEndpointEnabled: true
    customEndpoint: "search.example.com"
    customEndpointCertificateArn:
      valueFrom:
        kind: AwsCertManagerCert
        name: search-cert
        fieldPath: status.outputs.certificate_arn
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: true
    masterUserName: admin
    masterUserPassword:
      value: "Brand3d!Search#2024"
```

## 7. Log Publishing to CloudWatch

A domain with all four log types published to CloudWatch, including audit logs (requires FGAC). Designed for environments with strict observability and compliance requirements.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: logged-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.logged-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 3
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 3
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 100
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - value: subnet-0a1b2c3d4e5f00001
      - value: subnet-0a1b2c3d4e5f00002
      - value: subnet-0a1b2c3d4e5f00003
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f00001
  domainEndpointOptions:
    enforceHttps: true
    tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: true
    masterUserName: admin
    masterUserPassword:
      value: "L0gged!Secure#2024"
  logPublishingOptions:
    - logType: INDEX_SLOW_LOGS
      cloudwatchLogGroupArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/opensearch/logged-search/index-slow"
    - logType: SEARCH_SLOW_LOGS
      cloudwatchLogGroupArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/opensearch/logged-search/search-slow"
    - logType: ES_APPLICATION_LOGS
      cloudwatchLogGroupArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/opensearch/logged-search/app-logs"
    - logType: AUDIT_LOGS
      cloudwatchLogGroupArn:
        value: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/opensearch/logged-search/audit-logs"
  autoTuneEnabled: true
  autoSoftwareUpdateEnabled: true
```

## CLI Flows

Validate:

```bash
openmcf validate --manifest opensearch-domain.yaml
```

Pulumi deploy:

```bash
openmcf pulumi update --manifest opensearch-domain.yaml --stack org/project/stack --module-dir apis/org/openmcf/provider/aws/awsopensearchdomain/v1/iac/pulumi
```

Terraform deploy:

```bash
openmcf tofu apply --manifest opensearch-domain.yaml --auto-approve
```
