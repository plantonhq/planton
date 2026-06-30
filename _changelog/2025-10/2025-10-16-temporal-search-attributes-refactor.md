# Temporal Search Attributes Refactor: Enum to String Migration

**Date**: October 16, 2025  
**Type**: Breaking Change, Enhancement  
**Components**: TemporalKubernetes API, Pulumi Module

## Summary

Refactored the Temporal Kubernetes search attributes type system from enum-based to string-based validation, using Temporal's official naming convention. This eliminates conflicts with programming language keywords and provides a more intuitive, user-friendly API. Additionally, added an optional `version` field to the TemporalKubernetesSpec to enable Helm chart version control.

## Motivation

### The Problem

The previous implementation used a protobuf enum `TemporalKubernetesSearchAttributeType` with values like:
- `keyword_type`
- `text_type`
- `int_type`
- `bool_type`
- `double_type`
- `datetime_type`
- `keyword_list_type`

This approach created several issues:

1. **Language Keyword Conflicts**: Values like `int`, `text`, and `bool` are keywords in many programming languages (Java, Python, TypeScript), causing syntax errors or requiring awkward workarounds in generated code.

2. **Naming Inconsistency**: The enum values didn't match Temporal's official naming convention, forcing users to learn two different naming systems.

3. **Poor User Experience**: Users had to reference planton documentation instead of using values directly from Temporal's official documentation.

4. **Code Generation Issues**: Generated stubs in different languages had to handle keyword conflicts differently, leading to inconsistent APIs.

### The Solution

Replace the enum with a validated string field using Temporal's official type names:
- `Keyword`, `Text`, `Int`, `Double`, `Bool`, `Datetime`, `KeywordList`

This approach:
- ✅ Avoids all keyword conflicts
- ✅ Matches Temporal's official naming
- ✅ Enables users to reference Temporal docs directly
- ✅ Provides consistent experience across all languages
- ✅ Uses CEL (Common Expression Language) validation for type safety

## What's New

### 1. String-Based Search Attribute Types

**Before (Enum)**:
```protobuf
enum TemporalKubernetesSearchAttributeType {
  temporal_kubernetes_search_attribute_type_unspecified = 0;
  keyword_type = 1;
  text_type = 2;
  int_type = 3;
  double_type = 4;
  bool_type = 5;
  datetime_type = 6;
  keyword_list_type = 7;
}

message TemporalKubernetesSearchAttribute {
  string name = 1;
  TemporalKubernetesSearchAttributeType type = 2;
}
```

**After (Validated String)**:
```protobuf
message TemporalKubernetesSearchAttribute {
  string name = 1 [(buf.validate.field).required = true];
  
  string type = 2 [
    (buf.validate.field).required = true,
    (buf.validate.field).cel = {
      id: "type.valid_search_attribute_type"
      message: "type must be one of: Keyword, Text, Int, Double, Bool, Datetime, KeywordList"
      expression: "this in ['Keyword', 'Text', 'Int', 'Double', 'Bool', 'Datetime', 'KeywordList']"
    }
  ];
}
```

### 2. Optional Version Field

Added version control to TemporalKubernetesSpec:

```protobuf
message TemporalKubernetesSpec {
  // ... existing fields ...
  
  // version of the Temporal Helm chart to deploy (e.g., "0.62.0")
  // if not specified, the default version configured in the Pulumi module will be used
  string version = 9;
}
```

### 3. Updated YAML Syntax

**Before**:
```yaml
searchAttributes:
  - name: CustomerId
    type: keyword        # lowercase, with underscore for compound
  - name: Priority
    type: int
  - name: Tags
    type: keyword_list
```

**After**:
```yaml
searchAttributes:
  - name: CustomerId
    type: Keyword        # PascalCase, matches Temporal docs
  - name: Priority
    type: Int
  - name: Tags
    type: KeywordList
```

### 4. Simplified Pulumi Module

**Before** (Required Enum Mapping):
```go
func mapSearchAttributeType(attrType temporalkubernetesv1.TemporalKubernetesSearchAttributeType) string {
  switch attrType {
  case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_keyword_type:
    return "Keyword"
  case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_text_type:
    return "Text"
  // ... 7 cases total ...
  default:
    return "Keyword"
  }
}

// Usage
for _, attr := range locals.TemporalKubernetes.Spec.SearchAttributes {
  typeName := mapSearchAttributeType(attr.Type)
  searchAttrsMap[attr.Name] = pulumi.String(typeName)
}
```

