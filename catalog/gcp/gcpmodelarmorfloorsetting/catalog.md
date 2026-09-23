# GCP Model Armor Floor Setting

Sets the minimum AI safety screening for a whole project, folder, or organization -- and turns it on for Google's own AI services. With a floor in place, every Gemini call made through Vertex AI in that scope (and Google-hosted MCP server traffic, if you choose) is checked for prompt injection, harmful content, sensitive data, and malicious links without changing a line of application code, and no team can publish a Model Armor template weaker than the floor.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `modelarmor.googleapis.com` on the floor's project (project floors)
- **Floor setting** -- the scope's `modelarmor.Floorsetting`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Model Armor floor-setting admin permissions at the floor's scope. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Model Armor Floor Setting**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Project Rollout** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpModelArmorFloorSetting
metadata:
  name: ml-project-floor
  org: acme-corp
  env: prod
spec:
  enableFloorSettingEnforcement: true
  integratedServices:
    - AI_PLATFORM
  filterConfig:
    piAndJailbreakFilterSettings:
      filterEnforcement: ENABLED
      confidenceLevel: MEDIUM_AND_ABOVE
  aiPlatformFloorSetting:
    enforcementType: INSPECT_ONLY
    enableCloudLogging: true
```

```shell
planton apply -f model-armor-floor-setting.yaml
```

This puts every Vertex AI model call in the project behind prompt-injection screening in report-only mode. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpFolder` from `scope.folderId` to set one floor for every project in the folder; pair the floor with the `GcpModelArmorTemplate` resources teams build on top of it.

## Key Configuration

These are the most important decisions when configuring a floor. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Whose floor** -- a project, a folder (inherited by everything beneath it), or the organization. Empty means the connection's project.

**Which services it screens** -- `integratedServices` puts Vertex AI model calls (`AI_PLATFORM`) and Google MCP server traffic (`GOOGLE_MCP_SERVER`) behind the floor directly. Leave it empty to only govern templates.

**Report or block** -- each integrated service's `enforcementType` decides whether a hit is recorded (`INSPECT_ONLY`) or blocked (`INSPECT_AND_BLOCK`). Roll out by reporting first.

**The minimum** -- `filterConfig` is the weakest filter set any template in scope may carry.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `scope.projectId` | `status.outputs.project_id` |
| **GcpFolder** | `scope.folderId` | `status.outputs.folder_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The floor setting's full resource name | Audits and console links |
| `parent` | The scope the floor governs | Reporting |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Project rollout** -- a project floor screening Vertex AI traffic inspect-only with logging, the safe first step. Start from the **Project Rollout** preset.

**Organization baseline** -- an organization-wide floor that blocks on Vertex AI and Google MCP servers. Start from the **Organization Baseline** preset.

## Works With

- [**GCP Model Armor Template**](/cloud-catalog/gcp-model-armor-template) -- the templates the floor governs
- [**GCP Folder**](/cloud-catalog/gcp-folder) -- a floor inherited by every project in the folder
- [**GCP Project**](/cloud-catalog/gcp-project) -- a project floor
