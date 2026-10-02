# GcpModelArmorTemplate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpModelArmorTemplateSpec defines a Model Armor template
(`google_model_armor_template`) -- a named set of safety filters an AI
application sends prompts and model responses through before trusting
them. A template screens for prompt injection and jailbreak attempts,
harmful content (the Responsible AI categories), sensitive data
(Sensitive Data Protection), and malicious URLs; it can only inspect and
report, or inspect and block.

Who uses a template: an application calling Model Armor's
sanitizeUserPrompt / sanitizeModelResponse methods names it; Vertex AI
Search engines name one for prompts and one for responses; a
GcpModelArmorFloorSetting sets the minimum every template in a project,
folder, or organization must meet.

Immutable: location and template_id. Everything else updates in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpModelArmorTemplate
metadata:
  name: prompt-guard
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  labels:
    team: platform
  filterConfig:
    piAndJailbreakFilterSettings:
      filterEnforcement: ENABLED
      confidenceLevel: MEDIUM_AND_ABOVE
    maliciousUriFilterSettings:
      filterEnforcement: ENABLED
    raiSettings:
      raiFilters:
        - filterType: DANGEROUS
          confidenceLevel: MEDIUM_AND_ABOVE
        - filterType: HARASSMENT
          confidenceLevel: MEDIUM_AND_ABOVE
  templateMetadata:
    # Report first, block once the thresholds are tuned.
    enforcementType: INSPECT_ONLY
    logSanitizeOperations: true
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.templateId` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.filterConfig` | `GcpModelArmorTemplateFilterConfig` | yes |  |  |
| `spec.filterConfig.maliciousUriFilterSettings` | `GcpModelArmorTemplateMaliciousUriFilterSettings` |  |  |  |
| `spec.filterConfig.maliciousUriFilterSettings.filterEnforcement` | `string` |  |  |  |
| `spec.filterConfig.piAndJailbreakFilterSettings` | `GcpModelArmorTemplatePiAndJailbreakFilterSettings` |  |  |  |
| `spec.filterConfig.piAndJailbreakFilterSettings.filterEnforcement` | `string` |  |  |  |
| `spec.filterConfig.piAndJailbreakFilterSettings.confidenceLevel` | `string` |  |  |  |
| `spec.filterConfig.raiSettings` | `GcpModelArmorTemplateRaiSettings` |  |  |  |
| `spec.filterConfig.raiSettings.raiFilters` | `[]GcpModelArmorTemplateRaiFilter` | yes |  |  |
| `spec.filterConfig.raiSettings.raiFilters[].filterType` | `string` | yes |  |  |
| `spec.filterConfig.raiSettings.raiFilters[].confidenceLevel` | `string` |  |  |  |
| `spec.filterConfig.sdpSettings` | `GcpModelArmorTemplateSdpSettings` |  |  |  |
| `spec.filterConfig.sdpSettings.basicConfig` | `GcpModelArmorTemplateSdpBasicConfig` |  |  |  |
| `spec.filterConfig.sdpSettings.basicConfig.filterEnforcement` | `string` |  |  |  |
| `spec.filterConfig.sdpSettings.advancedConfig` | `GcpModelArmorTemplateSdpAdvancedConfig` |  |  |  |
| `spec.filterConfig.sdpSettings.advancedConfig.inspectTemplate` | `string` |  |  |  |
| `spec.filterConfig.sdpSettings.advancedConfig.deidentifyTemplate` | `string` |  |  |  |
| `spec.templateMetadata` | `GcpModelArmorTemplateMetadata` |  |  |  |
| `spec.templateMetadata.logTemplateOperations` | `bool` |  |  |  |
| `spec.templateMetadata.logSanitizeOperations` | `bool` |  |  |  |
| `spec.templateMetadata.enableMultiLanguageDetection` | `bool` |  |  |  |
| `spec.templateMetadata.ignorePartialInvocationFailures` | `bool` |  |  |  |
| `spec.templateMetadata.customPromptSafetyErrorCode` | `int32` |  |  |  |
| `spec.templateMetadata.customPromptSafetyErrorMessage` | `string` |  |  |  |
| `spec.templateMetadata.customLlmResponseSafetyErrorCode` | `int32` |  |  |  |
| `spec.templateMetadata.customLlmResponseSafetyErrorMessage` | `string` |  |  |  |
| `spec.templateMetadata.enforcementType` | `string` |  |  |  |
| `spec.templateMetadata.filterVersionSelector` | `GcpModelArmorTemplateFilterVersionSelector` |  |  |  |
| `spec.templateMetadata.filterVersionSelector.alias` | `string` |  |  |  |
| `spec.templateMetadata.filterVersionSelector.version` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the template lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Where the template lives: a region (e.g. "us-central1") or a
multi-region ("us", "eu"). Prompts are screened in this location, so
pick the one your application's data-residency rules allow; the
application calls Model Armor's endpoint for this location. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+(-[a-z]+[0-9]+)?$"}}

