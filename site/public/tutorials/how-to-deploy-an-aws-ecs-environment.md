---
title: "How to Deploy an AWS ECS Environment"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "aws"
  - "ecs"
  - "infra-chart"
  - "vpc"
  - "infra-catalog"
category: "aws"
excerpt: "Deploy a complete AWS ECS environment -- VPC, load balancer, container registry, and ECS cluster -- from a single Infra Chart using the Planton CLI."
---

# How to Deploy an AWS ECS Environment

A production-ready ECS environment on AWS requires a VPC with subnets across multiple availability zones, a NAT gateway, security groups, an Application Load Balancer, an ECS cluster with Fargate capacity, an ECR repository for container images, and IAM roles for task execution. Getting the dependency ordering right -- the ALB needs the VPC, the ECS service needs the ALB and the cluster and the IAM role -- adds complexity on top of the resource count.

This tutorial deploys all of it from a single Infra Chart with one command. You will customize a handful of parameters, run `planton chart install`, and watch the platform provision seven interdependent AWS resources in the correct order. By the end, you will have a running ECS environment with a sample nginx service accessible through the ALB.

> **Note**: The Planton web console provides a guided creation wizard for Infra Charts
> and other Infra Components. This tutorial uses the CLI/YAML approach for stability
> and reproducibility. The console UI evolves frequently — always check it for the
> latest experience.

## What You Will Learn

- What Infra Charts are and how they differ from deploying individual Infra Components
- How to install the AWS ECS Environment chart with customized parameters
- How the dependency graph determines deployment order and enables parallel execution
- How to monitor a multi-resource deployment through an Infra Pipeline
- How to verify the deployed environment and access the sample service
- How to tear down the environment when you are done

## Prerequisites

- [ ] An AWS provider connection configured and set as the default for your target environment (see [How to Connect Your AWS Account to Planton](/tutorials/how-to-connect-your-aws-account-to-planton))
- [ ] A Planton organization and at least one environment created
- [ ] The `planton` CLI installed and authenticated (`planton auth login`)
- [ ] `git` installed (to clone the chart repository)
- [ ] For the optional DNS/HTTPS enhancement: a domain name you control and willingness to create a Route 53 hosted zone

The AWS provider connection must have permissions to create VPCs, subnets, NAT gateways, security groups, ALBs, ECS clusters, ECS services, ECR repositories, and IAM roles. If you used the recommended IAM policy from the AWS connection tutorial, these permissions are already included.

## How Infra Charts Work

An Infra Chart bundles multiple Infra Components into a single deployable unit with a dependency graph -- similar to a Helm chart for infrastructure. You install it with `planton chart install`, customize it with a values file, and Planton provisions all resources in the correct order through an Infra Pipeline. For more on Infra Charts, see the [Infra Charts documentation](/docs/infrastructure/infra-charts).

## Step 1: Clone the Chart Repository

The AWS ECS Environment chart is maintained in the `plantonhq/infra-charts` repository. Clone it to your local machine:

```bash
git clone https://github.com/plantonhq/infra-charts.git
```

The chart lives at `infra-charts/aws/ecs-environment/` with this structure:

```text
aws/ecs-environment/
├── Chart.yaml          # Chart metadata (name, description, icon)
├── values.yaml         # Parameters with default values
├── templates/          # Infra Component templates with Jinja variables
│   ├── network.yaml        # VPC, security group, ALB
│   ├── ecs-cluster.yaml    # ECS cluster with Fargate
│   ├── ecs-service.yaml    # ECS service and task definition
│   ├── ecr-repo.yaml       # Container registry
│   └── task-execution-iam-role.yaml  # IAM role for task execution
└── README.md
```

The chart deploys these seven AWS resources:

| Resource | Kind | Purpose |
|---|---|---|
| VPC | `AwsVpc` | Isolated network with public and private subnets across 2 AZs, NAT gateway |
| Security group | `AwsSecurityGroup` | Allows HTTP/HTTPS inbound traffic and all outbound |
| ECR repository | `AwsEcrRepo` | Private container registry for your service images |
| ECS cluster | `AwsEcsCluster` | Fargate and Fargate Spot capacity for running containers |
| ALB | `AwsAlb` | Application Load Balancer for routing traffic to ECS services |
| IAM role | `AwsIamRole` | Task execution role with permissions for image pulling and logging |
| ECS service | `AwsEcsService` | Running service with a sample nginx container behind the ALB |

