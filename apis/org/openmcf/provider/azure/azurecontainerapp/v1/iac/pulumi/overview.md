# AzureContainerApp Pulumi Module: Architecture Overview

## Resource Graph

The AzureContainerApp module creates a single resource:

```
AzureContainerApp
└── containerapp.App (azurerm_container_app)
```

This is a single-resource component. The complexity lives in the rich spec-to-args translation: 21 protobuf message types are mapped into the deeply nested `containerapp.AppArgs` structure covering containers, probes, scale rules, ingress, secrets, registries, Dapr, and identity.

## Data Flow

```
AzureContainerAppStackInput
├── target.metadata    → Azure tags (resource, resource_name, resource_kind, org, env)
├── target.spec.resource_group → locals.ResourceGroupName (via .GetValue())
├── target.spec.name   → AppArgs.Name
├── target.spec.container_app_environment_id → AppArgs.ContainerAppEnvironmentId (via .GetValue())
├── target.spec.revision_mode → AppArgs.RevisionMode (default: "Single")
├── target.spec.workload_profile_name → AppArgs.WorkloadProfileName (optional)
├── target.spec.max_inactive_revisions → AppArgs.MaxInactiveRevisions (optional)
├── target.spec.containers → AppArgs.Template.Containers (via buildContainers)
│   ├── .env → AppTemplateContainerEnvArgs (via buildEnvVars)
│   ├── .liveness_probe → AppTemplateContainerLivenessProbeArgs (via buildLivenessProbe)
│   ├── .readiness_probe → AppTemplateContainerReadinessProbeArgs (via buildReadinessProbe)
│   ├── .startup_probe → AppTemplateContainerStartupProbeArgs (via buildStartupProbe)
│   └── .volume_mounts → AppTemplateContainerVolumeMountArgs (via buildVolumeMounts)
├── target.spec.init_containers → AppArgs.Template.InitContainers (via buildInitContainers)
├── target.spec.volumes → AppArgs.Template.Volumes (via buildVolumes)
├── target.spec.min_replicas → AppArgs.Template.MinReplicas
├── target.spec.max_replicas → AppArgs.Template.MaxReplicas
├── target.spec.revision_suffix → AppArgs.Template.RevisionSuffix (optional)
├── target.spec.http_scale_rules → AppArgs.Template.HttpScaleRules (via buildHttpScaleRules)
├── target.spec.tcp_scale_rules → AppArgs.Template.TcpScaleRules (via buildTcpScaleRules)
├── target.spec.azure_queue_scale_rules → AppArgs.Template.AzureQueueScaleRules (via buildAzureQueueScaleRules)
├── target.spec.custom_scale_rules → AppArgs.Template.CustomScaleRules (via buildCustomScaleRules)
├── target.spec.secrets → AppArgs.Secrets (via buildSecrets)
├── target.spec.registries → AppArgs.Registries (via buildRegistries)
├── target.spec.ingress → AppArgs.Ingress (via buildIngress)
├── target.spec.dapr → AppArgs.Dapr (via buildDapr)
└── target.spec.identity → AppArgs.Identity (via buildIdentity)
```

## Output Wiring

```
App.ID()                 → container_app_id       (ARM resource ID)
App.LatestRevisionName   → latest_revision_name   (CD pipeline verification)
App.LatestRevisionFqdn   → latest_revision_fqdn   (direct revision access)
App.OutboundIpAddresses  → outbound_ip_addresses  (firewall allowlists)
App.Ingress.Fqdn         → ingress_fqdn           (conditional: only when ingress configured)
```

## Design Notes

- **No region field**: Container App location is inherited from its environment. This is like
  AzureSubnet inheriting region from its VNet. The module does not set `Location` on AppArgs.

- **Secrets handling**: The module supports two mutually exclusive secret types:
  - Plain-text: `value` is set directly on `AppSecretArgs`
  - Key Vault: `key_vault_secret_id` + `identity` are set (value is ignored)
  Key Vault references require the app to have a managed identity with Key Vault read access.

- **Scale rules nesting**: Scale rules live inside the `Template` block in the Pulumi/Terraform
  schema, alongside containers and volumes. The OpenMCF spec promotes them to top-level fields
  (`http_scale_rules`, `tcp_scale_rules`, etc.) for better discoverability. The module maps
  them back into `TemplateArgs` during resource creation.

- **Conditional ingress FQDN**: The `ingress_fqdn` output uses `ApplyT` to safely extract the
  FQDN from the ingress block. When ingress is not configured, `ingress_fqdn` is an empty
  string rather than an error.

- **Optional field handling**: Fields like `workload_profile_name`, `max_inactive_revisions`,
  `revision_suffix`, and the various probe/scale configurations are only set on their respective
  args when non-nil/non-empty. This allows Azure/Pulumi defaults to apply correctly.

- **ForceNew awareness**: `name`, `resource_group`, and `container_app_environment_id` are all
  ForceNew. Changes to these fields cause the Container App to be destroyed and recreated.
  Template changes (containers, scale rules, volumes) create new revisions without destroying
  the app.

- **Probe type separation**: Pulumi requires separate types for liveness, readiness, and startup
  probes (each has its own `XxxProbeArgs`, `XxxProbeHeaderArgs`). The OpenMCF spec uses a
  single shared `AzureContainerAppProbe` message. The module handles the type mapping in
  separate builder functions (`buildLivenessProbe`, `buildReadinessProbe`, `buildStartupProbe`).
