<!-- 88707c44-afee-4310-8ed5-f4a90dd3dc1b c37d51d4-6b1f-448b-acbd-ce97f4cd56cb -->
# Remove Search Attributes from TemporalKubernetes

## Overview

Remove the search attributes field from the TemporalKubernetes API contract and IAC module since the Temporal Helm chart does not provide native support for configuring search attributes. This is a breaking change, but backward compatibility is not a concern as the feature hasn't been used in production yet.

## Changes Required

### 1. Update Protobuf Contract

**File**: `/Users/suresh/scm/github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/workload/temporalkubernetes/v1/spec.proto`

Remove:

- Lines 24-40: The `TemporalKubernetesSearchAttribute` message definition
- Lines 63-66: The `search_attributes` field (field 8) from `TemporalKubernetesSpec`

This will leave the spec with only the core fields: database, disable_web_ui, enableEmbeddedElasticsearch, enableMonitoringStack, cassandraReplicas, ingress, external_elasticsearch, and version.

### 2. Update IAC Module

**File**: `/Users/suresh/scm/github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/helm_chart.go`

Remove:

- Lines 157-181: The entire search attributes processing block that configures `customSearchAttributes` in the Helm chart values

The code between the elasticsearch configuration (line 155) and version configuration (line 183) should be removed.

### 3. Update Documentation

**File**: `/Users/suresh/scm/github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/workload/temporalkubernetes/v1/examples.md`

Remove:

- Lines 152-202: Example 6 that demonstrates search attributes configuration

Consider keeping the example number sequence or renumbering if there are examples after this one.

### 4. Regenerate Proto Stubs

After modifying the proto files, regenerate the Go stubs:

```bash
cd /Users/suresh/scm/github.com/plantonhq/planton/apis
make protos
```

This will update the generated `spec.pb.go` file and remove the search attribute types.

### 5. Create Changelog Entry (Optional)

If documenting this change, create a changelog entry explaining the removal and rationale (Temporal Helm chart limitation, search attributes should be configured via CLI instead).

## Testing

After the refactor:

1. Verify proto builds successfully (`make protos`)
2. Verify IAC module compiles without errors
3. Optionally test a deployment with the updated module to ensure no regression

## Migration Path

For users who might have manifests with search attributes (though unlikely based on the user's statement):

- Search attributes should be configured using Temporal CLI after deployment
- Update any existing manifests to remove the `searchAttributes` field

### To-dos

- [ ] Remove TemporalKubernetesSearchAttribute message and search_attributes field from spec.proto
- [ ] Remove search attributes processing code from helm_chart.go
- [ ] Remove Example 6 (search attributes) from examples.md
- [ ] Run make protos to regenerate Go stubs from updated proto definitions
- [ ] Verify that proto generation and IAC module compilation succeed