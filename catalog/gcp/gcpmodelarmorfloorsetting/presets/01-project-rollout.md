# Project Rollout

## Use Case

Put every Vertex AI model call in one project behind Model Armor without changing application code -- in report-only mode, so you can see what would be blocked before anything is.

## When to Use

- The first floor in a project that already runs Gemini workloads
- Measuring prompt-injection and malicious-link attempts across all apps in a project
- Setting a minimum for the templates teams build in the project

## What This Creates

- The project's floor: enforcement on, Vertex AI integrated in inspect-only mode with verdicts logged, and a minimum of prompt-injection detection at `MEDIUM_AND_ABOVE` plus malicious URL detection

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `aiPlatformFloorSetting.enforcementType` | `INSPECT_ONLY` | `INSPECT_AND_BLOCK` after reviewing the logged verdicts. |
| `integratedServices` | `AI_PLATFORM` | Add `GOOGLE_MCP_SERVER` for Google-hosted MCP traffic. |
| `filterConfig` | injection + URLs | Add content categories or sensitive-data detection to raise the minimum. |
