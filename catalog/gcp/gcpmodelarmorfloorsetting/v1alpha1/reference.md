# GcpModelArmorFloorSetting

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpModelArmorFloorSettingSpec defines the Model Armor floor setting of a
project, folder, or organization (`google_model_armor_floorsetting`) --
the minimum safety screening everything beneath it must meet, and the
switch that makes Google's own AI services enforce it.

A floor does two things, both only while enable_floor_setting_enforcement
is true. It sets the weakest filters any Model Armor template in its
scope may carry. And through integrated_services it screens traffic to
Vertex AI (Gemini generateContent calls) and Google's MCP servers
directly, with no application change -- the way to put every model call
in a project behind Model Armor at once. The same thresholds that bound
templates are the ones the integrated services' traffic is screened at.

Google keeps exactly one floor setting per project, folder, and
organization. Applying this block overwrites that setting. Destroying it
does NOT remove it: Google cannot delete a floor setting, so destroy only
stops managing it and the last applied floor stays in force. To relax a
floor, apply it with enable_floor_setting_enforcement false first.

A changed scope or location names a different floor, so both engines
plan a replacement: the old floor is only dropped from management and
stays in force in Google Cloud, and the new scope's floor is written.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpModelArmorFloorSetting
metadata:
  name: project-floor
spec:
  scope:
    projectId:
      value: my-gcp-project
  enableFloorSettingEnforcement: true
  integratedServices:
    - AI_PLATFORM
  filterConfig:
    piAndJailbreakFilterSettings:
      filterEnforcement: ENABLED
      confidenceLevel: HIGH
  aiPlatformFloorSetting:
    # Record verdicts on every Vertex AI model call without blocking any.
    enforcementType: INSPECT_ONLY
    enableCloudLogging: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpModelArmorFloorSettingScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.scope.organizationId` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.enableFloorSettingEnforcement` | `bool` |  |  |  |
| `spec.integratedServices` | `[]string` |  |  |  |
| `spec.filterConfig` | `GcpModelArmorFloorSettingFilterConfig` | yes |  |  |
| `spec.filterConfig.maliciousUriFilterSettings` | `GcpModelArmorFloorSettingMaliciousUriFilterSettings` |  |  |  |
| `spec.filterConfig.maliciousUriFilterSettings.filterEnforcement` | `string` |  |  |  |
| `spec.filterConfig.piAndJailbreakFilterSettings` | `GcpModelArmorFloorSettingPiAndJailbreakFilterSettings` |  |  |  |
| `spec.filterConfig.piAndJailbreakFilterSettings.filterEnforcement` | `string` |  |  |  |
| `spec.filterConfig.piAndJailbreakFilterSettings.confidenceLevel` | `string` |  |  |  |
| `spec.filterConfig.raiSettings` | `GcpModelArmorFloorSettingRaiSettings` |  |  |  |
| `spec.filterConfig.raiSettings.raiFilters` | `[]GcpModelArmorFloorSettingRaiFilter` | yes |  |  |
| `spec.filterConfig.raiSettings.raiFilters[].filterType` | `string` | yes |  |  |
| `spec.filterConfig.raiSettings.raiFilters[].confidenceLevel` | `string` |  |  |  |
| `spec.filterConfig.sdpSettings` | `GcpModelArmorFloorSettingSdpSettings` |  |  |  |
| `spec.filterConfig.sdpSettings.basicConfig` | `GcpModelArmorFloorSettingSdpBasicConfig` |  |  |  |
| `spec.filterConfig.sdpSettings.basicConfig.filterEnforcement` | `string` |  |  |  |
| `spec.filterConfig.sdpSettings.advancedConfig` | `GcpModelArmorFloorSettingSdpAdvancedConfig` |  |  |  |
| `spec.filterConfig.sdpSettings.advancedConfig.inspectTemplate` | `string` |  |  |  |
| `spec.filterConfig.sdpSettings.advancedConfig.deidentifyTemplate` | `string` |  |  |  |
| `spec.aiPlatformFloorSetting` | `GcpModelArmorFloorSettingServiceSetting` |  |  |  |
| `spec.aiPlatformFloorSetting.enforcementType` | `string` | yes |  |  |
| `spec.aiPlatformFloorSetting.enableCloudLogging` | `bool` |  |  |  |
| `spec.googleMcpServerFloorSetting` | `GcpModelArmorFloorSettingServiceSetting` |  |  |  |
| `spec.googleMcpServerFloorSetting.enforcementType` | `string` | yes |  |  |
| `spec.googleMcpServerFloorSetting.enableCloudLogging` | `bool` |  |  |  |
| `spec.enableMultiLanguageDetection` | `bool` |  |  |  |

