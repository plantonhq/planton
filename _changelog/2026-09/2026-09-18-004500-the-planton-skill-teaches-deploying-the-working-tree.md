# The planton skill teaches deploying the working tree

**Date**: September 18, 2026
**Type**: Documentation
**Components**: the `planton` skill's `service.delivery-verbs.md` and `service.kustomize-authoring.md` references; the `ci-cd` docs pages `deployment-stage.md` and `what-is-a-service.md`

## Summary

The connected `planton service deploy` verb gained `--from-tree`: the environment's `_kustomize` overlay renders on the person's machine and deploys through the control plane, which writes it onto the service's configuration for that one environment named as theirs (who, which commit, whether the tree carried uncommitted changes) and runs the same gates, rollout verification, URLs, and receipt as any other deploy; the next push with the same content hands the entry back to git. The skill now teaches the door, when to recommend it, what it accepts and refuses, and how to read the CLI's confirmation line back to a person. Two public docs pages that still showed retired CLI spellings (`kustomize build`, `dot-env`, `deploy --project .`) now show the live verbs, the working-tree door among them.

## What Changed

- **`service.delivery-verbs.md`**: a "Deploy from the working tree" paragraph after the deploy verb -- for a git-maintained service never pushed through the build lane, or an overlay someone wants to see run before committing; git-maintained services only; a person at a terminal, never a CI credential; the run's name and the receipt's provenance; git as the stronger writer; the confirmation line and what "uncommitted changes" means.
- **`service.kustomize-authoring.md`**: the tree named as the one door beside a commit while a service is git-maintained.
- **`deployment-stage.md`**: the CLI reference block rewritten with the live verbs (`kustomize init --env`, `kustomize eject`, `kustomize checkout`, `service env run`, `service deploy --from-tree`) and a paragraph on the working-tree door.
- **`what-is-a-service.md`**: the deploy example shows the connected verb with and without `--from-tree`.

## Why It Matters

A kustomize service that has never been pushed has an empty record, and until now an assistant could only tell the person to wire a git connection first. Now it can name the real door -- the tree they are holding -- and say honestly what the record will say afterwards. The Planton Assistant reads the new words on the next skill publish; coding agents installing the skill read them at the next release.
