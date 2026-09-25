# Build Contract

The compile loop's machine interface. Agents and CI must treat this as a
published contract — exit codes and JSON shape are stable.

## Command

```
planton chart build <chart-dir> -o json
```

Read-only: nothing is written to disk unless you pass `--output-dir`. JSON mode
puts the **entire report on stdout** and keeps stderr for human messages only
(connection errors, deprecation warnings).

Build a variant without editing the chart with `--set name=value` (repeatable):

```
planton chart build . -o json --set dns_enabled=true --set aws_region=eu-west-1
```

The value is parsed as YAML, exactly as if values.yaml carried it: `true` is a
bool, `3` is a number, `"1.29"` stays a string, `[a, b]` is a list. The chart
on disk is never touched — this is the way to prove a bool toggle in both
positions (build once per position, compare the `resources[]` arrays).

Compile an environment's values file with `-f` (repeatable) before anyone
installs it:

```
planton chart build . -f values-prod.yaml -o json
```

A values file has the values.yaml shape and lists only the params it changes;
every other param keeps the chart's value. Resolution is by name, in this
order: values.yaml, then each `-f` file, then `--set` (later wins) -- the same
order `chart install` uses, which adds `--set-file name=path` before `--set`.
An unknown name in a file or a flag is exit 2, with the declared names listed.

`--show` prints the rendered documents (the params applied), the same
documents an install stores as the project's render; `--output-dir` writes both
`template.yaml` (the combined templates) and `rendered.yaml`. To see exactly
what an install would apply against a real environment without creating
anything, use `planton chart install <name> <chart> -f <file> --dry-run`: it
prints every param with its source and the render, and `-o json` adds `params`
and `renderedYaml` to this report.

Pin the control plane when needed (local instance, CI, desktop studio):

```
PLANTON_API_ENDPOINT_INFRA_HUB=127.0.0.1:23802 planton chart build . -o json
```

## Exit codes

| Code | Meaning | Agent action |
|------|---------|--------------|
| **0** | Chart valid. Warnings may be present; they never fail the build. | Phase 4 self-check, then stop or offer publish. |
| **1** | Chart has fixable errors. JSON report on stdout lists every issue. | Parse `issues[]`, fix templates/values, rebuild. |
| **2** | Check could not run. **Stdout is empty.** | Do not edit the chart. Fix environment (not a chart dir, bad flags, a values file or `--set` naming an unknown param, control plane down) or report and stop. |

Exit 2 is the critical guardrail: an unreachable control plane has never been
fixed by changing a template.

## JSON report shape

```json
{
  "result": "passed",
  "errors": 0,
  "warnings": 0,
  "issues": [],
  "resources": [
    {
      "kind": "AwsVpc",
      "name": "dev-vpc",
      "sourceFile": "network.yaml"
    }
  ]
}
```

| Field | Meaning |
|-------|---------|
| `result` | `"passed"` or `"failed"` |
| `errors` / `warnings` | Counts (only errors affect exit code) |
| `issues[]` | Every problem. Each issue has `severity`, `message`, and usually `file`, `resourceKind`, `resourceName`, `fieldPath` when attributable. |
| `resources[]` | **Rendered truth** after Jinjava — every resource the chart produces post-conditionals. Use this to verify edits landed. |
| `overrides[]` | The applied `--set` values (name + parsed value). **Present only when `--set` was used** — its absence means the report reflects values.yaml as written. |

Resource names render with the literal placeholder `<env>` where
`{{ values.env }}` appears (JSON-escaped as `\u003cenv\u003e`) — the real
environment is only bound at deployment. Seeing `<env>-vpc` in the report
means substitution worked, not that it failed.

## The wire channel (no CLI)

On the platform-tools arm the compile loop is the tool
`build_infra_chart_from_files` — the same compiler, the same report, over
the wire:

- **Input**: `files`, a map of root-relative path → content carrying exactly
  the chart folder's layout (`Chart.yaml`, `values.yaml`, everything under
  `templates/`), plus an optional `params` map that plays the role of
  `--set` (same one-override-per-param semantics, echoed back in the
  report's `overrides[]`).
- **Output**: the identical JSON report above. Read `result` where the CLI
  reads the exit code: `"failed"` with `issues[]` is exit 1 — fix the chart
  and resubmit.
- **The two error classes split exactly like exit 1 vs exit 2**: chart
  content problems (template errors, schema violations, malformed
  Chart.yaml/values.yaml) come back as a NORMAL tool result carrying a
  failed report; invocation problems (the files are not a chart's shape, an
  unknown `params` name, an oversized payload, an unreachable backend) come
  back as a TOOL ERROR with no report. A tool error means fix the call or
  the environment — never edit the chart in response, exactly the exit-2
  rule.
- **Assemble `files` from a fresh directory listing on every call.** The
  compiler sees only what you submit: a template you composed but left out
  of the map produces a green report for a smaller chart than exists on
  disk — a silent failure no exit code will ever flag. List the chart
  directory, include every file, then call.

Local parse failures (malformed `Chart.yaml` or `values.yaml`) and
server-side rejection of the chart manifest itself (e.g. a missing
`spec.selector`) all arrive through the same `issues[]` channel with
exit 1 — one channel for everything fixable.

## Human mode (no `-o json`)

Exit codes are identical. The formatted report goes to **stderr** (with a
parameter summary when `-f` or `--set` changed anything); stdout stays empty
unless `--show` prints the rendered documents. Agents should always use `-o json`.

## Flag interactions

- `--show` + `-o json` → exit 2 (conflicting outputs).
- `--set unknown_name=x`, or a `-f` file naming an unknown param → exit 2, stderr lists the params values.yaml declares. Fix the flag or the file, never the chart.
- `--no-browser`, `--copy` → deprecated no-ops; fleet Makefiles may still pass them.
- `--output-dir <dir>` → writes `template.yaml` (the combined templates) and `rendered.yaml` (the render); optional for agents.

## Offline gates (no control plane)

| Command | Purpose |
|---------|---------|
| `planton explain <Kind>` | Field names, types, docs, outputs (drill with dotted paths) |
| `planton validate manifest.yaml` | Validate one draft manifest offline |

These reduce errors before the compile loop but do **not** replace
`chart build` — only the control plane runs full template rendering,
valueFrom validation, and cross-resource checks.

## CI usage

```bash
planton chart build ./charts/my-chart -o json > report.json
code=$?
if [ "$code" -eq 1 ]; then
  jq '.issues[] | select(.severity=="error")' report.json
  exit 1
fi
if [ "$code" -eq 2 ]; then
  echo "build could not run" >&2
  exit 2
fi
```

Treat exit 2 as infrastructure failure, not chart failure.
