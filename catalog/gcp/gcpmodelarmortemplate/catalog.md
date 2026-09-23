# GCP Model Armor Template

Puts a safety screen in front of your AI application. A Model Armor template is the set of checks every prompt and every model response goes through: attempts to hijack the model's instructions (prompt injection and jailbreaks), harmful content, leaked personal or confidential data, and links to malicious sites. Your team decides how strict each check is and whether a hit is only reported or blocked, and every application that calls Model Armor with the template gets the same protection.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `modelarmor.googleapis.com` on the project
- **Template** -- a `modelarmor.Template` with the filters, optional metadata, and labels

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Model Armor admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Model Armor Template**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Prompt Guard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpModelArmorTemplate
metadata:
  name: prompt-guard
  org: acme-corp
  env: prod
spec:
  location: us-central1
  filterConfig:
    piAndJailbreakFilterSettings:
      filterEnforcement: ENABLED
      confidenceLevel: MEDIUM_AND_ABOVE
    maliciousUriFilterSettings:
      filterEnforcement: ENABLED
  templateMetadata:
    enforcementType: INSPECT_ONLY
```

```shell
planton apply -f model-armor-template.yaml
```

This creates a template that reports prompt-injection attempts and malicious links without blocking yet. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the template's `name` output from a `GcpVertexAiSearchEngine` assistant's `modelArmorConfig`, or hand it to the applications that call Model Armor. Pair it with a `GcpModelArmorFloorSetting` to set the minimum for the whole project.

## Key Configuration

These are the most important decisions when configuring a template. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which checks run** -- each filter under `filterConfig` is off until you set it. Prompt injection and jailbreak detection is the one almost every AI app needs; add the content filters for user-facing chat, sensitive-data detection when prompts may carry personal data, and URL checks when responses include links.

**How strict each check is** -- `LOW_AND_ABOVE` catches the most and flags the most false positives; `HIGH` flags only clear cases. `MEDIUM_AND_ABOVE` is the balanced start.

**Report or block** -- `templateMetadata.enforcementType` decides whether a hit is only reported (`INSPECT_ONLY`) or blocked (`INSPECT_AND_BLOCK`). Start by reporting, then block once the thresholds are tuned.

**Where prompts are screened** -- the template's `location` is where your traffic is inspected. Pick it for your data-residency rules.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The template's full resource name | Vertex AI Search assistants, application sanitize calls |
| `template_id` | The template's id | SDK calls |
| `location` | Where the template screens traffic | The Model Armor endpoint to call |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Prompt guard** -- injection, jailbreak, and malicious-URL detection in report-only mode, the first template for any AI app. Start from the **Prompt Guard** preset.

**Strict customer-facing screen** -- every content filter, sensitive-data redaction through your own templates, and blocking with a custom message. Start from the **Strict Customer Chat** preset.

## Works With

- [**GCP Model Armor Floor Setting**](/cloud-catalog/gcp-model-armor-floor-setting) -- the minimum every template must meet, and enforcement on Vertex AI traffic
- [**GCP Vertex AI Search Engine**](/cloud-catalog/gcp-vertex-ai-search-engine) -- assistants screen prompts and responses through templates
- [**GCP Vertex AI Agent Engine**](/cloud-catalog/gcp-vertex-ai-agent-engine) -- agents that screen their input
