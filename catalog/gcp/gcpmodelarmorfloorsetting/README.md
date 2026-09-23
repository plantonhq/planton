# GCP Model Armor Floor Setting

The Model Armor floor setting of a project, folder, or organization -- the minimum safety screening everything beneath it must meet, and the switch that makes Google's own AI services enforce it. A floor sets the weakest filters any Model Armor template in its scope may carry, and, through integrated services, screens Vertex AI model calls and Google MCP server traffic directly, inspect-only or blocking, with no application change.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `modelarmor.googleapis.com` on the floor's project (project floors only; never disabled on destroy)
- **Floor setting** -- the scope's `model_armor_floorsetting`, applied over whatever floor it had

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Model Armor floor-setting admin permissions at the floor's scope (the project, the folder, or the organization).
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpProject`** / **`GcpFolder`** -- the scope, by reference (`scope.projectId`, `scope.folderId`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpModelArmorFloorSetting
metadata:
  name: project-floor
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `filterConfig` | `object` | The minimum filters: `piAndJailbreakFilterSettings`, `raiSettings.raiFilters[]`, `sdpSettings` (`basicConfig` or `advancedConfig`), `maliciousUriFilterSettings`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `object` | provider project | At most one of `projectId` (`GcpProject` ref), `folderId` (`GcpFolder` ref), `organizationId` (numeric). |
| `location` | `string` | `global` | Where Google manages the floor. |
| `enableFloorSettingEnforcement` | `bool` | `false` | Turn the floor on; while false nothing is checked or screened. |
| `integratedServices` | `string[]` | none | `AI_PLATFORM` (Vertex AI model calls), `GOOGLE_MCP_SERVER` (Google MCP servers). |
| `aiPlatformFloorSetting` | `object` | none | `enforcementType` (`INSPECT_ONLY` / `INSPECT_AND_BLOCK`, required) and `enableCloudLogging` for Vertex AI traffic. |
| `googleMcpServerFloorSetting` | `object` | none | The same for Google MCP server traffic. |
| `enableMultiLanguageDetection` | `bool` | `false` | Screen non-English prompts under the floor. |

### Validation Rules

- `scope` names at most one of project, folder, organization; `organizationId` is numeric without the `organizations/` prefix.
- `integratedServices` holds each service at most once.
- An integrated service setting needs `enforcementType`.
- The filter rules match `GcpModelArmorTemplate`'s.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `{parent}/locations/{location}/floorSetting` |
| `parent` | `string` | `projects/{id}`, `folders/{id}`, or `organizations/{id}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Destroy does not remove the floor.** Google cannot delete a floor setting: destroying this block only stops managing it, and the last applied floor stays in force. To relax a floor, apply it with `enableFloorSettingEnforcement: false` first, then destroy.
- **One floor per scope.** Applying overwrites whatever floor the project, folder, or organization had; two blocks for the same scope fight each other.
- **Roll out inspect-only.** `INSPECT_ONLY` on an integrated service records verdicts without blocking any model call -- see what would be blocked before switching to `INSPECT_AND_BLOCK`.
- **Templates below the floor are flagged or refused.** Raise templates first, then the floor.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpModelArmorTemplate** -- the templates the floor governs
- **GcpFolder** -- a folder floor every project beneath it inherits
- **GcpProject** -- a project floor

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