### spec.templateId

`string`

The template's id -- the last segment of its resource name, what an
application passes to Model Armor. Letters, digits, hyphens, and
underscores, up to 63 characters. Defaults to metadata.name.
Immutable.

- rule: template_id must be 1-63 letters, digits, hyphens, or underscores

### spec.labels

`map<string, string>`

Labels on the template. The platform attribution labels are merged in
and win on key conflicts.

### spec.filterConfig

`GcpModelArmorTemplateFilterConfig` · required

Which filters run and how sensitive each one is. Required: a template
with no filter configuration screens nothing.

- rule: {"required":true}

### spec.filterConfig.maliciousUriFilterSettings

`GcpModelArmorTemplateMaliciousUriFilterSettings`

Malicious URL detection: flags links to known phishing and malware
sites in prompts and responses.

### spec.filterConfig.maliciousUriFilterSettings.filterEnforcement

`string`

"ENABLED" or "DISABLED". Unset leaves Google's default (disabled).

- rule: filter_enforcement must be ENABLED or DISABLED

### spec.filterConfig.piAndJailbreakFilterSettings

`GcpModelArmorTemplatePiAndJailbreakFilterSettings`

Prompt injection and jailbreak detection: flags attempts to override
the model's instructions or talk it out of its safety rules.

### spec.filterConfig.piAndJailbreakFilterSettings.filterEnforcement

`string`

"ENABLED" or "DISABLED". Unset leaves Google's default (disabled).

- rule: filter_enforcement must be ENABLED or DISABLED

### spec.filterConfig.piAndJailbreakFilterSettings.confidenceLevel

`string`

How confident Model Armor must be before it flags a prompt:
  "LOW_AND_ABOVE"    -- flags the most; the strictest setting, with
                        the most false positives
  "MEDIUM_AND_ABOVE" -- the balanced choice
  "HIGH"             -- flags only clear attempts

- rule: confidence_level must be LOW_AND_ABOVE, MEDIUM_AND_ABOVE, or HIGH

### spec.filterConfig.raiSettings

`GcpModelArmorTemplateRaiSettings`

Responsible AI content filters: sexually explicit, hate speech,
harassment, and dangerous content, each with its own threshold.

### spec.filterConfig.raiSettings.raiFilters

`[]GcpModelArmorTemplateRaiFilter` · required

One entry per content category to screen for; a category not listed
is not screened.

- rule: each filter_type may appear only once
- rule: {"repeated":{"minItems":"1"}}

### spec.filterConfig.raiSettings.raiFilters[].filterType

`string` · required

The content category: "SEXUALLY_EXPLICIT", "HATE_SPEECH",
"HARASSMENT", or "DANGEROUS".

- rule: {"required":true,"string":{"in":["SEXUALLY_EXPLICIT","HATE_SPEECH","HARASSMENT","DANGEROUS"]}}

### spec.filterConfig.raiSettings.raiFilters[].confidenceLevel

`string`

How confident Model Armor must be before it flags content in this
category: "LOW_AND_ABOVE" (strictest), "MEDIUM_AND_ABOVE", or "HIGH".

- rule: confidence_level must be LOW_AND_ABOVE, MEDIUM_AND_ABOVE, or HIGH

### spec.filterConfig.sdpSettings

`GcpModelArmorTemplateSdpSettings`

Sensitive Data Protection: finds (and optionally redacts) personal and
confidential data such as card numbers, credentials, and IDs.

- rule: set at most one of basic_config or advanced_config

### spec.filterConfig.sdpSettings.basicConfig

`GcpModelArmorTemplateSdpBasicConfig`

Google's predefined set of sensitive-data detectors (card numbers,
government IDs, credentials, and similar). The quick start.

### spec.filterConfig.sdpSettings.basicConfig.filterEnforcement

`string`

"ENABLED" or "DISABLED".

- rule: filter_enforcement must be ENABLED or DISABLED

