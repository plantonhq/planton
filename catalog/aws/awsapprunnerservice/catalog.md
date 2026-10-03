# AWS App Runner Service

Deploys an App Runner service — the shortest path from a container image or source repository to an HTTPS endpoint on AWS. App Runner handles build, deploy, TLS, load balancing, and concurrency-based auto scaling. The service composes with its shared, versioned companion resources by reference rather than embedding them: an [auto scaling configuration](/infra-catalog/aws-app-runner-auto-scaling-configuration) tunes scaling for a fleet, a [VPC connector](/infra-catalog/aws-app-runner-vpc-connector) routes outbound traffic into a VPC, and an [observability configuration](/infra-catalog/aws-app-runner-observability-configuration) enables X-Ray tracing — each shared by any number of services. Custom domains with managed TLS and a REGIONAL WAF web ACL complete the public edge.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **App Runner Service** -- a fully managed container service with source configuration (image or code), instance sizing, port mapping, health checks, ingress settings, encryption, and the companion-resource attachments
- **Custom Domain Associations** -- one per `customDomains` entry, each with an App Runner-managed TLS certificate whose validation records export in outputs
- **WAF Association** -- when `webAclArn` is set, every request passes the referenced REGIONAL web ACL before reaching the application
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Companion resources** (optional, declaration before reference) -- create the [AwsAppRunnerAutoScalingConfiguration](/infra-catalog/aws-app-runner-auto-scaling-configuration), [AwsAppRunnerVpcConnector](/infra-catalog/aws-app-runner-vpc-connector), and [AwsAppRunnerObservabilityConfiguration](/infra-catalog/aws-app-runner-observability-configuration) resources first; the service references their ARN outputs.

### AWS Account

- **A container image in ECR or ECR Public** if using image-based deployment. For private ECR, an IAM access role assumable by `build.apprunner.amazonaws.com` with ECR read permissions (the AWS-managed `AWSAppRunnerServicePolicyForECRAccess` policy covers it).
- **An App Runner Connection** if using code-based deployment. Connections require a one-time OAuth handshake in the AWS Console and are referenced here by ARN; one connection is shared across services.
- **A KMS key** (optional) for encrypting the stored deployment source beyond the AWS-managed key. ForceNew: changing it replaces the service.
- **A REGIONAL WAF web ACL** (optional) in the service's region — CloudFront-scope ACLs cannot attach to App Runner.

## Deploy

### Console

Open the deployment store, find **AWS App Runner Service**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Public Image**, **Production VPC-Connected and Encrypted Service**, or **GitHub Code Source (Node.js)** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerService
metadata:
  name: my-api
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  imageSource:
    imageIdentifier: "public.ecr.aws/nginx/nginx:latest"
    imageRepositoryType: "ECR_PUBLIC"
```

```shell
planton apply -f app-runner.yaml
```

This creates a publicly accessible App Runner service running an ECR Public image with default settings: 1 vCPU, 2 GB memory, port 8080, and the account's default auto scaling posture. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the service to its companion resources and IAM roles deployed in the same InfraPipeline:

```yaml
spec:
  imageSource:
    accessRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: ecr-access-role
        fieldPath: status.outputs.role_arn
  instanceRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: api-instance-role
      fieldPath: status.outputs.role_arn
  autoScalingConfigurationArn:
    valueFrom:
      kind: AwsAppRunnerAutoScalingConfiguration
      name: latency-sensitive-api
      fieldPath: status.outputs.configuration_arn
  vpcConnectorArn:
    valueFrom:
      kind: AwsAppRunnerVpcConnector
      name: private-backend-access
      fieldPath: status.outputs.vpc_connector_arn
  observabilityConfigurationArn:
    valueFrom:
      kind: AwsAppRunnerObservabilityConfiguration
      name: xray-tracing
      fieldPath: status.outputs.configuration_arn
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: service-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the companion resources and IAM roles first, then provisions the App Runner service with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring an App Runner service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Image vs. code source** -- Choose `imageSource` to deploy a pre-built container image from ECR or ECR Public. Choose `codeSource` to deploy from a repository through an App Runner connection, with the build configured in the spec (`configurationSource: API`) or read from an `apprunner.yaml` in the repo. Exactly one must be provided.

**Companions attach by reference, and presence IS the switch** -- `autoScalingConfigurationArn`, `vpcConnectorArn`, and `observabilityConfigurationArn` each reference a shared, versioned companion kind. Omitted at creation: AWS's account default scaling applies, egress is public-internet-only, and tracing is off. There are no separate enable toggles to keep in sync. Two one-way doors on a LIVE service (provider contracts): removing `autoScalingConfigurationArn` does not return the service to the account default (reference the default's ARN explicitly instead), and removing `observabilityConfigurationArn` does not disable tracing (an upstream provider gap -- disabling requires replacing the service).

