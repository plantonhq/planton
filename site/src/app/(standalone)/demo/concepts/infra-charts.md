# Infra Charts: Orchestrated Infrastructure Bundles

## The Problem

You've deployed an AWS ALB using a Lego Block. Great! But production environments aren't just one resource.

**An ECS environment needs**:
- VPC with public and private subnets (2 AZs)
- Security groups
- Application Load Balancer
- SSL certificate from ACM
- Route53 DNS zone
- ECS cluster
- IAM task execution role
- ECR repository
- And all of these must be **connected correctly**

**The challenge**: Manually deploying and connecting 9+ resources is:
- Complex (what depends on what?)
- Error-prone (easy to misconfigure connections)
- Time-consuming (weeks of work)
- Hard to replicate (copy-paste Terraform across projects)

## The Inspiration: Kubernetes Helm Charts

We didn't invent this pattern. Kubernetes solved a similar problem years ago.

**Kubernetes Helm Charts**:
- Package multiple Kubernetes resources together
- Parameterized templates (same chart, different values)
- Reusable across applications
- Community-driven library of charts

**Infra Charts**:
- Package multiple Lego Blocks together
- Parameterized templates (same chart, different values)
- Reusable across projects and clients
- Library of production-ready patterns

Same concept, different domain: **Applications → Infrastructure**

## What Infra Charts Are

Infra Charts are **reusable templates** that package multiple Lego Blocks into complete, production-ready environments.

### Structure of a Chart

Each chart contains:

1. **Template files** (YAML with Jinjava templating):
   ```yaml
   kind: AwsVpc
   metadata:
     name: "{{ values.env }}-vpc"
   spec:
     vpcCidr: "{{ values.vpc_cidr }}"
     availabilityZones:
       - "{{ values.az_1 }}"
       - "{{ values.az_2 }}"
   ```

2. **values.yaml** (parameters and defaults):
   ```yaml
   - name: env
     description: Environment name (dev, staging, prod)
     value: dev
   
   - name: vpc_cidr
     description: VPC CIDR block
     value: 10.0.0.0/16
   ```

3. **Chart metadata** (description, icon, links)

### How Charts Work

1. **Select a chart**: "AWS ECS Quick Start with HTTPS"
2. **Provide values**: Domain name, regions, resource sizes
3. **System renders templates**: Replaces `{{ values.x }}` with your values
4. **Generates complete manifest**: All resources with correct dependencies
5. **Builds DAG**: Determines deployment order
6. **Orchestrates deployment**: Creates resources in topological order

## The DAG: Directed Acyclic Graph

Charts automatically build a dependency graph:

**Example - ECS Chart DAG**:
```
VPC (no dependencies - deploys first)
  ↓
Security Group (depends on VPC)
  ↓
Route53 Zone (no dependencies - deploys in parallel)
  ↓
Certificate (depends on Route53)
  ↓
ALB (depends on VPC, Security Group, Certificate)
  ↓
ECS Cluster (no dependencies - deploys in parallel)
  ↓
IAM Role (no dependencies - deploys in parallel)
  ↓
ECS Service (depends on ALB, ECS Cluster, IAM Role)
```

The platform:
- Validates all dependencies exist
- Deploys independent resources in parallel
- Waits for dependencies before proceeding
- Shows live progress for each resource
- Handles failures gracefully

## Real Chart Example: AWS ECS with HTTPS

**What it creates**:
1. VPC with public/private subnets across 2 AZs
2. Security group allowing HTTP/HTTPS
3. Route53 hosted zone for your domain
4. ACM certificate (DNS validated)
5. Application Load Balancer with SSL
6. ECS cluster (Fargate/Fargate Spot)
7. IAM role for task execution
8. ECR repository for images
9. Sample ECS service connected to ALB

**Parameters you provide** (7 values):
- `availability_zone_1`: us-east-1a
- `availability_zone_2`: us-east-1b
- `domain_name`: example.com
- `load_balancer_domain_name`: app.example.com
- `service_name`: nginx
- `service_image_repo_name`: shopping-cart-service
- `service_port`: 8080

**Result**: Production-ready ECS environment in ~20 minutes

## Curated Library of Charts

We provide pre-built charts for common scenarios:

- **AWS ECS Quick Start**: Complete ECS environment with HTTPS
- **GCP GKE Baseline**: Private GKE cluster with VPC, NAT, node pools
- **DigitalOcean App Platform**: App platform environment with domain
- **AWS Terraform State Backend**: S3 + DynamoDB for Terraform state
- **AWS Pulumi State Backend**: S3 for Pulumi state

Each chart is:
- Production-tested with real customers
- Open source in our infra-charts repository
- Documented with README and examples
- Continuously improved based on feedback

## Custom Charts for Organizations

Companies can create custom charts to encode their own patterns:

**Example Use Case**: Your company always deploys GKE clusters with:
- Specific network policies
- Workload identity enabled
- Standard monitoring setup
- Security baselines
- Compliance requirements

**Solution**: Create a custom chart once, use it for all projects. Your teams deploy compliant, secure clusters every time without remembering the details.

## Templating Engine

