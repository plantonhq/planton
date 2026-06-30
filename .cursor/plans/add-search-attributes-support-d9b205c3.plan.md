<!-- d9b205c3-fd19-41b8-bfd6-e1157866d5a6 f814989d-027d-4b23-8ff0-d9674588aa03 -->
# Add Search Attributes Support to Temporal Module

## Summary

Enable users to configure custom search attributes in Temporal deployments. Search attributes are indexed fields that allow filtering and querying workflows in the Temporal UI and via SDK. This enhancement adds the capability at both the API contract level (proto) and the IaC implementation level (Pulumi module).

## Background: Search Attributes in Temporal

**What they are**: Key-value pairs attached to workflow executions that enable filtering, searching, and querying.

**Two operational aspects**:

1. **Cluster/namespace-level registration** (admin): Define which custom attributes exist and their types
2. **Workflow-level usage** (application): Set values on individual executions (not handled by IaC)

**Visibility backends**:

- **Standard SQL** (PostgreSQL/MySQL): Basic key-value filtering, requires schema updates
- **Elasticsearch**: Advanced queries, full-text search, aggregations
- Both can support search attributes, but Elasticsearch offers richer capabilities

**Our focus**: Enable users to declare custom search attributes in their manifest so the IaC module registers them during deployment.

## Changes Required

### 1. Update Protobuf Contract (`spec.proto`)

**File**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/spec.proto`

**Add new enum and message** after line 22 (after `TemporalKubernetesDatabaseBackend`):

```protobuf
// temporal kubernetes search attribute type enumerates supported types
enum TemporalKubernetesSearchAttributeType {
  // unspecified should not be used
  temporal_kubernetes_search_attribute_type_unspecified = 0;
  
  // keyword type - exact match string (indexed, not analyzed)
  keyword = 1;
  
  // text type - full-text searchable string (analyzed, only works with Elasticsearch)
  text = 2;
  
  // int type - 64-bit integer
  int = 3;
  
  // double type - floating point number
  double = 4;
  
  // bool type - boolean value
  bool = 5;
  
  // datetime type - timestamp
  datetime = 6;
  
  // keyword_list type - array of keywords
  keyword_list = 7;
}

// defines a custom search attribute for Temporal workflows
message TemporalKubernetesSearchAttribute {
  // name of the search attribute (e.g., "CustomerId", "Environment")
  string name = 1 [(buf.validate.field).required = true];
  
  // type of the search attribute
  TemporalKubernetesSearchAttributeType type = 2 [
    (buf.validate.field).required = true
  ];
}
```

**Update `TemporalKubernetesSpec` message** (after line 48, before closing brace):

```protobuf
  // custom search attributes to register in the Temporal cluster
  // these attributes can be used to filter and query workflows
  // note: 'text' type only works with Elasticsearch backend
  repeated TemporalKubernetesSearchAttribute search_attributes = 8;
```

### 2. Regenerate Go Stubs

**Command**:

```bash
cd /Users/suresh/scm/github.com/plantonhq/planton/apis
make protos
```

This generates updated `spec.pb.go` with the new enum and message types.

### 3. Update Pulumi Module (Helm Chart Configuration)

**File**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/helm_chart.go`

**After line 155** (after elasticsearch configuration, before chart installation), add:

```go
	// ------------------------------------------------- search attributes
	if len(locals.TemporalKubernetes.Spec.SearchAttributes) > 0 {
		searchAttrsMap := pulumi.Map{}
		for _, attr := range locals.TemporalKubernetes.Spec.SearchAttributes {
			typeName := mapSearchAttributeType(attr.Type)
			searchAttrsMap[attr.Name] = pulumi.String(typeName)
		}
		
		// Configure via server dynamic config
		if serverCfg, ok := values["server"].(pulumi.Map); ok {
			if configMap, ok := serverCfg["config"].(pulumi.Map); ok {
				configMap["customSearchAttributes"] = searchAttrsMap
			} else {
				serverCfg["config"] = pulumi.Map{
					"customSearchAttributes": searchAttrsMap,
				}
			}
		} else {
			values["server"] = pulumi.Map{
				"config": pulumi.Map{
					"customSearchAttributes": searchAttrsMap,
				},
			}
		}
	}
```