## Step 2: Customize the Values

Copy the default values file to create your own configuration:

```bash
cp infra-charts/aws/ecs-environment/values.yaml my-values.yaml
```

Edit `my-values.yaml` with your settings:

```yaml
params:
  - name: availability_zone_1
    description: First AZ for the public / private subnet pair
    value: us-east-1a

  - name: availability_zone_2
    description: Second AZ for the public / private subnet pair
    value: us-east-1b

  - name: domain_name
    description: Route 53 Hosted-Zone domain
    value: example.com

  - name: load_balancer_domain_name
    description: DNS name served by the ALB
    value: app.example.com

  - name: service_name
    description: ECS service & task family name
    value: nginx

  - name: service_image_repo_name
    description: ECR repository for your service images
    value: my-app

  - name: service_port
    description: Container port the task listens on
    value: "80"

  - name: dnsEnabled
    description: Create Route53 zone and configure ALB DNS records
    type: bool
    value: false

  - name: httpsEnabled
    description: Create ACM cert and terminate TLS on the ALB (requires dnsEnabled)
    type: bool
    value: false

  - name: alb_idle_timeout_seconds
    description: ALB idle timeout
    value: "60"
```

Here is what to customize:

- **`availability_zone_1` and `availability_zone_2`**: Set these to two AZs in the AWS region your provider connection targets. The VPC creates public and private subnets in each AZ.
- **`service_image_repo_name`**: The name for the ECR repository where you will push container images. Choose a name that matches your application (e.g., `checkout-service`, `api-gateway`).
- **`service_name`**: The name for the ECS service and task family. The chart deploys a sample nginx container -- you will replace this with your own application later.
- **`service_port`**: The port your container listens on. The default `80` matches the sample nginx container.
- **`dnsEnabled: false`** and **`httpsEnabled: false`**: For this tutorial, DNS and HTTPS are disabled. The ALB gets an AWS-generated DNS name that you can use immediately without owning a domain. The optional section later in this tutorial covers enabling DNS and HTTPS.
- **`domain_name`** and **`load_balancer_domain_name`**: Ignored when `dnsEnabled` is `false`. You can leave the placeholder values.

## Step 3: Install the Chart

Run the following command to create the Infra Stack and trigger the deployment pipeline:

```bash
planton chart install my-ecs-env \
  ./infra-charts/aws/ecs-environment \
  --org your-org \
  --env production \
  -f my-values.yaml
```

Replace `your-org` with your Planton organization slug and `production` with your target environment slug.

The command creates an Infra Stack named `my-ecs-env` from the chart, renders the templates with your parameter values, builds the dependency graph, and triggers an Infra Pipeline to deploy all seven resources. The output includes the Infra Stack details and a console URL for monitoring:

```text
infra-stack 'my-ecs-env' applied

  Follow live: https://planton.ai/your-org/infra-stack/my-ecs-env?ipid=infpipe_...
```

Open the console URL to see the deployment pipeline in real time, including the DAG visualization showing which resources are deploying and which are waiting on dependencies.

## Step 4: Monitor the Pipeline

The console provides the richest monitoring experience -- a live DAG visualization showing each resource's status, with logs accessible by clicking on individual nodes.

From the CLI, you can stream the pipeline status:

```bash
planton infra infra-pipeline stream-status <infra-pipeline-id>
```

Use the pipeline ID from the install output (the `infpipe_...` value in the console URL).

The pipeline deploys resources in dependency order across four layers:

1. **Layer 0** (parallel, ~2-5 minutes): VPC, ECR repository, ECS cluster, and IAM role deploy simultaneously. These resources have no dependencies on each other.
2. **Layer 1** (~1-2 minutes): The security group deploys after the VPC completes, because it references the VPC ID.
3. **Layer 2** (~2-3 minutes): The ALB deploys after both the VPC (for subnets) and the security group complete.
4. **Layer 3** (~3-5 minutes): The ECS service deploys last. It depends on the ALB (for routing), the ECS cluster (for capacity), the IAM role (for permissions), the VPC (for subnets), and the security group (for network access).

Total deployment time is typically 10-20 minutes. The parallel execution in Layer 0 saves significant time compared to deploying each resource sequentially.

