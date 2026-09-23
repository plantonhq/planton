# GCP Model Armor Template

A Model Armor template -- the named set of safety filters an AI application screens prompts and model responses through before trusting them. One template decides which filters run and how sensitive each is: prompt injection and jailbreak detection, Responsible AI content filters (sexually explicit, hate speech, harassment, dangerous), Sensitive Data Protection (Google's basic detectors or your own inspect and de-identify templates), and malicious URL detection. It can only report what it finds or report and block. Applications name the template in Model Armor's sanitize calls; Vertex AI Search assistants name one for prompts and one for responses.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `modelarmor.googleapis.com` on the project (never disabled on destroy)
- **Template** -- a `model_armor_template` with the filter configuration, optional metadata, and labels

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Model Armor admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **Sensitive Data Protection templates** -- for `filterConfig.sdpSettings.advancedConfig`, an inspect template (and optionally a de-identify template) you manage in Sensitive Data Protection, named by full resource name.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpModelArmorTemplate
metadata:
  name: prompt-guard
spec:
  location: us-central1
  filterConfig:
    piAndJailbreakFilterSettings:
      filterEnforcement: ENABLED
      confidenceLevel: MEDIUM_AND_ABOVE
    maliciousUriFilterSettings:
      filterEnforcement: ENABLED
```

```shell
planton apply -f model-armor-template.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | A region (`us-central1`) or multi-region (`us`, `eu`) where prompts are screened. Immutable. |
| `filterConfig` | `object` | Which filters run: `piAndJailbreakFilterSettings`, `raiSettings.raiFilters[]`, `sdpSettings` (`basicConfig` or `advancedConfig`), `maliciousUriFilterSettings`. A filter is off unless its block is set. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `templateId` | `string` | `metadata.name` | Letters, digits, hyphens, underscores; up to 63. Immutable. |
| `labels` | `map<string,string>` | none | Merged under the platform attribution labels. |
| `templateMetadata` | `object` | Google defaults | `enforcementType` (`INSPECT_ONLY` / `INSPECT_AND_BLOCK`), sanitize and template logging, `enableMultiLanguageDetection`, `ignorePartialInvocationFailures`, custom block codes and messages, `filterVersionSelector` (`alias` or `version`). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Confidence levels are `LOW_AND_ABOVE`, `MEDIUM_AND_ABOVE`, or `HIGH`; enforcement switches are `ENABLED` or `DISABLED`.
- Each Responsible AI category appears at most once.
- `sdpSettings` takes `basicConfig` or `advancedConfig`, never both; advanced templates are full `projects/.../inspectTemplates/...` and `.../deidentifyTemplates/...` names.
- `filterVersionSelector` takes exactly one of `alias` and `version` (`v1`, `v2`, ...).

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/templates/{template_id}` -- what sanitize calls and search assistants take |
| `template_id` | `string` | The template's id |
| `location` | `string` | The location; applications call Model Armor's endpoint for it |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Start in inspect-only.** `templateMetadata.enforcementType: INSPECT_ONLY` reports verdicts without blocking, so you can tune thresholds on real traffic before switching to `INSPECT_AND_BLOCK`.
- **A template protects only the traffic that is sent through it.** Your application (or a Vertex AI Search assistant) must call Model Armor with the template. To screen every Vertex AI model call in a project without code changes, pair it with a `GcpModelArmorFloorSetting`.
- **A floor setting can reject a weak template.** Once a floor is enforced in the project, folder, or organization, a template below it is flagged or refused.
- **Sanitize logging records the screened text.** Turn `logSanitizeOperations` on deliberately; the logs carry prompts and responses.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpModelArmorFloorSetting** -- the minimum every template in a scope must meet, and enforcement on Vertex AI traffic
- **GcpVertexAiSearchEngine** -- assistants screen prompts and responses through templates
- **GcpVertexAiAgentEngine** -- agents that call Model Armor before trusting input

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