**After** (Direct String Usage):
```go
// Usage - no mapping needed!
for _, attr := range locals.TemporalKubernetes.Spec.SearchAttributes {
  searchAttrsMap[attr.Name] = pulumi.String(attr.Type)
}
```

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/spec.proto`

1. **Removed**: `TemporalKubernetesSearchAttributeType` enum (26 lines)
2. **Updated**: `TemporalKubernetesSearchAttribute.type` to validated string
3. **Added**: `TemporalKubernetesSpec.version` field (optional)

**Validation Strategy**: Uses CEL (Common Expression Language) to validate the type field against allowed values, providing the same type safety as enums while avoiding code generation issues.

### Pulumi Module Updates

**File**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/helm_chart.go`

**Changes**:
1. **Removed**: `mapSearchAttributeType()` function (21 lines)
2. **Simplified**: Search attributes processing uses `attr.Type` directly
3. **Added**: Version override logic:
   ```go
   chartVersion := vars.HelmChartVersion
   if locals.TemporalKubernetes.Spec.Version != "" {
     chartVersion = locals.TemporalKubernetes.Spec.Version
   }
   ```

### Documentation Updates

Updated all documentation files with correct type names:
- `examples.md` - Example 6 with search attributes
- `README.md` - Quick start example
- `hack/manifest.yaml` - Test manifest

## Migration Guide

### Breaking Change Impact

This is a **breaking change** for existing TemporalKubernetes resources with search attributes.

**Affected Users**: Only users who have deployed Temporal with custom search attributes (introduced in October 2025).

### Migration Steps

#### Step 1: Update Your Manifests

Find all `searchAttributes` sections in your manifests and update the type values:

```bash
# Before migration
searchAttributes:
  - name: CustomerId
    type: keyword          # ❌ Old format
  - name: Priority  
    type: int              # ❌ Old format
  - name: Amount
    type: double           # ❌ Old format
  - name: IsActive
    type: bool             # ❌ Old format
  - name: DeploymentDate
    type: datetime         # ❌ Old format
  - name: Description
    type: text             # ❌ Old format
  - name: Tags
    type: keyword_list     # ❌ Old format
```

```bash
# After migration
searchAttributes:
  - name: CustomerId
    type: Keyword          # ✅ New format
  - name: Priority
    type: Int              # ✅ New format
  - name: Amount
    type: Double           # ✅ New format
  - name: IsActive
    type: Bool             # ✅ New format
  - name: DeploymentDate
    type: Datetime         # ✅ New format
  - name: Description
    type: Text             # ✅ New format
  - name: Tags
    type: KeywordList      # ✅ New format
```

**Quick Reference**:
| Old Value       | New Value     |
|----------------|---------------|
| `keyword`      | `Keyword`     |
| `text`         | `Text`        |
| `int`          | `Int`         |
| `double`       | `Double`      |
| `bool`         | `Bool`        |
| `datetime`     | `Datetime`    |
| `keyword_list` | `KeywordList` |

#### Step 2: Update CLI and Regenerate Code

```bash
# Update CLI
brew update && brew upgrade planton

# Or fresh install
brew install plantonhq/tap/planton

# Verify version
planton version

# For developers: regenerate protobuf stubs
cd apis
make protos
```

#### Step 3: Validate Your Manifests

```bash
# Test manifest validation (no deployment)
planton pulumi preview --manifest temporal.yaml --stack dev/temporal
```

If you see validation errors, check that all type values use the new PascalCase format.

#### Step 4: Apply Changes

```bash
# Update your deployment
planton pulumi up --manifest temporal.yaml --stack dev/temporal
```

**Note**: The search attributes configuration doesn't require destroying and recreating the Temporal cluster. The Pulumi module will update the Helm chart values in place.

### Automated Migration Script

For users with many manifests, use this sed script to update files:

```bash
#!/bin/bash
# migrate-temporal-types.sh

find . -name "*.yaml" -type f -exec sed -i '' \
  -e 's/type: keyword$/type: Keyword/g' \
  -e 's/type: text$/type: Text/g' \
  -e 's/type: int$/type: Int/g' \
  -e 's/type: double$/type: Double/g' \
  -e 's/type: bool$/type: Bool/g' \
  -e 's/type: datetime$/type: Datetime/g' \
  -e 's/type: keyword_list$/type: KeywordList/g' \
  {} \;

echo "✅ Migration complete!"
```

