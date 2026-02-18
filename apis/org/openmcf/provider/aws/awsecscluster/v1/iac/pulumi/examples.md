```markdown
# Create using CLI

Create a YAML file using one of the examples below. After the YAML is created, use the following command to apply with
OpenMCF (under the hood, you can choose Pulumi or Terraform):

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>

# or, if you prefer Terraform:

openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

If your environment is set up for shorthand, you might also use:

```shell
planton apply -f <yaml-path>
```

---

# Basic Example

A straightforward ECS cluster on Fargate with container insights enabled.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: basic-aws-ecs-cluster
  version:
    message: "First ECS cluster"
spec:
  region: us-east-1
  enable_container_insights: true
  capacity_providers:
    - FARGATE
  enable_execute_command: false
```

**Key Points**:

- **Container Insights**: Collects performance metrics through CloudWatch.
- **Fargate Only**: A fully serverless approach with no EC2 capacity management.

---

# Example with Multiple Capacity Providers

Enable both Fargate and Fargate Spot for a cost-optimized cluster.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: multi-capacity-aws-ecs-cluster
  version:
    message: "Using multiple capacity providers"
spec:
  region: us-east-1
  enable_container_insights: true
  capacity_providers:
    - FARGATE
    - FARGATE_SPOT
  enable_execute_command: false
```

**Key Points**:

- **Spot Usage**: FARGATE_SPOT reduces costs by leveraging available spot capacity.
- **Observability**: Continues to capture logs and metrics via CloudWatch.

---

# Example with ECS Exec Enabled

Allow secure shell access into running containers for debugging or operational tasks.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: exec-enabled-cluster
  version:
    message: "Debugging with ECS Exec"
spec:
  region: us-east-1
  enable_container_insights: true
  capacity_providers:
    - FARGATE
  enable_execute_command: true
```

**Key Points**:

- **ECS Exec**: Great for on-demand troubleshooting without full redeployment.
- **Fargate**: Maintains a serverless experience while still allowing in-container access.

---

# Minimal Example

Omitting optional fields; this assumes default settings such as `enable_container_insights = true` from the recommended
defaults.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: minimal-cluster
  version:
    message: "Minimal ECS cluster example"
spec:
  region: us-east-1
# No capacity_providers or ECS Exec enabled, which means basic defaults apply
```

**Key Points**:

- **Minimal Config**: Uses recommended defaults (e.g., container insights enabled).
- **Easy Starting Point**: Helpful when you just need a cluster and plan to refine later.

---

**Next Steps**:

- Adjust the `capacity_providers`, `enable_container_insights`, and `enable_execute_command` fields to suit your
  operational needs.
- Refer to the [README.md](./README.md) for deeper configuration details.
- Use the OpenMCF CLI to validate and deploy these manifests across multiple environments.