**Private services get their ingress inline** -- set `isPubliclyAccessible: false` and declare `vpcIngressConnections`: each entry publishes the service into one VPC through an interface VPC endpoint for the App Runner requests service (AWS PrivateLink), and clients there resolve the per-connection private domain name exported in outputs. Ingress and egress are independent -- a fully VPC-internal service pairs `vpcIngressConnections` (inbound) with `vpcConnectorArn` (outbound).

**Instance sizing** -- Defaults to 1 vCPU (`cpu: "1024"`) and 2 GB memory (`memory: "2048"`). AWS accepts specific CPU/memory pairings only (e.g. 4 vCPU requires 8–12 GB); the console derives the legal memory choices from the selected CPU.

**Deterministic rollouts by default** -- `autoDeploymentsEnabled` is deliberately false: deployments happen when this resource is applied, so every rollout is recorded and graph-ordered. Enabling it makes App Runner watch the source (private ECR or code repositories only — AWS rejects it for ECR Public images).

**Custom domains export their DNS story** -- each `customDomains` entry gets a managed TLS certificate; the validation CNAME records and the service's DNS target export per domain in outputs, ready to compose into [AwsRoute53DnsRecord](/infra-catalog/aws-route53-dns-record) resources.

**KMS encryption is the one-way door** -- `kmsKeyArn` is ForceNew: changing it replaces the service. Everything else updates in place; source changes roll a new deployment.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** (optional) | `instanceRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (conditional) | `imageSource.accessRoleArn` | `status.outputs.role_arn` |
| **AwsAppRunnerAutoScalingConfiguration** (optional) | `autoScalingConfigurationArn` | `status.outputs.configuration_arn` |
| **AwsAppRunnerVpcConnector** (optional) | `vpcConnectorArn` | `status.outputs.vpc_connector_arn` |
| **AwsAppRunnerObservabilityConfiguration** (optional) | `observabilityConfigurationArn` | `status.outputs.configuration_arn` |
| **AwsKmsKey** (optional) | `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsWafWebAcl** (optional) | `webAclArn` | `status.outputs.web_acl_arn` |
| **AwsVpc** (per ingress connection) | `vpcIngressConnections[].vpcId` | `status.outputs.vpc_id` |
| **AwsVpcEndpoint** (per ingress connection) | `vpcIngressConnections[].vpcEndpointId` | `status.outputs.vpc_endpoint_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_arn` | Full ARN of the App Runner Service | IAM policies, CloudWatch alarms, deployment triggers |
| `service_id` | Unique identifier assigned by App Runner | Operational dashboards, event filtering |
| `service_url` | HTTPS URL for the service | Application configuration, DNS CNAME records |
| `service_name` | Computed name derived from metadata | Monitoring labels, log correlation |
| `service_status` | Current operational status (e.g., `RUNNING`) | Health monitoring, deployment verification |
| `custom_domains` | Per-domain DNS target, association status, and certificate-validation records | Route53 record composition for ownership proof |
| `vpc_ingress_connections` | Per-connection private domain name, ARN, and status | Internal DNS / service discovery inside the connected VPCs |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic public image** -- ECR Public image with default instance sizing and TCP health checks. The simplest path to an HTTPS endpoint for prototyping and internal tools. Start from the **Basic Public Image** preset.

**Production VPC-connected service** -- Private ECR image with a VPC connector for private egress, customer-managed KMS encryption, a tuned auto scaling configuration (warm instance pool), and HTTP health checks. The recommended production baseline. Start from the **Production VPC-Connected and Encrypted Service** preset.

**GitHub code source** -- Deploys directly from a repository using a managed runtime (Node.js, Python, Go, and more). App Runner handles the build-to-deploy lifecycle with no container image or CI/CD pipeline required. Start from the **GitHub Code Source (Node.js)** preset.

## Works With

- [**AWS App Runner Auto Scaling Configuration**](/infra-catalog/aws-app-runner-auto-scaling-configuration) -- the fleet's scaling posture, attached via `autoScalingConfigurationArn`
- [**AWS App Runner VPC Connector**](/infra-catalog/aws-app-runner-vpc-connector) -- private VPC egress, attached via `vpcConnectorArn`
- [**AWS App Runner Observability Configuration**](/infra-catalog/aws-app-runner-observability-configuration) -- X-Ray tracing, attached via `observabilityConfigurationArn`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the ECR access role for image pulling and the instance role for runtime AWS API access
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- customer-managed encryption for the stored deployment source
- [**AWS WAF Web ACL**](/infra-catalog/aws-waf-web-acl) -- request inspection in front of the endpoint via `webAclArn`
- [**AWS Route 53 DNS Record**](/infra-catalog/aws-route53-dns-record) -- the exported domain-validation and alias records' natural home
