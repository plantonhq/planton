# Azure Planton Runner

Deploys a standing Planton runner appliance on Azure Container Apps -- an always-on, outbound-only worker that receives deploy operations from the control plane and executes them from within your network perimeter. It is the piece that makes private endpoints (most notably private AKS API servers) deployable and operable, with the resource group and the Container App Environment wired through ValueFromRef. Enrollment is token-first: the runner joins with a token, registers itself, and receives its own individually revocable identity.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Container App** -- the runner container itself, a single-revision app pinned to exactly one replica (min = max = 1), with the runner token in the app's own secret store (`runner-token`) referenced as a secret-backed environment variable, no ingress at all, and a startup probe on the runner's health server
- **Azure tags** -- resource tags derived from the resource's organization, environment, and name

The resource group and the Container App Environment are **referenced prerequisites, not created**: the module never creates or mutates them. The environment decides the network boundary -- a VNet-integrated environment gives the runner reach into that network's private endpoints.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- the credential used to provision the appliance itself. Required.
- **Runner token** -- nothing to create by hand: the platform mints a runner token and writes it at exactly the managed-secret reference the manifest declares, before the infrastructure applies. Choose a secret slug and reference it as `$secret/<slug>` in the `token` field; never inline plaintext. The token authorizes joining and is never the runner's identity -- revoking it never touches runners it already admitted.

### Azure Subscription

- **Resource group** -- where the app lives. Reference an `AzureResourceGroup` via ValueFromRef or pass a literal name.
- **Container App Environment** -- the placement decision: the environment decides what the runner can reach. Pick (or compose) a VNet-integrated environment when the runner must deploy to private endpoints (a private AKS API server, a private-endpoint database). Reference an `AzureContainerAppEnvironment` via ValueFromRef or pass a literal environment resource ID.

## Deploy

### Console

Open the deployment store, find **Azure Planton Runner**, and click **Deploy**. The creation wizard walks you through environment and connection configuration and the spec fields.

### CLI

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePlantonRunner
metadata:
  name: aks-runner
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: my-resource-group
  containerAppEnvironmentId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-resource-group/providers/Microsoft.App/managedEnvironments/my-environment
  token: $secret/aks-runner-token
```

```shell
planton apply -f runner.yaml
```

This minimal manifest deploys a single always-on worker at the default Consumption-plan sizing (0.5 vCPU, 1Gi) tracking the latest runner release -- sizing, version pinning, and the control-plane endpoint are not configured. An Infra Job tracks the provisioning in real time.

### InfraChart

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: private-env
      fieldPath: status.outputs.environment_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and the environment first, then provisions the runner with the resolved values.

## Key Configuration

These are the most important decisions when configuring the runner. Explore the full field reference in the [API Explorer](#api-explorer) tab.

- **Environment placement** -- `containerAppEnvironmentId` is the whole point of the appliance: the environment decides the network boundary, so a VNet-integrated environment is what gives the runner reach into private endpoints. The runner itself carries no network knobs -- change the reach by changing (or recomposing) the environment.
- **Sizing** -- `cpu` must be one of the Consumption-plan sizes (0.25 through 2 vCPUs), and `memory` is fixed at cpu x 2 (0.5/1Gi, 1/2Gi, ...). Azure would reject an invalid pairing only at deploy time, so the spec validates it up front. Memory pressure shows up as failed operations mid-apply; when in doubt, size the pairing up.
- **Runner build** -- empty `runnerVersion` tracks the newest release on every replica (re)start; pin a version tag for change control. `imageRepository` is only for air-gapped or mirrored registries hosting a digest-identical copy.
- **Control plane endpoint** -- leave `controlPlaneEndpoint` unset for Planton's hosted control plane; set host:port for a self-hosted instance.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureContainerAppEnvironment** | `containerAppEnvironmentId` | `status.outputs.environment_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `container_app_id` | The Azure resource ID of the Container App | Inspecting the appliance with Azure tooling |
| `container_app_name` | The Container App's name | Console and CLI lookups |
| `token_secret_name` | The Container App secret holding the runner token | Auditing secret configuration; rotation tooling |
| `runner_name` | The name the runner registers itself under | Finding the runner in `planton runner list` |
| `resource_group_name` | The resource group the runner was deployed in | Targeting follow-up Azure operations correctly |

## Common Patterns

**One runner per network perimeter** -- deploy a runner only when a target is invisible from outside the network (a private AKS API server, a private-endpoint database); if every endpoint is public, Planton's hosted runner fleet already covers you, and a self-hosted runner buys nothing except standing infrastructure to size, pin, and pay for. One runner covers a VNet, not a workload: whatever the environment's network can reach, the runner can deploy to, so deploy one per VNet rather than one per cluster. Start from the **Environment Runner** preset.

**Production hardening with a pinned version and full sizing** -- pin `runnerVersion` so nothing tracks the latest release (upgrades and rollbacks become deliberate re-pins), and take the largest Consumption pairing (2 vCPU / 4Gi) when stacks are large or operations run concurrently -- memory pressure shows up as failed IaC operations mid-apply, and the pairing law means sizing memory up always means sizing CPU with it. Start from the **High Capacity (Production Hardened)** preset.

**Calm token rotation** -- the token is only read at join, so rotating it interrupts nothing: the running replica keeps serving on its minted identity, and the next replica replacement joins with the new value. Revoke the runner's own identity (not the token) to cut off a runner; revoking the token never touches runners it already admitted.

**Destroy the runner last** -- the runner is the deploy path for the private workloads behind it. Tear down in-cluster workloads through the runner, then the cluster over the Azure path, then the runner at the end; destroying the runner first strands everything it deploys, because nothing else can reach the private endpoints to tear them down. The environment and resource group are referenced, never owned, so destroying the runner never disturbs neighbors sharing them.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where the appliance lives; the teardown boundary for everything in it
- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- the placement that defines what the runner can reach; VNet-integrate it for private endpoints
- [**Azure AKS Cluster**](/infra-catalog/azure-aks-cluster) -- the canonical private target: a private API server the runner makes deployable