## Examples

### Complete TemporalKubernetes Manifest

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: temporal-production
spec:
  # Database configuration
  database:
    backend: postgresql
    externalDatabase:
      host: postgres.example.com
      port: 5432
      username: temporal_user
      password: secure_password
  
  # External Elasticsearch for advanced visibility
  externalElasticsearch:
    host: elasticsearch.example.com
    port: 9200
    user: elastic
    password: elastic_pass
  
  # Custom search attributes with new string-based types
  searchAttributes:
    - name: CustomerId
      type: Keyword          # Exact match, indexed
    - name: Environment
      type: Keyword          # e.g., "prod", "staging"
    - name: Priority
      type: Int              # Numeric priority level
    - name: Amount
      type: Double           # Transaction amounts
    - name: IsActive
      type: Bool             # Boolean flags
    - name: DeploymentDate
      type: Datetime         # Timestamps
    - name: Description
      type: Text             # Full-text search (requires Elasticsearch)
    - name: Tags
      type: KeywordList      # Array of keywords
  
  # Optional: specify Temporal Helm chart version
  version: "0.62.0"
  
  # Ingress configuration
  ingress:
    enabled: true
    host: temporal-prod.example.com
```

### Using Search Attributes in Temporal Workflows

Once deployed, use these search attributes in your Temporal workflows:

**Go SDK**:
```go
import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/workflow"
)

func StartWorkflow() {
    c, _ := client.NewClient(client.Options{})
    
    searchAttrs := map[string]interface{}{
        "CustomerId":     "cust-12345",
        "Environment":    "production",
        "Priority":       1,
        "Amount":         99.99,
        "IsActive":       true,
        "DeploymentDate": time.Now(),
        "Tags":           []string{"urgent", "payment"},
    }
    
    options := client.StartWorkflowOptions{
        ID:               "workflow-123",
        TaskQueue:        "my-task-queue",
        SearchAttributes: searchAttrs,
    }
    
    c.ExecuteWorkflow(context.Background(), options, MyWorkflow)
}
```

**Query in Temporal UI**:
```sql
CustomerId = 'cust-12345' AND Environment = 'production' AND Priority < 5
```

## Benefits

### 1. Eliminates Keyword Conflicts
No more issues with `int`, `bool`, `text` being reserved keywords in target languages.

### 2. Matches Temporal's Official Naming
Users can reference Temporal documentation directly without translation:
- [Temporal Search Attributes Docs](https://docs.temporal.io/visibility#search-attribute)

### 3. Cleaner Generated Code
Generated stubs in Go, Java, Python, and TypeScript are cleaner and more idiomatic.

**Before (Java)**:
```java
// Awkward enum usage
searchAttr.setType(TemporalKubernetesSearchAttributeType.keyword_type);
```

**After (Java)**:
```java
// Clean string usage
searchAttr.setType("Keyword");
```

### 4. Simplified Implementation
- Removed 21 lines of enum mapping code
- Reduced maintenance burden
- Fewer potential bugs

### 5. Better Validation Errors
CEL validation provides clear error messages:
```
type must be one of: Keyword, Text, Int, Double, Bool, Datetime, KeywordList
```

### 6. Version Control
New `version` field enables:
- Pinning to specific Temporal versions for stability
- Testing new versions in development before production
- Consistent deployments across environments

## Version Control Example

```yaml
# Development: test latest version
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: temporal-dev
spec:
  version: "0.63.0"  # Latest version
  # ... rest of config ...

---
# Production: stable version
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: temporal-prod
spec:
  version: "0.62.0"  # Proven stable
  # ... rest of config ...
```

## Related Documentation

- **Temporal Search Attributes**: https://docs.temporal.io/visibility#search-attribute
- **Temporal Type System**: https://docs.temporal.io/workflows#search-attributes
- **CEL Validation**: https://github.com/bufbuild/protovalidate
- **TemporalKubernetes API**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/`

## Testing

### Unit Test Coverage

Comprehensive unit tests have been added to validate the CEL validation rules:

