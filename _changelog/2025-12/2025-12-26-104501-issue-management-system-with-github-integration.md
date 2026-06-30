# Issue Management System with GitHub Integration

**Date**: December 26, 2025
**Type**: Feature
**Components**: Developer Experience, Issue Tracking, GitHub Integration, Documentation

## Summary

Implemented a comprehensive issue management system for Planton with four cursor rules that enable structured issue tracking, image handling, GitHub issue creation, and automated issue archival. The system adapts the proven planton issue management workflow to Planton's repository structure, with deployment-component-aware area detection and CLI-focused labeling.

## Problem Statement

During development of Planton, we frequently discover bugs, feature requests, and refactoring opportunities that should be tracked for future implementation. Previously, there was no standardized process for:

### Pain Points

- **No issue tracking workflow**: Bugs and features discovered during development were lost in conversation history
- **Missing context preservation**: Important details about why issues exist and how to fix them weren't captured
- **Inconsistent documentation**: No standard structure for documenting discovered issues
- **No GitHub integration**: Manual GitHub issue creation meant context was often incomplete or copy-pasted incorrectly
- **Image evidence lost**: Screenshots and error outputs weren't systematically preserved with issues
- **No archival process**: Resolved issues remained mixed with open issues, cluttering the workspace

## Solution

Created a complete issue management system modeled after the successful planton implementation, with adaptations specific to Planton's structure and focus areas.

### System Architecture

```
planton/
├── _issues/                          # Issue tracking
│   ├── {timestamp}.{area}.{type}.{slug}.md
│   ├── images/                       # Open issue images
│   │   └── workspace/                # Staging for new images
│   └── closed/                       # Resolved issues
│       ├── {original}.{closed-timestamp}.md
│       └── images/                   # Closed issue images
├── .cursor/rules/
│   ├── issues/
│   │   ├── create-planton-issue.mdc
│   │   └── close-planton-issue.mdc
│   └── git/github/issues/
│       ├── create-planton-github-issue.mdc
│       └── generate-planton-issue-info.mdc
└── tools/local-dev/
    └── create_github_issue.py       # Non-interactive gh CLI wrapper
```

### Four Core Rules

**1. Issue Creation** (`@create-planton-issue`)
- Creates structured issue files with intelligent area detection
- Supports complete image workflow (analyze, rename, copy, reference)
- Flexible content structure (freeform or structured)
- Explicit invocation only (no auto-creation)

**2. Issue Info Generation** (`@generate-planton-issue-info`)
- Generates GitHub issue title and description as copyable code blocks
- Deployment-component-aware label inference
- Comprehensive templates for bugs, features, and tasks
- Visual enhancements (emojis, checkboxes, status indicators)

**3. GitHub Issue Creation** (`@create-planton-github-issue`)
- Non-interactive GitHub issue creation via gh CLI
- Auto-infers labels from file paths and context
- Deterministic Python script for reliable execution
- Optional assignees, milestones, projects, browser opening

**4. Issue Closure** (`@close-planton-issue`)
- Moves resolved issues to `_issues/closed/` with closing timestamp
- Two modes: with resolution context (recent fix) or simple close
- Relocates associated images maintaining references
- Links to changelogs when available

## Implementation Details

### Area Detection System

Adapted from planton to match Planton's structure:

**Planton Areas**:
- `deployment-component` - Deployment component changes
- `cli` - CLI command and flag changes
- `pkg` - Package and library changes
- `forge` - Deployment component forge system
- `apis` - API and protobuf changes
- `site` - Planton website
- `tooling` - Build tools and scripts
- `repo` - Repository-wide changes

**File Path Mapping**:
```
apis/dev/planton/provider/**  → deployment-component
cmd/planton/**                → cli
pkg/**                                → pkg
_rules/deployment-component/forge/** → forge
site/**                               → site
tools/**, hack/**                     → tooling
```

**Keyword Detection**:
```
"deployment component", "IaC", "Pulumi" → deployment-component
"CLI", "command line"                   → cli
"forge", "code generation"              → forge
"package", "library", "pkg/"            → pkg
```

