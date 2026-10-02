# GcpModelArmorTemplate Guide

The judgment this guide protects: a template only protects the traffic that is sent through it, and a template that blocks before it has been tuned breaks your product. Report first, tune the thresholds on real traffic, then block -- and put a floor under every template so nothing in the project falls below the minimum.

## What a template is, and what it is not

A template is a named list of checks. It does nothing on its own: an application calls Model Armor's `sanitizeUserPrompt` and `sanitizeModelResponse` with the template's name, or a Vertex AI Search assistant names one template for prompts and one for responses. To screen Vertex AI model calls without touching application code, a `GcpModelArmorFloorSetting` with `AI_PLATFORM` integrated does it at the platform level.

## Choosing filters

- **Prompt injection and jailbreak** (`piAndJailbreakFilterSettings`) -- attempts to override the model's instructions or talk it out of its rules. Almost every AI app needs it.
- **Responsible AI content** (`raiSettings.raiFilters[]`) -- one entry per category (`SEXUALLY_EXPLICIT`, `HATE_SPEECH`, `HARASSMENT`, `DANGEROUS`), each with its own threshold. A category not listed is not screened.
- **Sensitive Data Protection** (`sdpSettings`) -- `basicConfig` turns on Google's predefined detectors; `advancedConfig` names your own inspect template (what counts as sensitive) and optional de-identify template (how it is redacted in the sanitized text). Every info type the de-identify template names must be in the inspect template.
- **Malicious URLs** (`maliciousUriFilterSettings`) -- links to known phishing and malware sites.

## Thresholds

`LOW_AND_ABOVE` flags the most (strictest, most false positives), `MEDIUM_AND_ABOVE` is the balanced start, `HIGH` flags only clear cases. Tune per filter: a customer-support bot may want strict harassment detection and a looser dangerous-content threshold for, say, a chemistry tutor.

## Report, then block

`templateMetadata.enforcementType: INSPECT_ONLY` returns the verdicts without blocking; with `logSanitizeOperations` on, every verdict lands in Cloud Logging so you can see what would have been blocked. Switch to `INSPECT_AND_BLOCK` once the false-positive rate is acceptable. `customPromptSafetyErrorMessage` and `customLlmResponseSafetyErrorMessage` set what the end user sees when something is blocked. The sanitize logs carry the screened text -- treat them like the conversations they record.

## Filter versions

Google improves its filters over time. `filterVersionSelector.alias: FILTER_VERSION_ALIAS_STABLE` follows Google's recommended version; `FILTER_VERSION_ALIAS_LATEST` gets new filters first; `version: v2` pins an exact version so verdicts never shift under you. Pin when you need reproducible verdicts (audits, regulated flows).

## Floors and templates

A `GcpModelArmorFloorSetting` sets the weakest filters any template in its project, folder, or organization may carry. Once the floor is enforced, a template below it is flagged or refused -- so change the floor and its templates together, and keep templates at least as strict as the floor.

## Identity, location, and destroy

`templateId` defaults to `metadata.name` and is fixed at creation, as is `location` (a region or `us` / `eu`). The `name` output is the full path applications and assistants take. Under `DELETE`, destroying the template makes every caller that still names it fail its screen; `PREVENT` makes destroy fail.