Charts use **Jinjava** (HubSpot's Java implementation of Jinja2):

**Supports**:
- Variable substitution: `{{ values.name }}`
- Conditionals: `{% if values.ssl_enabled %}...{% endif %}`
- Loops: `{% for az in values.azs %}...{% endfor %}`
- Filters: `{{ values.name | upper }}`

**Safety**:
- Sandboxed execution
- Disabled dangerous operations
- Size limits
- Depth limits

## From Chart to Deployment

The journey of a chart deployment:

1. **InfraChart**: The template definition (stored in platform)
2. **InfraProject**: Chart + your values (rendered manifest + DAG)
3. **InfraPipeline**: Execution orchestration (triggered deployment)
4. **CloudResources**: Individual resources created from DAG nodes
5. **StackJobs**: One per resource, runs Terraform/Pulumi
6. **Infrastructure**: Actual resources on cloud provider

**Separation of concerns**:
- Configuration (chart + values) separate from execution (pipeline)
- Store templates once, deploy many times
- Version-controlled, auditable

## Value Demonstration in Demo Flow

**Why we show individual Lego Blocks first**:
1. Users understand the concept of one component
2. They see the form-based configuration
3. They watch live deployment
4. They understand what a "block" is

**Then we introduce charts**:
1. Explain the problem: "Real environments need many interconnected blocks"
2. Show the inspiration: Helm Charts for Kubernetes
3. Present the solution: Infra Charts
4. Show real example: Large ECS DAG visualization (9 resources with dependencies)
5. Demonstrate deployment: Form (7 params) → DAG (9 resources) → Live progress

**The Chart Deployment Screen** (Side-by-Side Layout):

**Left (40%) - Configuration Form**:
- Simple values.yaml parameters (7 fields)
- Availability zones, domain, service name, ECR repo, port
- Educational header explaining template rendering
- Deploy Chart button

**Right (60%) - DAG Visualization**:
- Shows all 9 resources and their dependencies
- Visual preview of what will be deployed
- User understands the orchestration before clicking Deploy
- When deployed: Stats appear (9 resources, ~20 min, in progress)

**Why This Layout Works**:
- **Visual impact**: 7 simple inputs → 9 complex resources
- **Transparency**: User sees exactly what will be created
- **Education**: Dependencies are visible and understandable
- **Credibility**: Real DAG from actual product

**Key Technical Points**:
- Each DAG node = independent Terraform/Pulumi execution (StackJob)
- Resources deploy when dependencies are satisfied
- Parallel execution where possible
- ~20 minutes to production-ready environment

**Perfect Segue**: "Infrastructure is ready. Now let's deploy services to it via Service Hub."

This progression makes the value immediately clear and shows the actual product experience.

## Customer Impact

**IT Consulting Agency** (paying customer):
- Deployed same ECS chart for 3 different clients
- Each client: different domain, different config values
- Same reliable infrastructure pattern every time
- Client onboarding time: **3 weeks → 20 minutes**
- Time to production: **<1 hour** from code to live URL

**Small Product Company**:
- Developers deploy complete GKE environments
- No DevOps team needed
- Self-service via forms
- Production-ready defaults

## Open Source Charts Repository

All curated charts are open source:

**Repository**: [github.com/plantonhq/infra-charts](https://github.com/plantonhq/infra-charts)

Each chart includes:
- `Chart.yaml` (metadata)
- `values.yaml` (parameters with defaults)
- `templates/` (YAML files with templating)
- `README.md` (documentation and examples)

Organizations can:
- Use as-is
- Fork and customize
- Contribute improvements
- Create pull requests

## Comparison: Individual vs Charts

| Aspect | Individual Lego Blocks | Infra Charts |
|--------|----------------------|---------------------|
| **Complexity** | Single resource | Multiple interconnected resources |
| **Use Case** | Quick additions, experiments | Complete environments |
| **Configuration** | One form | Parameterized template |
| **Deployment** | Direct StackJob | Orchestrated pipeline with DAG |
| **Time** | 1-5 minutes | 15-25 minutes |
| **Reusability** | One-off | Reuse across projects |

## When to Use Infra Charts

**Use Charts when**:
- Deploying complete environments (dev, staging, prod)
- Standardizing infrastructure across teams
- Onboarding new clients (consulting agencies)
- Need reproducible deployments
- Want orchestration and dependency management

**Use Individual Lego Blocks when**:
- Adding one resource to existing infrastructure
- Quick experiments or prototyping
- Granular control needed
- Simple, standalone resources

## The Future: Community Charts

We envision a future where:
- Organizations share their charts publicly
- Community contributes domain-specific patterns (FinTech, HealthTech)
- Best practices encoded and distributed
- New teams bootstrap from proven patterns

Infra Charts transform infrastructure from "write it yourself every time" to "reuse proven patterns with your values."

## Technical Details

**Templating**: Jinjava (HubSpot's Java implementation of Jinja2)
- Variable substitution: `{{ values.name }}`
- Conditionals and loops supported
- Sandboxed for security

**DAG Generation**: Automatic from ValueFromRef dependencies
- Platform analyzes all resource definitions
- Builds dependency graph automatically
- Determines topological deployment order
- Enables parallel execution

**Deployment Orchestration**: InfraPipeline
- Creates CloudResource for each DAG node
- Each CloudResource triggers a StackJob
- StackJobs execute in dependency order
- Platform tracks overall progress

## Related Concepts

- [Deployment Component Store](./deployment-store.md) - Individual Lego Blocks
- [Full Terraform Parity](./terraform-parity.md) - Why the catalog targets the full provider surface