## Field Details

### spec.scope

`GcpModelArmorFloorSettingScope`

Whose floor this is. Omit for the provider's default project: an empty
scope and a project_id naming that same project are the same floor.
A deploy whose provider has no default project is refused by both
engines before anything is sent. Changing the scope replaces the
floor (see above).

- rule: folder_id must be the numeric folder ID, without the folders/ prefix
- rule: set at most one of project_id, folder_id, or organization_id (empty means the provider's default project)

### spec.scope.projectId

`string | valueFrom`

Project floor: a literal project ID or a GcpProject reference.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scope.folderId

`string | valueFrom`

Folder floor: the folder's numeric ID, without the folders/ prefix --
a literal or a GcpFolder reference. Every project and folder beneath
it inherits the floor.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.scope.organizationId

`string`

Organization floor: the numeric organization ID, without the
organizations/ prefix. The floor for the whole estate.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.location

`string`

The location of the floor setting. Google manages floor settings at
"global", the default; set a region only if Google documents a
regional floor for your case. Changing it replaces the floor (see
above).

- rule: location must be global, a region, or a multi-region

### spec.enableFloorSettingEnforcement

`bool`

Turn the floor on. While false the floor is recorded but nothing is
checked or screened -- the state to apply before walking away from a
floor, since destroy leaves the last applied floor in force.

### spec.integratedServices

`[]string`

The Google services whose traffic the floor screens directly:
  "AI_PLATFORM"       -- Vertex AI model calls (configure the behavior
                         in ai_platform_floor_setting, required with it)
  "GOOGLE_MCP_SERVER" -- Google-hosted MCP servers (configure it in
                         google_mcp_server_floor_setting, required with
                         it)
Each service and its setting come together: a service without its
setting, or a setting without its service, is refused. Empty: the
floor only governs templates.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["AI_PLATFORM","GOOGLE_MCP_SERVER"]}}}}

### spec.filterConfig

`GcpModelArmorFloorSettingFilterConfig` · required

The minimum filters. Required, but an empty block is accepted and sets
no minimum: a filter constrains templates and traffic only once its
own block is set.

- rule: {"required":true}

### spec.filterConfig.maliciousUriFilterSettings

`GcpModelArmorFloorSettingMaliciousUriFilterSettings`

Malicious URL detection.

### spec.filterConfig.maliciousUriFilterSettings.filterEnforcement

`string`

"ENABLED" or "DISABLED". Unset is Google's default, which is disabled.

- rule: filter_enforcement must be ENABLED or DISABLED

### spec.filterConfig.piAndJailbreakFilterSettings

`GcpModelArmorFloorSettingPiAndJailbreakFilterSettings`

Prompt injection and jailbreak detection.

### spec.filterConfig.piAndJailbreakFilterSettings.filterEnforcement

`string`

"ENABLED" or "DISABLED". Unset is Google's default, which is disabled.

- rule: filter_enforcement must be ENABLED or DISABLED

### spec.filterConfig.piAndJailbreakFilterSettings.confidenceLevel

`string`

The weakest threshold a template may use, and the threshold integrated
services' traffic is screened at: "LOW_AND_ABOVE" (strictest),
"MEDIUM_AND_ABOVE", or "HIGH". Unset is Google's default, which is
LOW_AND_ABOVE.

- rule: confidence_level must be LOW_AND_ABOVE, MEDIUM_AND_ABOVE, or HIGH

### spec.filterConfig.raiSettings

`GcpModelArmorFloorSettingRaiSettings`

Responsible AI content filters.

### spec.filterConfig.raiSettings.raiFilters

`[]GcpModelArmorFloorSettingRaiFilter` · required

One entry per required content category.

- rule: each filter_type may appear only once
- rule: {"repeated":{"minItems":"1"}}

### spec.filterConfig.raiSettings.raiFilters[].filterType

`string` · required

"SEXUALLY_EXPLICIT", "HATE_SPEECH", "HARASSMENT", or "DANGEROUS".

- rule: {"required":true,"string":{"in":["SEXUALLY_EXPLICIT","HATE_SPEECH","HARASSMENT","DANGEROUS"]}}

### spec.filterConfig.raiSettings.raiFilters[].confidenceLevel

`string`

The weakest threshold a template may use for this category, and the
threshold integrated services' traffic is screened at:
"LOW_AND_ABOVE" (strictest), "MEDIUM_AND_ABOVE", or "HIGH". Unset is
Google's default, which is LOW_AND_ABOVE.

