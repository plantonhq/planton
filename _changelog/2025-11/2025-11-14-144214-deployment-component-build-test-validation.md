# Deployment Component Rules: Mandatory Build and Test Validation

**Date**: November 14, 2025
**Type**: Enhancement
**Components**: Deployment Component Rules, Build System, Testing Framework, Development Process

## Summary

Enhanced the deployment component rules (update, fix, complete) to mandate explicit execution of `make protos`, component-specific tests, `make build`, and `make test` commands at appropriate checkpoints. This ensures consistency across all deployment component operations and prevents build/test failures from being discovered late in the development cycle.

## Problem Statement / Motivation

The deployment component rules previously provided guidance on when to validate changes, but lacked explicit command execution requirements. This created several issues:

### Pain Points

- **Ambiguous validation steps**: Rules mentioned "validate build" without specifying the exact command to run
- **Inconsistent workflows**: Different developers followed different validation sequences
- **Late failure discovery**: Build and test failures were often discovered after completing significant work
- **Missing proto regeneration**: Developers sometimes forgot to run `make protos` after proto changes
- **Incomplete test coverage**: Component-specific tests (spec_test.go) weren't consistently executed to validate buf.validate rules
- **No clear sequencing**: The order of validation commands wasn't explicitly defined

These gaps led to rework, wasted time, and inconsistent quality across deployment components.

## Solution / What's New

We've updated all deployment component rules to include explicit, mandatory command execution at specific checkpoints:

### Command Execution Requirements

**After Proto File Changes:**
```bash
make protos  # Regenerate Go stubs from proto definitions
go test ./apis/dev/planton/provider/<provider>/<component>/v1/  # Validate buf.validate rules
make build   # Ensure complete build succeeds
```

**After Pulumi/Go Code Changes:**
```bash
make build   # Validate complete build
go test ./apis/dev/planton/provider/<provider>/<component>/v1/  # Validate component tests
```

**Always (Final Validation):**
```bash
make test    # Run full test suite to catch regressions
```

### Updated Rules

1. **update-planton-component.mdc**
   - Proto Changed scenario now explicitly calls out `make protos`, component tests, `make build`, and `make test`
   - Update IaC scenario includes `make build` requirement
   - Validation checkpoints table expanded with command column
   - Examples updated to show exact command sequences

2. **fix-planton-component.mdc**
   - Source code fix sections now mandate specific commands based on change type
   - Step 5 (Comprehensive Validation) rewritten with conditional command execution
   - All scenario examples updated with complete command sequences
   - Validation section emphasizes sequential execution

3. **complete-planton-component.mdc**
   - Validation subsection enhanced with sequential command execution
   - Progress tracking examples updated to show all validation steps
   - Success criteria expanded to include all validation commands

4. **update/README.md**
   - Scenario descriptions updated with explicit commands
   - Validation Checkpoints section completely rewritten with command table
   - Build and Test Execution section added with complete sequence

## Implementation Details

### Key Changes by Rule File

#### update-planton-component.mdc (Primary Rule)

**Step 2: Determine Update Scope - Proto Changed**
```markdown
- **Must execute**: `make protos` to regenerate Go stubs
- **Must execute**: `go test ./apis/dev/planton/provider/<provider>/<component>/v1/` to validate spec tests
- **Must execute**: `make build` for complete build validation
```

**Step 4: Final Validation**
```markdown
1. **If proto files changed**: Run `make protos` to regenerate Go stubs
2. **Always run component tests**: `go test ./apis/dev/planton/provider/<provider>/<component>/v1/`
   - This validates buf.validate rules in spec.proto
   - Ensures spec_test.go passes with current validation logic
3. **If Pulumi/Go code changed**: Run `make build` to validate complete build (rule 018)
4. **Always run full test suite**: Run `make test` to validate all tests (rule 019)
```

**Validation Checkpoints Section**
Completely rewritten to include:
- Conditional execution based on change type (proto vs Go code)
- Specific commands with purpose explained
- Critical note about component tests validating buf.validate rules

#### fix-planton-component.mdc

**Step 2: Make the Fix in Source Code**
```markdown
1. **Fix Proto Schema** (if proto change needed)
   - **Must execute**: `make protos` from project root to regenerate Go stubs
   - **Must execute**: `go test ./apis/dev/planton/provider/<provider>/<component>/v1/` to validate spec tests

2. **Fix IaC Modules** (if deployment logic change needed)
   - **Must execute**: `make build` from project root to validate complete build
```

**Step 5: Run Comprehensive Validation**
```bash
# Execute in order (based on what was changed):

# 1. If proto files changed: Regenerate Go stubs
make protos

# 2. Component-specific tests (validates buf.validate rules)
go test ./apis/dev/planton/provider/<provider>/<component>/v1/

# 3. If Pulumi/Go code changed: Build validation
make build

# 4. Full test suite (validates all tests)
make test
```

All six scenario examples updated with 8-10 step sequences including specific commands.

#### complete-planton-component.mdc

**Validation Section**
```markdown
**Validation (Execute in Sequence):**
- After proto file creation/changes: Run `make protos` to regenerate Go stubs
- After each creation: Validate builds and tests pass
- Run component-specific tests: `go test ./apis/dev/planton/provider/<provider>/<component>/v1/`
  - Validates buf.validate rules in spec.proto are correct
  - Validates spec_test.go tests pass with current validation logic
- After Pulumi/Go code changes: Run `make build` to validate complete build
- Run full test suite: `make test` to validate all tests pass
```

Progress tracking examples updated from 13 steps to 14 steps, adding proto regeneration explicitly.

#### update/README.md