### spec.filterConfig.sdpSettings.advancedConfig

`GcpModelArmorTemplateSdpAdvancedConfig`

Your own Sensitive Data Protection templates: an inspect template
decides what counts as sensitive, and an optional de-identify template
decides how findings are redacted in the sanitized text.

### spec.filterConfig.sdpSettings.advancedConfig.inspectTemplate

`string`

Inspect template:
projects/{project}/locations/{location}/inspectTemplates/{template}.
With only an inspect template, Model Armor reports findings.

- rule: inspect_template must be projects/{project}/locations/{location}/inspectTemplates/{template}

### spec.filterConfig.sdpSettings.advancedConfig.deidentifyTemplate

`string`

De-identify template:
projects/{project}/locations/{location}/deidentifyTemplates/{template}.
Adds redaction to the sanitized result. Every info type the
de-identify template names must also be in the inspect template.

- rule: deidentify_template must be projects/{project}/locations/{location}/deidentifyTemplates/{template}

### spec.templateMetadata

`GcpModelArmorTemplateMetadata`

How the template behaves around the filters: whether it blocks or only
reports, what an end user sees when a prompt or response is blocked,
logging, and which filter version it runs. Omit for Google's defaults.

### spec.templateMetadata.logTemplateOperations

`bool`

Log template create, update, and delete operations to Cloud Logging.

### spec.templateMetadata.logSanitizeOperations

`bool`

Log every sanitize call (the prompt or response screened and the
verdict) to Cloud Logging. Useful for tuning thresholds; the logs
carry the screened text.

### spec.templateMetadata.enableMultiLanguageDetection

`bool`

Detect and screen prompts in languages other than English. Sent only
when true.

### spec.templateMetadata.ignorePartialInvocationFailures

`bool`

When one of several detectors fails, return the other detectors'
verdicts instead of failing the whole call.

### spec.templateMetadata.customPromptSafetyErrorCode

`int32`

The error code a service extension returns to the end user when a
prompt trips a filter (for example 403). Sent only when set.

### spec.templateMetadata.customPromptSafetyErrorMessage

`string`

The error message returned to the end user when a prompt trips a
filter.

### spec.templateMetadata.customLlmResponseSafetyErrorCode

`int32`

The error code returned to the end user when a model response trips a
filter. Sent only when set.

### spec.templateMetadata.customLlmResponseSafetyErrorMessage

`string`

The error message returned to the end user when a model response
trips a filter.

### spec.templateMetadata.enforcementType

`string`

What happens when a filter trips:
  "INSPECT_ONLY"      -- the verdict is reported and the traffic
                         passes (a safe way to tune a new template)
  "INSPECT_AND_BLOCK" -- the traffic is blocked
Unset keeps Google's default.

- rule: enforcement_type must be INSPECT_ONLY or INSPECT_AND_BLOCK

### spec.templateMetadata.filterVersionSelector

`GcpModelArmorTemplateFilterVersionSelector`

Which version of Google's filters the template runs. Omit to follow
Google's default.

- rule: set exactly one of alias or version

### spec.templateMetadata.filterVersionSelector.alias

`string`

Follow a moving version: "FILTER_VERSION_ALIAS_STABLE" (Google's
recommended version) or "FILTER_VERSION_ALIAS_LATEST" (newest filters
first, verdicts may shift as Google updates them).

- rule: alias must be FILTER_VERSION_ALIAS_STABLE or FILTER_VERSION_ALIAS_LATEST

### spec.templateMetadata.filterVersionSelector.version

`string`

Pin an exact, immutable filter version such as "v1" or "v2"
(case-sensitive), so verdicts never shift under you.

- rule: version must look like v1, v2, ...

### spec.deletionPolicy

`string`

What happens to the template when this resource is destroyed:
  "" / "DELETE" -- the template is deleted; applications and engines
                   that still name it start failing their screens
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the template leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpModelArmorTemplate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/templates/{template_id} -- what sanitize calls and GcpVertexAiSearchEngine's model_armor_config take. |
| `status.outputs.template_id` | `string` | The template's id (the last segment of name). |
| `status.outputs.location` | `string` | The location the template lives in; the application calls Model Armor's endpoint for this location. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpVertexAiSearchEngine | `spec.assistants[].customerPolicy.modelArmorConfig.userPromptTemplate` | `status.outputs.name` |
| GcpVertexAiSearchEngine | `spec.assistants[].customerPolicy.modelArmorConfig.responseTemplate` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