- rule: confidence_level must be LOW_AND_ABOVE, MEDIUM_AND_ABOVE, or HIGH

### spec.filterConfig.sdpSettings

`GcpModelArmorFloorSettingSdpSettings`

Sensitive Data Protection.

- rule: set at most one of basic_config or advanced_config

### spec.filterConfig.sdpSettings.basicConfig

`GcpModelArmorFloorSettingSdpBasicConfig`

Google's predefined sensitive-data detectors.

### spec.filterConfig.sdpSettings.basicConfig.filterEnforcement

`string`

"ENABLED" or "DISABLED". Unset is Google's default, which is disabled.

- rule: filter_enforcement must be ENABLED or DISABLED

### spec.filterConfig.sdpSettings.advancedConfig

`GcpModelArmorFloorSettingSdpAdvancedConfig`

Your own Sensitive Data Protection templates.

- rule: deidentify_template needs inspect_template: the inspect template decides what the de-identify template redacts

### spec.filterConfig.sdpSettings.advancedConfig.inspectTemplate

`string`

Inspect template:
projects/{project}/locations/{location}/inspectTemplates/{template}.

- rule: inspect_template must be projects/{project}/locations/{location}/inspectTemplates/{template}

### spec.filterConfig.sdpSettings.advancedConfig.deidentifyTemplate

`string`

De-identify template:
projects/{project}/locations/{location}/deidentifyTemplates/{template}.
Every info type it names must also be in the inspect template.

- rule: deidentify_template must be projects/{project}/locations/{location}/deidentifyTemplates/{template}

### spec.aiPlatformFloorSetting

`GcpModelArmorFloorSettingServiceSetting`

How the floor treats Vertex AI traffic. Set exactly when AI_PLATFORM
is in integrated_services.

### spec.aiPlatformFloorSetting.enforcementType

`string` · required

What happens when the floor's filters trip on the service's traffic:
  "INSPECT_ONLY"      -- the verdict is recorded and the call proceeds
                         (the safe way to roll a floor out)
  "INSPECT_AND_BLOCK" -- the call is blocked
Required: Google needs exactly one of the two.

- rule: {"required":true,"string":{"in":["INSPECT_ONLY","INSPECT_AND_BLOCK"]}}

### spec.aiPlatformFloorSetting.enableCloudLogging

`bool`

Write the floor's verdicts for this service to Cloud Logging.

### spec.googleMcpServerFloorSetting

`GcpModelArmorFloorSettingServiceSetting`

How the floor treats Google MCP server traffic. Set exactly when
GOOGLE_MCP_SERVER is in integrated_services.

### spec.googleMcpServerFloorSetting.enforcementType

`string` · required

What happens when the floor's filters trip on the service's traffic:
  "INSPECT_ONLY"      -- the verdict is recorded and the call proceeds
                         (the safe way to roll a floor out)
  "INSPECT_AND_BLOCK" -- the call is blocked
Required: Google needs exactly one of the two.

- rule: {"required":true,"string":{"in":["INSPECT_ONLY","INSPECT_AND_BLOCK"]}}

### spec.googleMcpServerFloorSetting.enableCloudLogging

`bool`

Write the floor's verdicts for this service to Cloud Logging.

### spec.enableMultiLanguageDetection

`bool`

Screen prompts in languages other than English under the floor. Sent
only when true.

## Validation Rules

- `ai_platform_floor_setting.needs_service`: ai_platform_floor_setting applies only to integrated Vertex AI traffic: add AI_PLATFORM to integrated_services, or remove the setting
- `integrated_services.ai_platform_needs_setting`: AI_PLATFORM is integrated but ai_platform_floor_setting is not set: set it to choose INSPECT_ONLY or INSPECT_AND_BLOCK for Vertex AI traffic
- `google_mcp_server_floor_setting.needs_service`: google_mcp_server_floor_setting applies only to integrated MCP server traffic: add GOOGLE_MCP_SERVER to integrated_services, or remove the setting
- `integrated_services.google_mcp_server_needs_setting`: GOOGLE_MCP_SERVER is integrated but google_mcp_server_floor_setting is not set: set it to choose INSPECT_ONLY or INSPECT_AND_BLOCK for MCP server traffic

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpModelArmorFloorSetting, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: {parent}/locations/{location}/floorSetting. |
| `status.outputs.parent` | `string` | The parent the floor governs: projects/{id}, folders/{id}, or organizations/{id}. |
| `status.outputs.location` | `string` | The location the floor is managed at ("global" unless the spec names another). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |

## See Also

- [Overview](../README.md)