### Issue File Naming Convention

```
YYYY-MM-DD-HHMMSS.{area}.{type}.{slug}.md
```

**Components**:
- **Timestamp**: Actual timestamp from `date +"%Y-%m-%d-%H%M%S"`
- **Area**: Simple identifier (deployment-component, cli, pkg, forge, etc.)
- **Type**: Issue category (feat, bug, refactor, docs, test, perf, chore)
- **Slug**: Descriptive identifier (30-50 chars, kebab-case)

**Examples**:
```
2025-12-26-143022.deployment-component.bug.postgres-spec-validation.md
2025-12-26-150815.cli.feat.manifest-validation-command.md
2025-12-26-162430.forge.bug.pulumi-code-generation.md
2025-12-26-091205.pkg.refactor.kubernetes-client-helpers.md
```

### Image Handling Workflow

9-step process for seamless image integration:

1. **Ask user**: Confirm images in workspace directory
2. **List images**: Read `_issues/images/workspace/`
3. **Analyze**: Understand what each image shows
4. **Rename**: Apply convention `{issue-filename}.{image-slug}.{ext}`
5. **Reference**: Add markdown links in issue content
6. **Copy**: Move to `_issues/images/` directory
7. **Verify**: Confirm successful copy
8. **Cleanup**: Remove from workspace
9. **Confirm**: Report to user

**Image Naming Example**:
```
Original: screenshot-1.png
Renamed:  2025-12-26-143022.deployment-component.bug.postgres-spec-validation.error-output.png
```

**Markdown Reference**:
```markdown
![Pulumi error output](images/2025-12-26-143022.deployment-component.bug.postgres-spec-validation.error-output.png)
```

### GitHub Label Inference

Deployment-component and provider-aware labeling:

**Area Labels**:
```
apis/dev/planton/provider/** → area/deployment-component,area/<provider>
cmd/planton/**              → area/cli
pkg/**                              → area/pkg
.cursor/rules/.../forge/**          → area/forge
```

**Component Labels**:
```
kubernetes deployments → kubernetes
aws deployments       → aws
gcp deployments       → gcp
pulumi-related        → pulumi
terraform-related     → terraform
```

**Example Label Combination**:
```
bug,area/deployment-component,area/kubernetes,priority/critical,P0
```

### Issue Closure Workflow

**Scenario A** (with resolution context):
1. Get closing timestamp
2. Read original issue file
3. Append resolution section with changelog links
4. Move images to `_issues/closed/images/`
5. Write updated content to closed location
6. Delete original file from `_issues/`
7. Verify both files exist in closed/ and original is gone

**Scenario B** (simple close):
1. Get closing timestamp
2. Move images to `_issues/closed/images/`
3. Move issue file to closed location with timestamp
4. Verify moved successfully

**Resolution Section Example**:
```markdown
---

## Resolution

**Closed**: December 28, 2025

This issue was fixed by adding proto validation rules to the `spec.proto` file 
that check for conflicting port configurations before deployment.

### Changes Made

Added buf-validate constraints to ensure default and custom ports don't conflict, 
catching errors at validation time rather than during Pulumi deployment.

### Changelog

See [Postgres Spec Validation Fix](../../_changelog/2025-12/2025-12-28-073049-postgres-spec-validation-fix.md) for detailed implementation.
```

## GitHub Integration Tooling

### create_github_issue.py

Non-interactive Python script that wraps gh CLI:

**Features**:
- Explicit command-line arguments (no interactive prompts)
- Support for title, body-file, labels, assignees, milestone, project
- Opens in browser with `--web` flag
- Comprehensive error handling and prerequisite checks
- Deterministic execution (no LLM-to-shell flakiness)

**Example Usage**:
```bash
python3 tools/local-dev/create_github_issue.py \
  --title "Postgres deployment component spec validation broken" \
  --body-file .cursor/workspace/issue-description.md \
  --labels "bug,area/deployment-component,priority/high,kubernetes" \
  --web
```

