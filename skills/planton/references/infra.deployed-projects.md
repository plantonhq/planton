# Deployed Projects — Working Copies, Fixing Failures, Saving Changes

A chart is a blueprint; an **infra project** is a deployment of one. The
moment a project is created it carries its OWN copy of the templates and
values — detached from the chart it came from. Editing the chart never
changes a deployed project, and editing a project never changes the chart.
This file governs the second mode of your work: modifying something that is
already running.

## Am I working on a chart or on a deployed project?

Check for the binding file before assuming:

```
cat .planton/project.yaml
```

The workspace file tree does not show hidden paths, so LOOK for it with the
shell — never conclude from the tree alone. A folder carrying this file is a
**working copy of a deployed project**; the file names the project
(`projectId`, `projectName`), its organization, and its environment:

```yaml
projectId: infpj_01ABC…
projectName: eks-mumbai
org: acme
env: prod
```

No binding file → a plain chart folder; this reference does not apply.

What the binding changes:

- **Your edits target the running project.** Saving records a new project
  version and the platform starts a deployment pipeline for it immediately —
  there is no separate "deploy" step to offer. The organization's catalog
  policy never blocks these saves: it refuses only NEW creations of
  disabled kinds, and a deployed project's resources already exist — never
  warn about availability on a working-copy save
  (`catalog-availability.md`).
- **The folder is laid out fresh from the server** — by the app when the
  conversation starts, or by your own checkout — so what you read here is
  the deployed truth as of that moment, placed for THIS conversation alone.
- **The chart is out of the picture.** Never suggest publishing this folder
  as a chart or `chart publish` unless the user explicitly wants to turn
  their project's state into a reusable chart.

## When you are NOT in a working copy: check one out

A conversation about a deployed project does not always start inside its
working copy — in your own workspace, the request may simply name the
project. Never reconstruct a project's files from its record by hand; pull
the real thing:

```
planton infra project checkout <id-or-name> --output-dir <project-dir>
```

The checkout lays out the working copy — templates, values, and the hidden
binding — as its own top-level subfolder of your workspace, and everything
in this reference applies to it from that moment. It writes many files at
once, so the composing declaration (SKILL.md's live-screen rule) comes
FIRST. Re-running it against the same folder refreshes the managed files
from server truth and leaves any other files alone.

## Opening posture: diagnose FIRST when there is a failure

When the conversation opens from a failed pipeline (the opening message
names a pipeline id), **the four-step diagnosis is your FIRST ACT — run it
before any reply that merely acknowledges the request.** The user already
asked for the investigation by arriving here; a first turn that restates
the failure and waits is a failed turn. Consent gates CHANGES, never
diagnosis — reading status, logs, and records requires nobody's permission.

1. Run the four-step diagnosis from `planton-cli.md` (status → failed node →
   logs → stack job when needed), dumping large records to `.scratch/`
   files and reading those (the large-records rule in `planton-cli.md`).
2. Explain what went wrong in the user's language — name the resource and
   the cause, not the plumbing.
3. Recommend the fix with a reason, and say what it will do when applied
   (new version, new pipeline, roughly how long).
4. **Ask what they want** — even when the fix is obvious. Modifying a
   deployed project is always consent-gated.

The same compile loop applies while editing: `planton chart build . -o json`
works on a working copy exactly as on a chart folder (the binding file is
invisible to it). Keep the folder green before proposing to save.

## Saving: `chart install` IS the update

The Helm-consistent verb updates the existing project when the name matches
— it never creates a duplicate:

```
planton chart install <projectName> . --org <org> --env <env> \
  -m "one line saying WHY, like a commit message" --plain
```

Read every flag from the binding file — never guess the name, org, or env,
and never omit `--org`/`--env` (your CLI context may point elsewhere; the
project's environment is not yours to move). The `-m` message is mandatory
practice: it lands on the project's version history where the whole team
reads it.

What one invocation does, in order: builds the folder server-side (refusing
on errors), updates the project (new version), starts the deployment
pipeline, prints its id, and **follows the run to completion** — the exit
code is non-zero when the followed pipeline fails. That single command is
your save AND your monitor.

When holding one long-running command open is unsuitable, split it:

```
planton chart install <projectName> . --org <org> --env <env> -m "…" --no-follow
planton infra pipeline status <infpipe_id>     # poll between other work
planton infra pipeline logs <infpipe_id>       # when a node fails
```

## Narrate every transition — the user is watching the app

The app watches the project's pipelines: the moment your install creates
one, the screen announces it and navigates to the run. Say what you did the
moment you did it, so the voice and the screen agree:

- After the install starts: state that the project was updated and a new
  pipeline is running — and that they are being taken to it.
- On success: say it plainly, summarize what changed.
- On failure: diagnose again (same four steps), explain, recommend, ask.

## When the fix does not take

A correct fix can still fail with the SAME error as before. Verified failure
mode: a resource whose deploy failed can be left in a failed state that keeps
feeding the old configuration to new pipeline runs — the project's record and
even the resource's record show your corrected values, yet the run fails with
the stale ones. Recognize it by the signature (identical error after a save
that provably changed the relevant field), then clear it:

```
planton search cloud-resources --org <org> -e <env>      # find the resource id
planton get cloud-resource <cr_id> -o yaml               # confirm the record already carries your fix
planton infra cloud-resource cleanup <cr_id>             # clear the failed state (confirm with the user first)
```

Then save again (a fresh `chart install`, its own confirmation). Cleanup only
clears the failed deployment state — it does not touch cloud infrastructure —
but it is still a platform mutation: say what it does and get a yes.

A different non-taking failure has its own playbook: a rerun that fails
because a resource **already exists** (an earlier run created it in the
cloud, then died before the state recorded it). Do not delete-and-retry —
adopt the orphan into state with the platform's import commands. Read
`state-import.md`.

## Iterating until green

Failures can be layered. After the first fix attempt, ask ONE question that
sets the collaboration mode for the rest of the session:

> "Want me to keep iterating until this deploys green — asking you before
> each save — or check in with you after every diagnosis?"

Honor the answer. Even in keep-iterating mode, EVERY save still gets its own
explicit confirmation (what will change and why, one line each); what the
mode changes is whether you pause after each diagnosis or proceed to the
proposed fix directly.

## Boundaries

- Never `chart install` without the user's go-ahead for that specific change.
- Never change the target env/org to "make it work" — a failure that smells
  like wrong-environment is a finding to report, not to route around.
- Undeploying, purging, or deleting the project is out of scope here —
  surface the need, let the user drive it in the product.
