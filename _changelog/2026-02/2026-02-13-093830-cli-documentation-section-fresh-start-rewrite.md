# CLI Documentation Section — Fresh-Start Rewrite

**Date**: February 13, 2026
**Type**: Enhancement
**Components**: Documentation, CLI Reference, User Experience

## Summary

Complete fresh-start rewrite of the CLI documentation section: 5 existing pages rewritten from scratch and 3 new pages created, totaling 1,300 lines across 8 pages. Established a deduplication architecture where `cli-reference.md` serves as the single source of truth for all flags, eliminating systemic content overlap that existed across the previous 5 pages.

## Problem Statement / Motivation

The existing 5 CLI documentation pages (2,958 lines) had structural quality problems that could not be fixed with targeted edits.

### Pain Points

- **Pervasive duplication**: Common flags, credential handling, CI/CD patterns, and provisioner selection repeated across 4 of 5 pages. Updating a flag description required editing multiple files.
- **Factual errors**: `-f, -f` typo appeared in flag tables across multiple pages (should be `-f, --manifest`). Duplicate usage lines in `unified-commands.md`.
- **Stale content**: "NEW!" and "New in this release" labels throughout — meaningless to new visitors.
- **Emoji in prose**: Trailing rocket emoji, bullet-style emoji in examples — violated style conventions.
- **Missing coverage**: Terraform commands fully implemented in source (5 subcommands in `cmd/openmcf/root/terraform/`) but had no documentation page.
- **No structural integrity**: No clear "single source of truth" for any piece of information.

## Solution / What's New

Designed and executed a fresh-start rewrite with a deduplication architecture that assigns clear content ownership to each page.

### Page Architecture

```
cli/
  index.md                  (104 lines)  Gateway: installation, CLI overview, navigation
  cli-reference.md          (234 lines)  Master reference: ALL flags, command tree, exit codes
  unified-commands.md       (170 lines)  Provisioner routing concept and 5 unified commands
  pulumi-commands.md        (173 lines)  7 Pulumi subcommands and workflows
  tofu-commands.md          (171 lines)  7 OpenTofu subcommands and workflows
  terraform-commands.md     (157 lines)  5 Terraform subcommands (NEW)
  module-management.md      (149 lines)  Module lifecycle: checkout, pull, upgrade, etc. (NEW)
  configuration.md          (142 lines)  Config, validate, load-manifest, version (NEW)
```

### Deduplication Principle

Every fact appears in exactly one authoritative location:

- **cli-reference.md** owns: complete command tree, all flag reference tables organized by group, exit codes, file system paths
- **unified-commands.md** owns: provisioner routing concept, the `openmcf.org/provisioner` label, unified vs. direct comparison
- **Engine pages** own: engine-specific subcommands, engine-specific flags, engine-specific workflows
- **module-management.md** owns: module resolution chains, staging area, version pinning
- **configuration.md** owns: config, validate, load-manifest, manifest source resolution priority

Other pages link to the authoritative location rather than duplicating content.

## Implementation Details

### Source Code Verification

Every claim was cross-referenced against source code:

- **Command registration**: `cmd/openmcf/root.go` — verified the complete command tree
- **Flag definitions**: `internal/cli/iacflags/*.go` — verified every flag name, shorthand, and default value
- **Flag constants**: `internal/cli/flag/flag.go` — verified constant names match registration
- **Manifest resolution**: `internal/cli/manifest/resolver.go` — verified priority order
- **Module resolution**: `pkg/iac/tofu/tofumodule/module_directory.go` and `pkg/iac/pulumi/pulumimodule/module_directory.go`

### Key Discovery: `-f` Shorthand Availability

During source verification, discovered that the `-f` shorthand for `--manifest` is only registered on unified commands (via `AddManifestSourceFlags` which calls `StringP`), not on direct engine commands (which register `--manifest` via plain `String` without shorthand). Similarly, `--clipboard` and `--stack-input` are only available on unified commands.

This was inaccurately documented in the previous pages and is now correctly reflected in all 8 pages.

### Information Architecture Expansion

The original plan called for 7 CLI pages (5 existing + 2 new). Analysis of the source code revealed that Terraform commands are fully implemented with 5 subcommands sharing the same execution engine as OpenTofu (`tofumodule.RunCommand` with `"terraform"` as the binary name). Added `terraform-commands.md` as the 8th page.

## Benefits

- **Single source of truth**: Flag descriptions live in one place. Updating a flag requires editing one file, not four.
- **Zero factual errors**: All `-f, -f` typos fixed. All "NEW!" labels removed. All emoji removed from prose.
- **Complete coverage**: Terraform commands documented for the first time. Module management and configuration utilities have dedicated pages.
- **Accurate flag documentation**: `-f` shorthand correctly shown as unified-command-only. Direct engine commands accurately show `--manifest` without shorthand.
- **Cross-reference integrity**: All 8 pages link to each other with relative paths, all verified to exist.

## Impact

- **New users**: Clean installation instructions, clear command overview, and obvious navigation to the right page for their use case
- **Power users**: Complete flag reference in one page (`cli-reference.md`) — no need to hunt across multiple pages
- **Evaluators**: Concise pages that demonstrate documentation quality without bloat
- **Maintainers**: Deduplication means flag changes only need to be updated in one location

## Files Changed

### New Files (3)
- `site/public/docs/cli/terraform-commands.md` — 157 lines
- `site/public/docs/cli/module-management.md` — 149 lines
- `site/public/docs/cli/configuration.md` — 142 lines

### Rewritten Files (5)
- `site/public/docs/cli/cli-reference.md` — 234 lines (was 420)
- `site/public/docs/cli/unified-commands.md` — 170 lines (was 641)
- `site/public/docs/cli/pulumi-commands.md` — 173 lines (was 858)
- `site/public/docs/cli/tofu-commands.md` — 171 lines (was 802)
- `site/public/docs/cli/index.md` — 104 lines (was 237)

### Metrics
- **Before**: 5 pages, 2,958 lines, systemic duplication
- **After**: 8 pages, 1,300 lines, zero duplication
- **Net reduction**: 1,658 fewer lines while adding 3 new pages and covering more commands

## Related Work

- Phase 0 audit (2026-02-12): Identified all the issues fixed in this rewrite
- Phase 1 concepts rewrite (2026-02-13): Established the fresh-start principle applied here
- Changelog: `2026-02-13-090244-concepts-section-complete-rewrite.md`

---

**Status**: Production Ready
**Timeline**: Single session (~2 hours)