**Validation Checkpoints Table**
| Checkpoint | Command | Validates | Fails If |
|------------|---------|-----------|----------|
| After proto changes | `make protos` | Proto compiles, stubs generated | Import errors, syntax errors |
| Component tests | `go test ./apis/.../v1/` | buf.validate rules work | Any spec_test.go failure |
| After Go/Pulumi changes | `make build` | Complete build succeeds | Compilation errors |
| Final validation | `make test` | Full test suite passes | Any test failure |

Complete bash code block added showing the full sequence with comments.

## Benefits

### For Developers

- **No ambiguity**: Exact commands to run at each checkpoint
- **Consistent workflow**: Same process for update, fix, and complete operations
- **Early failure detection**: Build/test issues caught immediately after changes
- **Better understanding**: Comments explain what each command validates
- **Reduced rework**: No discovering failures after completing hours of work

### For Code Quality

- **Mandatory proto regeneration**: Can't forget to run `make protos`
- **Validation coverage**: buf.validate rules tested via component tests
- **Build confidence**: `make build` ensures complete compilation
- **Regression prevention**: `make test` catches unintended side effects

### For Process

- **Teachable workflow**: New contributors have clear steps to follow
- **Audit trail**: Rules document the exact validation sequence
- **Automation foundation**: Clear commands enable CI/CD integration
- **Rule consistency**: Update, fix, and complete all follow same pattern

## Impact

### Developer Experience

All developers working on deployment components now have:
- Clear validation checkpoints with specific commands
- Conditional execution guidance (proto vs Go changes)
- Sequential ordering that prevents dependency issues
- Explicit connection between commands and what they validate

### Documentation Quality

The enhanced rules provide:
- Executable examples in code blocks
- Before/after context for when to run commands
- Purpose statements for each validation step
- Cross-references to forge flow rules (017-019)

### Workflow Reliability

The mandatory commands ensure:
- Proto stubs are always in sync with proto definitions
- Component tests validate buf.validate rules work correctly
- Complete builds succeed before moving forward
- Full test suite passes to catch regressions

## Usage Example

### Scenario: Adding a new field to spec.proto

**Before this enhancement** (ambiguous):
```bash
# Developer might do:
vi spec.proto  # Add field
# ... then update examples, maybe build, maybe test, maybe forget protos
```

**After this enhancement** (explicit):
```bash
# 1. Make proto change
vi apis/dev/planton/provider/gcp/gcpcloudrun/v1/spec.proto

# 2. Execute mandatory validation sequence
make protos  # Regenerate Go stubs

# 3. Validate component tests
go test ./apis/dev/planton/provider/gcp/gcpcloudrun/v1/

# 4. Update related files (Terraform, examples)
vi apis/dev/planton/provider/gcp/gcpcloudrun/v1/iac/tf/variables.tf
vi apis/dev/planton/provider/gcp/gcpcloudrun/v1/examples.md

# 5. Validate complete build
make build

# 6. Final validation
make test
```

### Scenario: Fixing Pulumi implementation

```bash
# 1. Make code change
vi apis/dev/planton/provider/gcp/gcpcloudrun/v1/iac/pulumi/module/service.go

# 2. Validate build immediately
make build

# 3. Run component tests
go test ./apis/dev/planton/provider/gcp/gcpcloudrun/v1/

# 4. Update documentation
vi apis/dev/planton/provider/gcp/gcpcloudrun/v1/iac/pulumi/README.md

# 5. Final validation
make test
```

## Related Work

This enhancement builds on existing infrastructure:
- **Rule 017**: Generate Proto Stubs (`make protos`) - now explicitly required after proto changes
- **Rule 018**: Build Validation (`make build`) - now mandatory after Go/Pulumi changes
- **Rule 019**: Test Validation (`make test`) - now required as final checkpoint
- **Component spec_test.go files**: Now explicitly tied to buf.validate rule validation

Complements:
- **forge-planton-component**: Flow rules already include these commands; update/fix/complete now consistent
- **audit-planton-component**: Audit can verify these validation steps were followed

## Code Metrics

**Files Modified**: 4
- `_rules/deployment-component/update/update-planton-component.mdc` (+108 lines)
- `_rules/deployment-component/fix/fix-planton-component.mdc` (+95 lines)
- `_rules/deployment-component/complete/complete-planton-component.mdc` (+42 lines)
- `_rules/deployment-component/update/README.md` (+35 lines)

**Total Enhancement**: +280 lines of documentation and guidance

**Commands Standardized**: 4 core commands now explicitly documented
- `make protos`
- `go test ./apis/dev/planton/provider/<provider>/<component>/v1/`
- `make build`
- `make test`

**Scenarios Updated**: 18 scenario examples across all rule files

## Future Enhancements

Potential follow-ups:
- **CI/CD integration**: Automate these checks in GitHub Actions
- **Pre-commit hooks**: Validate proto changes trigger `make protos`
- **Rule enforcement**: Script to verify command execution in PRs
- **Metrics collection**: Track validation failures by checkpoint
- **IDE integration**: Cursor/VS Code tasks for common sequences

## Backward Compatibility

**Fully backward compatible**: 
- No breaking changes to existing commands or tools
- Enhancement is additive (adds clarity, doesn't change behavior)
- Existing workflows still work; they're now better documented
- Previous rule versions remain valid, just less explicit

Developers who were already following best practices see their workflow formally documented. Those who weren't now have clear guidance.

---

**Status**: ✅ Production Ready
**Impact Level**: Medium - Affects all deployment component development workflows
**Adoption**: Immediate - Rules take effect for all future update/fix/complete operations