## Step 5: Verify the Environment

After the pipeline completes, verify the individual Infra Components that were created. Each resource has its own status and outputs.

To find the ALB's DNS name (which you need to access the sample service):

```bash
planton get AwsAlb production-alb -o yaml
```

Look for `status.outputs.loadBalancerDnsName` in the output (`planton get` prints camelCase keys, like every manifest). This is the AWS-generated DNS name for the load balancer, something like `production-alb-1234567890.us-east-1.elb.amazonaws.com`.

Verify the sample nginx service is running:

```bash
curl http://<load-balancer-dns-name>
```

You should see the default nginx welcome page. This confirms the full chain is working: the ALB is routing traffic to the ECS service, which is running the nginx container in a Fargate task, inside the VPC you provisioned.

To inspect other resources:

```bash
planton get AwsVpc production-vpc -o yaml
planton get AwsEcsCluster production-ecs-cluster -o yaml
planton get AwsEcrRepo ecr-repo -o yaml
```

The resource names follow the pattern `{env}-{resource-type}` as defined in the chart templates. The environment slug (`production` in this example) is injected by the Infra Stack.

## Adding DNS and HTTPS (Optional)

If you have a domain name and want production-grade TLS termination on the ALB, update your values file to enable DNS and HTTPS:

```yaml
  - name: domain_name
    value: yourdomain.com

  - name: load_balancer_domain_name
    value: app.yourdomain.com

  - name: dnsEnabled
    type: bool
    value: true

  - name: httpsEnabled
    type: bool
    value: true
```

Re-run the install command with the same stack name to update:

```bash
planton chart install my-ecs-env \
  ./infra-charts/aws/ecs-environment \
  --org your-org \
  --env production \
  -f my-values.yaml
```

Using the same name (`my-ecs-env`) updates the existing Infra Stack rather than creating a new one. The pipeline will create two additional resources:

- **AwsRoute53Zone**: A hosted zone for your domain. After creation, you need to update your domain registrar's nameservers to point to the Route 53 nameservers in the zone's outputs.
- **AwsCertManagerCert**: A DNS-validated ACM certificate for your load balancer domain. ACM creates a CNAME validation record in the Route 53 zone automatically. Certificate validation can take a few minutes.

The ALB is updated with the certificate and DNS configuration. Once the certificate validates and DNS propagates, your ECS service is accessible at `https://app.yourdomain.com`.

**Important**: DNS validation requires the Route 53 zone's nameservers to be authoritative for your domain. If you are using a new domain, update nameserver delegation at your registrar before enabling HTTPS. If validation fails, the ACM certificate resource will report the issue in its Infra Job logs.

## Tearing Down the Environment

When you are done, tear down the environment to stop incurring AWS charges. The platform offers two options:

**Uninstall** destroys all Infra Components but keeps the Infra Stack record in Planton. This is useful if you want to redeploy later with the same configuration:

```bash
planton chart uninstall my-ecs-env
```

**Purge** destroys all Infra Components AND deletes the Infra Stack from the database:

```bash
planton chart purge my-ecs-env
```

Both commands trigger an undeploy pipeline that destroys resources in reverse dependency order -- the ECS service is removed first, then the ALB and security group, then the VPC, cluster, ECR, and IAM role. The console URL printed by the command lets you monitor the teardown progress.

**Cost awareness**: While deployed, this environment incurs charges for the VPC (NAT gateway), ALB, ECS Fargate tasks, and ECR storage. The NAT gateway and ALB are the largest fixed costs. For non-production use, tear down the environment when you are not actively using it.

## What to Do Next

Your AWS ECS environment is running. From here:

- **Deploy your own application** by pushing a container image to the ECR repository and updating the ECS service to use it instead of the sample nginx image. The ECR repository name is in the `AwsEcrRepo` outputs.
- **Deploy a backend service through Service Hub** that targets this ECS cluster. See [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd) for the complete push-to-deploy workflow.
- **Explore other Infra Charts** in the [plantonhq/infra-charts](https://github.com/plantonhq/infra-charts) repository. Similar charts exist for AWS EKS environments, GCP GKE environments, and other common infrastructure patterns.
- **Read the full Infra Charts documentation** at [Infra Charts](/docs/infrastructure/infra-charts) to learn how to create custom charts for your organization's infrastructure patterns.
