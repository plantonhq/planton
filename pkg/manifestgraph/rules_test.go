package manifestgraph

import (
	"strings"
	"testing"

	testgenericv1alpha2 "github.com/plantonhq/planton/catalog/_test/testcloudresourcegeneric/v1alpha2"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// compositionKeyCases is the contract for how a valueFrom is judged against
// its field's composition keys (pkg/refannotations). The platform's Java
// reader copies this table case for case (ValueFromRefsValidatorAnnotationTest),
// so a change here is a change to both lanes.
//
// The fixture fields live on the hermetic TestCloudResourceGeneric spec:
//   - candidate_ref: default TestCloudResourceGeneric `status.outputs.id`,
//     plus TestCloudResourceKubernetes through two outputs
//     (`status.outputs.external_hostname`, `status.outputs.internal_hostname`);
//   - annotated_ref: default TestCloudResourceGeneric `status.outputs.id`
//     only;
//   - optional_ref: no annotations.
var compositionKeyCases = []struct {
	name     string
	field    protoreflect.Name
	kind     cloudresourcekind.CloudResourceKind
	path     string
	wantKind cloudresourcekind.CloudResourceKind
	wantPath string
	// wantProblem is a substring of the one problem expected; empty means the
	// reference is accepted.
	wantProblem string
}{
	{
		name: "no kind and no path take the default key", field: "candidate_ref",
		wantKind: cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric, wantPath: "status.outputs.id",
	},
	{
		name: "a kind with one key defaults its path", field: "candidate_ref",
		kind:     cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric,
		wantKind: cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric, wantPath: "status.outputs.id",
	},
	{
		name: "a kind with two keys must name one", field: "candidate_ref",
		kind:        cloudresourcekind.CloudResourceKind_TestCloudResourceKubernetes,
		wantKind:    cloudresourcekind.CloudResourceKind_TestCloudResourceKubernetes,
		wantProblem: "composes from more than one output (status.outputs.external_hostname, status.outputs.internal_hostname)",
	},
	{
		name: "either key of a two-key kind is accepted", field: "candidate_ref",
		kind: cloudresourcekind.CloudResourceKind_TestCloudResourceKubernetes, path: "status.outputs.internal_hostname",
		wantKind: cloudresourcekind.CloudResourceKind_TestCloudResourceKubernetes, wantPath: "status.outputs.internal_hostname",
	},
	{
		name: "an output that is not a key of a candidate kind is refused", field: "candidate_ref",
		kind: cloudresourcekind.CloudResourceKind_TestCloudResourceKubernetes, path: "status.outputs.endpoint",
		wantKind:    cloudresourcekind.CloudResourceKind_TestCloudResourceKubernetes,
		wantPath:    "status.outputs.endpoint",
		wantProblem: `the field's contract is one of "status.outputs.external_hostname", "status.outputs.internal_hostname"`,
	},
	{
		name: "another output of the default kind is refused", field: "annotated_ref",
		kind: cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric, path: "status.outputs.name",
		wantKind:    cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric,
		wantPath:    "status.outputs.name",
		wantProblem: `the field's contract is "status.outputs.id"`,
	},
	{
		name: "a kind outside the keys is accepted with an explicit path", field: "candidate_ref",
		kind: cloudresourcekind.CloudResourceKind_KubernetesNamespace, path: "spec.name",
		wantKind: cloudresourcekind.CloudResourceKind_KubernetesNamespace, wantPath: "spec.name",
	},
	{
		name: "a kind outside the keys needs a path", field: "candidate_ref",
		kind:        cloudresourcekind.CloudResourceKind_KubernetesNamespace,
		wantKind:    cloudresourcekind.CloudResourceKind_KubernetesNamespace,
		wantProblem: "no annotated default applies",
	},
	{
		name: "a field with no annotations needs a kind", field: "optional_ref",
		wantProblem: "does not name a kind and the field declares no default kind",
	},
}

func TestCheckRef_CompositionKeys(t *testing.T) {
	spec := (&testgenericv1alpha2.TestCloudResourceGenericSpec{}).ProtoReflect().Descriptor()
	for _, tc := range compositionKeyCases {
		t.Run(tc.name, func(t *testing.T) {
			fd := spec.Fields().ByName(tc.field)
			if fd == nil {
				t.Fatalf("fixture field %s missing", tc.field)
			}
			target, problems := CheckRef(RefUse{
				FieldPath: "spec." + string(tc.field),
				Field:     fd,
				Ref:       &foreignkeyv1.ValueFromRef{Kind: tc.kind, Name: "producer", FieldPath: tc.path},
			})
			assert.Equal(t, tc.wantKind, target.Kind)
			if tc.wantProblem == "" {
				assert.Empty(t, problems)
				assert.Equal(t, tc.wantPath, target.FieldPath)
				return
			}
			if assert.Len(t, problems, 1) {
				assert.True(t, strings.Contains(problems[0], tc.wantProblem), "problem %q lacks %q", problems[0], tc.wantProblem)
			}
		})
	}
}
