# Organization Baseline

## Use Case

An organization-wide AI safety baseline: every project's Vertex AI and Google MCP server traffic screened and blocked on a hit, and no template anywhere allowed below the minimum.

## When to Use

- A security team setting the floor for the whole estate after a project rollout proved the thresholds
- Regulated organizations that must show AI traffic is screened everywhere
- Folders and projects then tighten on top of this baseline

## What This Creates

- The organization's floor: enforcement on, Vertex AI and Google MCP servers integrated in blocking mode with verdicts logged, and a minimum of injection detection, three content categories, Google's basic sensitive-data detectors, malicious URL detection, and multi-language screening

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scope.organizationId` | placeholder | Your numeric organization id (or `folderId` for a folder baseline). |
| `*FloorSetting.enforcementType` | `INSPECT_AND_BLOCK` | Keep `INSPECT_ONLY` until a project rollout has proven the thresholds. |
| `filterConfig.raiSettings` | three categories | The categories every template must screen. |