## Benefits

### For Issue Discovery
- **Structured tracking**: Every issue follows consistent pattern
- **Complete context**: Future implementers have all information needed
- **Visual evidence**: Screenshots embedded where they provide value
- **Easy categorization**: Area and type in filename enable quick filtering
- **Intelligent defaults**: AI-powered area detection reduces manual decisions

### For Issue Management
- **Chronological sorting**: Timestamp ensures issues sort by discovery date
- **Searchable filenames**: Descriptive slugs make issues easy to find
- **Flexible structure**: Adapts to simple or complex issues appropriately
- **Quality checklist**: Ensures no critical details are missed
- **Archival workflow**: Closed issues separated with resolution context

### For GitHub Integration
- **Comprehensive issues**: Auto-generated content includes all investigation findings
- **Smart labeling**: Auto-inferred labels from file paths and context
- **Non-interactive**: Deterministic creation via Python script
- **Context preservation**: All technical details preserved in issue body
- **Time savings**: No manual copy-paste or formatting

### For Developer Experience
- **Automated workflow**: Rules handle all mechanics
- **Image analysis**: AI understands image content for better naming
- **No context loss**: Visual + written evidence captured together
- **Clear guidance**: Comprehensive examples and patterns
- **Explicit invocation**: User controls when issues are created/closed

## Impact

### Immediate Benefits
- **Standardized process**: Issue creation follows consistent workflow
- **Better tracking**: Issues are organized and discoverable
- **GitHub integration**: Seamless issue creation with proper labeling
- **Context preservation**: Nothing lost in conversation history

### Long-term Benefits
- **Reduced debugging time**: Future developers have complete information
- **Faster implementation**: Clear acceptance criteria and file references
- **Better prioritization**: Impact and severity clearly documented
- **Knowledge preservation**: Visual + written context captured systematically

### Developer Workflow
- **Easy invocation**: Simply use `@create-planton-issue`
- **Guided process**: Rules ask questions and provide defaults
- **Automated naming**: Timestamp, area, type, slug all handled
- **Image workflow**: Complete automation from workspace to embedding
- **Quality assurance**: Checklist ensures completeness

## Usage Examples

### Issue Discovery During Development

```
User: "I noticed the Postgres component validation is broken. @create-planton-issue"

Agent:
- Gets timestamp: 2025-12-26-143022
- Detects area: deployment-component (Postgres component)
- Determines type: bug
- Creates slug: postgres-spec-validation
- Asks about images
- Analyzes, renames, copies, references images
- Creates: _issues/2025-12-26-143022.deployment-component.bug.postgres-spec-validation.md
```

### Closing Issue After Fix

```
User: "@close-planton-issue postgres-spec-validation"

Agent:
- Gets timestamp: 2025-12-28-073049
- Finds recent changelog: postgres-spec-validation-fix.md
- Detects Scenario A (recent fix)
- Adds resolution section with changelog link
- Moves images to closed/
- Writes to: _issues/closed/2025-12-26-143022...2025-12-28-073049.md
- Deletes original
- Confirms to user
```

### Creating GitHub Issue

```
User: "@create-planton-github-issue"

Agent:
- Generates via @generate-planton-issue-info
- Title: "Postgres deployment component spec validation broken for port conflicts"
- Infers labels: bug,area/deployment-component,area/kubernetes,priority/critical,P0
- Writes body to .cursor/workspace/issue-description.md
- Executes Python script
- Returns GitHub issue URL
```

## Files Created

### Rules
- `_rules/issues/create-planton-issue.mdc` (653 lines)
- `_rules/issues/close-planton-issue.mdc` (662 lines)
- `_rules/git/github/issues/generate-planton-issue-info.mdc` (341 lines)
- `_rules/git/github/issues/create-planton-github-issue.mdc` (281 lines)

### Tooling
- `tools/local-dev/create_github_issue.py` (265 lines)

