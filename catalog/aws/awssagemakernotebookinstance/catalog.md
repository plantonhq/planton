# AWS SageMaker Notebook Instance

Deploys an Amazon SageMaker notebook instance — a managed EC2 workstation running Jupyter, bootstrapped by declarative lifecycle scripts and billed hourly whether or not a notebook is open. The spec folds the lifecycle configuration into the instance (plain shell in `onCreate` and `onStart`; the modules base64-encode before sending), attaches an ML storage volume, and optionally confines the notebook to your VPC with direct internet access disabled. Most changes ride a stop-update-start cycle that takes the notebook offline for several minutes, and shrinking the volume replaces the instance — the sizing and network decisions deserve to be made once.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Notebook Instance Lifecycle Configuration** — created only when `lifecycleConfig` is set; carries the `onCreate` (runs once) and `onStart` (runs every start) scripts under a stable derived name (`<name>-lifecycle`), run as root under AWS's 5-minute limit
- **SageMaker Notebook Instance** — the Jupyter workstation named from `metadata.name`, on an `ml.*` instance type with a 5–16384 GB storage volume, optional VPC placement, KMS volume encryption, root-access and IMDSv2 lockdown, platform selection, and up to four Git repositories cloned into the working directory

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreateNotebookInstance` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `sagemaker.amazonaws.com`, wired via `roleArn` — the notebook assumes it for every AWS call made from Jupyter, so its grants define what users can reach.
- For VPC confinement (`directInternetAccess: Disabled`): a subnet and security groups, plus a NAT or VPC-endpoint path so the notebook can still reach SageMaker APIs (only for that posture).

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Notebook Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region, instance type, and role, and the optional bootstrap scripts and VPC placement. Start from the **Starter Notebook** preset in the [Presets](#presets) tab for a ready workstation, or the **Locked-Down GPU Notebook** preset for the hardened GPU shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerNotebookInstance
metadata:
  name: analysis-notebook
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  instanceType: ml.t3.medium
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sagemaker-notebook-role
      fieldPath: status.outputs.role_arn
  volumeSizeGb: 50
  lifecycleConfig:
    onCreate: |
      #!/bin/bash
      set -e
      pip install --quiet pandas scikit-learn matplotlib
```

```shell
planton apply -f notebook-instance.yaml
```

This creates an ml.t3.medium Jupyter workstation with a 50 GB volume, bootstrapped once with the listed Python libraries. An Infra Job tracks the provisioning in real time.

### InfraChart

When the notebook deploys alongside its role and VPC wiring in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  instanceType: ml.t3.medium
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sagemaker-notebook-role
      fieldPath: status.outputs.role_arn
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: private-subnet-a
      fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: notebook-sg
        fieldPath: status.outputs.security_group_id
  directInternetAccess: Disabled
```

The InfraPipeline resolves the dependency graph, creates the role, subnet, and security group first, then the confined notebook.

## Key Configuration

These are the most important decisions when configuring a notebook instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The meter runs while it exists** — `instanceType` sets an hourly rate that bills whether or not anyone has Jupyter open. ml.t3.medium is the cheapest current-generation choice for exploration; stop or delete idle notebooks rather than letting them idle at GPU rates.

**Never shrink the volume** — Growing `volumeSizeGb` updates in place; shrinking replaces the instance and everything on the volume. Size generously up front — the storage is cheap relative to a rebuilt workspace.

**Decide network posture up front** — `subnetId`, `directInternetAccess`, and `platformIdentifier` changes all replace the instance. `Disabled` internet access requires the subnet and security groups (enforced at manifest time) plus a NAT or endpoint path for SageMaker API calls — wire that path before flipping the switch, or training jobs launched from the notebook will hang.

**Keep bootstrap scripts under five minutes** — Lifecycle scripts run as root at create/start, and a script exceeding AWS's limit fails the instance start. Push long installs to the background (`nohup … &`) or bake a custom environment. One provider quirk: clearing a script in the spec does NOT clear it in AWS — replace the text (even with a no-op `#!/bin/bash`) when retiring behavior.

**Batch your changes** — Most updates require a Stopped instance; the modules ride the stop-update-start choreography and the notebook is unavailable through it. Trickling one change at a time multiplies the downtime.

**Lock down what users don't need** — `rootAccess: Disabled` keeps users out of root (lifecycle scripts still run as root); `imdsMinimumVersion: "2"` is the hardened metadata-service choice. Prefer `notebook-al2-v3` or `notebook-al2023-v1` — the older platforms are deprecated, kept only for existing workloads.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsSubnet** | `subnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `securityGroupIds[]` | `status.outputs.security_group_id` |
| **AwsKmsKey** | `kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `url` | URL that opens the Jupyter notebook | Bookmarks, team portals, onboarding docs |
| `notebook_instance_arn` | Amazon Resource Name of the instance | IAM policies scoping start/stop access |
| `network_interface_id` | The ENI SageMaker created in your subnet (VPC notebooks only) | Network debugging; flow-log filters |

`notebook_instance_name` and `lifecycle_config_name` are also exported — they echo the derived names and are audit values rather than composition inputs.

## Common Patterns

**Starter data-science workstation** — ml.t3.medium, a modest volume, and an `onCreate` script installing the team's Python stack: a ready Jupyter environment at the cheapest current-generation rate. Start from the **Starter Notebook** preset.

**Locked-down GPU notebook** — a GPU instance type confined to a VPC with `directInternetAccess: Disabled`, `rootAccess: Disabled`, IMDSv2 enforced, and a KMS-encrypted volume. The posture for notebooks touching regulated data; budget the NAT/endpoint path first. Start from the **Locked-Down GPU Notebook** preset.

**Repository-anchored notebooks** — `defaultCodeRepository` clones a Git repository as the working directory (plus up to three `additionalCodeRepositories`), so the notebook's state of record lives in Git, not on the instance volume — which is what makes the instance safely disposable.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role Jupyter's AWS calls run as, wired via `roleArn`
- [**AWS Subnet**](/infra-catalog/aws-subnet) — VPC placement for private notebooks, wired via `subnetId`
- [**AWS Security Group**](/infra-catalog/aws-security-group) — the notebook ENI's security groups, wired via `securityGroupIds`
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for the ML storage volume
- [**AWS SageMaker Image**](/infra-catalog/aws-sagemaker-image) — custom kernel images for teams that outgrow lifecycle-script bootstrapping