**File**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/api_test.go`

**Test Coverage** (16 new test cases):

1. **Valid Search Attribute Types** (7 tests):
   - Individual test for each valid type: `Keyword`, `Text`, `Int`, `Double`, `Bool`, `Datetime`, `KeywordList`
   - Ensures each type passes validation independently

2. **Multiple Search Attributes** (1 test):
   - Tests multiple search attributes with different types together
   - Verifies array handling and type validation across multiple entries

3. **Invalid Search Attribute Types** (6 tests):
   - Lowercase variants: `keyword`, `text`, `int` (should fail)
   - Invalid type value: `InvalidType` (should fail)
   - Empty string: `""` (should fail)
   - Snake_case variant: `keyword_list` (should fail)
   - Ensures strict validation of exact type names

4. **Missing Required Fields** (2 tests):
   - Missing `type` field (should fail)
   - Missing `name` field (should fail)
   - Validates required field enforcement

**Test Results**:
```bash
cd apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1
go test -v

Running Suite: TemporalKubernetes Suite
Will run 19 of 19 specs
SUCCESS! -- 19 Passed | 0 Failed | 0 Pending | 0 Skipped
```

**Test Strategy**:
- Uses Ginkgo/Gomega BDD testing framework
- Tests use `protovalidate.Validate()` for validation
- Covers both positive (valid) and negative (invalid) cases
- Validates exact error behavior for invalid inputs

**Example Test**:
```go
ginkgo.It("should not return a validation error for Keyword type", func() {
    input.Spec.SearchAttributes = []*TemporalKubernetesSearchAttribute{
        {
            Name: "CustomerId",
            Type: "Keyword",
        },
    }
    err := protovalidate.Validate(input)
    gomega.Expect(err).To(gomega.BeNil())
})

ginkgo.It("should return a validation error for lowercase keyword type", func() {
    input.Spec.SearchAttributes = []*TemporalKubernetesSearchAttribute{
        {
            Name: "CustomerId",
            Type: "keyword",  // Invalid: must be "Keyword"
        },
    }
    err := protovalidate.Validate(input)
    gomega.Expect(err).NotTo(gomega.BeNil())
})
```

### Validation Testing

Verify CEL validation works correctly:

```yaml
# This should fail validation
searchAttributes:
  - name: Invalid
    type: integer  # ❌ Not in allowed list
```

Expected error:
```
type must be one of: Keyword, Text, Int, Double, Bool, Datetime, KeywordList
```

### Deployment Testing

Test manifest with new syntax:

```bash
# Preview changes
planton pulumi preview \
  --manifest temporal.yaml \
  --stack dev/temporal

# Deploy
planton pulumi up \
  --manifest temporal.yaml \
  --stack dev/temporal

# Verify search attributes in Temporal
kubectl exec -it temporal-frontend-0 -n temporal -- \
  tctl --ns default search-attributes list
```

## Performance Impact

No performance impact:
- Search attributes are configured once at deployment time
- Removed enum mapping function has negligible performance benefit
- CEL validation occurs at manifest load time (before deployment)

## Security Considerations

No security impact:
- Validation ensures only valid type values are accepted
- No changes to runtime behavior or data handling
- Search attributes remain namespace-scoped as before

## Deployment Status

✅ **Protobuf Contract**: Updated with string-based validation  
✅ **Pulumi Module**: Simplified, enum mapping removed  
✅ **Documentation**: All examples updated  
✅ **Migration Guide**: Complete with scripts  
✅ **Unit Tests**: 16 comprehensive test cases added (19 total tests passing)  
✅ **Validation Coverage**: All valid and invalid type scenarios tested

**Ready for**: Protobuf regeneration and deployment

## Future Enhancements

1. **CLI Validation**: Add pre-flight validation for search attribute types
2. **Fuzzy Matching**: Suggest correct type names for typos (e.g., "keyword" → "Did you mean Keyword?")
3. **Type Inference**: Detect type from value in advanced manifest syntax
4. **Documentation Links**: Add inline links to Temporal docs for each type

## Breaking Change Checklist

- [x] Migration guide provided
- [x] Documentation updated
- [x] Examples updated with new syntax
- [x] Validation in place to catch old format
- [x] Clear error messages for invalid types
- [x] Automated migration script available
- [x] Testing instructions included
- [x] Timeline communicated (effective immediately)

## Support

For questions or issues with migration:
1. Review the [migration guide](#migration-guide) above
2. Check [examples](#examples) for reference
3. Run the [automated migration script](#automated-migration-script)
4. Contact Planton support if issues persist