### Directory Structure
- `_issues/` - Open issues directory
- `_issues/images/` - Images for open issues
- `_issues/images/workspace/` - Staging area for new images
- `_issues/closed/` - Closed/resolved issues
- `_issues/closed/images/` - Images for closed issues

### Documentation
- Updated `.cursor/rules/README.md` with issue management section

## Planton Adaptations

Key adaptations from planton to planton:

**Area Changes**:
```
planton              → planton
-----------------          → -------------------
console                    → cli (primary interface)
backend                    → deployment-component (core artifacts)
infra-hub                  → forge (code generation system)
planton (area)     → pkg (libraries)
```

**File Path Mappings**:
```
planton                          → planton
---------------------------------      → ----------------------------------
client-apps/web/console/**            → cmd/planton/**
backend/services/**                    → apis/dev/planton/provider/**
.cursor/rules/product/.../forge/**    → .cursor/rules/.../forge/**
```

**Component Labels**:
```
Added for planton:
- kubernetes (for K8s providers)
- aws, gcp, azure (for cloud providers)
- pulumi, terraform (for IaC backends)
```

## Design Decisions

### Why Explicit Invocation Only?

- Prevents noise from trivial discoveries
- User controls what gets tracked
- Matches changelog rule philosophy
- Avoids accidentally creating duplicate issues

### Why Simple Area Names?

- Easier to type and remember than full paths
- Maps naturally to project structure
- Consistent with existing area detection patterns
- Flexible for multi-area issues

### Why Flexible Content Structure?

- Simple issues don't need heavy structure
- Complex issues benefit from sections
- Examples guide both approaches
- Writer chooses appropriate level of detail

### Why 9-Step Image Workflow?

- Ensures no images are lost
- Verification at each critical step
- Clean separation (workspace → images)
- Descriptive naming for discoverability
- Markdown references keep content portable

### Why Two Closure Scenarios?

- Fresh context (Scenario A) deserves documentation
- Old issues (Scenario B) don't need speculation
- Resolution links provide full details
- Safer to omit context than guess incorrectly

## Related Work

This implementation builds on:

- **planton issue system**: Proven workflow adapted for planton
- **planton changelog rule**: Similar structure and philosophy
- **Existing PR info rule**: Area detection patterns
- **GitHub CLI (gh)**: Non-interactive issue creation

## Testing and Validation

Created test issue to validate the system:
- Successfully created deployment-component issue
- Area detection worked correctly
- File naming convention applied
- Quality checklist verified

## Future Enhancements

Potential improvements:

1. **GitHub sync**: Auto-create GitHub issues from local issue files
2. **Issue templates**: Specialized templates for security, performance, etc.
3. **Related issue detection**: Find similar existing issues
4. **Priority scoring**: Auto-suggest priority based on keywords and context
5. **Image optimization**: Compress images during workflow
6. **OCR integration**: Extract text from screenshots for searchability
7. **Issue statistics**: Dashboard of open/closed issues by area

## Documentation

All rules include:
- Comprehensive usage guidelines
- Area detection heuristics
- File naming conventions
- Quality checklists
- Extensive examples
- Troubleshooting guidance
- Writing guidelines

Updated `.cursor/rules/README.md` with:
- Issue management workflow overview
- All four rules documented
- Usage examples
- Integration patterns

## Quality Metrics

**Rules created**: 4 comprehensive rules (1,937 total lines)
**Tooling added**: 1 Python script (265 lines)
**Documentation**: Updated README with complete workflow guide
**Directory structure**: 5 directories with .gitkeep files
**Examples**: 12+ detailed usage examples across all rules
**Quality checks**: 4 comprehensive checklists

## Remember

**Issue files are discovery artifacts** that capture:
- What problem or opportunity was found
- Why it matters
- What context is needed to address it
- Who is impacted

**Write for the person who will implement this later.** Give them:
- **Context**: Why does this issue exist?
- **Clarity**: What exactly needs to be done?
- **Completeness**: What information do they need?

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation
**Files changed**: 11 (4 rules + 1 script + 5 .gitkeep + 1 README)

*"Documentation is a love letter to your future self."* - Damian Conway