**Add helper function** at the end of `helm_chart.go` (after line 174):

```go
// mapSearchAttributeType converts proto enum to Temporal search attribute type string
func mapSearchAttributeType(attrType temporalkubernetesv1.TemporalKubernetesSearchAttributeType) string {
	switch attrType {
	case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_keyword:
		return "Keyword"
	case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_text:
		return "Text"
	case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_int:
		return "Int"
	case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_double:
		return "Double"
	case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_bool:
		return "Bool"
	case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_datetime:
		return "Datetime"
	case temporalkubernetesv1.TemporalKubernetesSearchAttributeType_keyword_list:
		return "KeywordList"
	default:
		return "Keyword"
	}
}
```

### 4. Update Documentation

**File**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/README.md`

**Add new section** after "Observability and Monitoring" (around line 35):

```markdown
### Search Attributes

- **Custom Search Attributes**: Define indexed fields for filtering and querying workflows in the Temporal UI and SDK.
- **Type Safety**: Support for all Temporal search attribute types (Keyword, Text, Int, Double, Bool, Datetime, KeywordList).
- **Backend Compatibility**: Works with both SQL and Elasticsearch visibility stores (note: Text type requires Elasticsearch).
```

**Add new example** in the "Usage Examples" section (after existing examples):

````markdown
### Example: Deploying with Custom Search Attributes

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: TemporalKubernetes
metadata:
  name: temporal-with-search-attrs
spec:
  database:
    backend: postgresql
    externalDatabase:
      host: "postgres.example.com"
      port: 5432
      username: "temporal_user"
      password: "secure_password"
  
  externalElasticsearch:
    host: "elasticsearch.example.com"
    port: 9200
    user: "elastic"
    password: "elastic_pass"
  
  searchAttributes:
    - name: CustomerId
      type: keyword
    - name: Environment
      type: keyword
    - name: Priority
      type: int
    - name: Amount
      type: double
    - name: IsActive
      type: bool
    - name: DeploymentDate
      type: datetime
    - name: Tags
      type: keyword_list
````

```

### 5. Update Examples File

**File**: `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/examples.md`

Add the same search attributes example from step 4.

### 6. Testing

**Create or update spec test** if exists:

1. Test valid search attribute configurations
2. Test required field validations
3. Test with different backend combinations (SQL vs Elasticsearch)
4. Test that Text type works with Elasticsearch but has fallback behavior with SQL

**Manual verification** (after implementation):

1. Deploy with the hack manifest including search attributes
2. Verify Helm values include `server.config.customSearchAttributes`
3. Check Temporal Web UI to confirm custom search attributes are registered
4. Test workflow filtering with custom attributes

## Implementation Order

1. Update `spec.proto` with enum, message, and field
2. Run `make protos` to regenerate Go stubs
3. Update Pulumi `helm_chart.go` with search attributes configuration
4. Update `README.md` and `examples.md` documentation
5. Test with a sample manifest containing search attributes
6. Verify in Temporal UI that attributes are registered

## Notes

- **Text type caveat**: The `Text` type only works with Elasticsearch. If users specify `text` with SQL backend, Temporal will fall back to `Keyword` behavior (exact match).
- **No credential changes needed**: Your colleague's concern about "giving credentials" for Elasticsearch is already handled - the existing `external_elasticsearch` field provides those credentials.
- **Registration vs Usage**: This IaC change handles registration (admin operation). Application code still needs to set attribute values on workflows.
- **Backwards compatible**: Adding the optional `search_attributes` field won't break existing deployments.

### To-dos

- [ ] Add TemporalKubernetesSearchAttributeType enum and TemporalKubernetesSearchAttribute message to spec.proto, plus search_attributes field to TemporalKubernetesSpec
- [ ] Run 'make protos' in apis/ directory to regenerate Go stub files from updated proto definitions
- [ ] Update helm_chart.go to configure search attributes in Helm values and add mapSearchAttributeType helper function
- [ ] Update README.md and examples.md with search attributes documentation and usage examples
- [ ] Test with sample manifest containing search attributes and verify in Temporal deployment