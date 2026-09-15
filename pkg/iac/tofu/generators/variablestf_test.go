package generators

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"

	testkubernetesv1 "github.com/plantonhq/planton/catalog/_test/testcloudresourcekubernetes/v1alpha1"
	awsecrrepov1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecrrepo/v1alpha1"
	awsiamrolev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamrole/v1alpha1"
	awsroute53zonev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsroute53zone/v1alpha1"
	awssubnetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssubnet/v1alpha1"
	kubernetescronjobv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescronjob/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestProtoToVariablesTF_FreeFormJsonMapIsAnyNotMapAny(t *testing.T) {
	// inline_policies is map<string, google.protobuf.Struct>: each entry is an
	// independently-shaped JSON document, so it must be typed `any` (optional with a {}
	// default), never map(any) -- map(any) forces a single common element type and a
	// Terraform module fails input validation with "all map elements must have the same
	// type" the moment two differently-shaped policies are passed. trust_policy is a single
	// Struct and stays `any`.
	msg := &awsiamrolev1alpha1.AwsIamRole{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	if !strings.Contains(got, "inline_policies = optional(any, {})") {
		t.Errorf("inline_policies (map<string, Struct>) must render as optional(any, {}), got:\n%s", got)
	}
	if strings.Contains(got, "map(any)") {
		t.Errorf("a free-form JSON map must not render as map(any) (heterogeneous values), got:\n%s", got)
	}
}

func TestTFFreeFormMap_RendersAnyWithEmptyMapDefault(t *testing.T) {
	if got := (TFFreeFormMap{}).Format(1); got != "any" {
		t.Errorf("TFFreeFormMap.Format = %q, want \"any\"", got)
	}
	def, ok := zeroDefaultLiteral(TFFreeFormMap{})
	if !ok || def != "{}" {
		t.Errorf("zeroDefaultLiteral(TFFreeFormMap) = (%q, %v), want (\"{}\", true)", def, ok)
	}
	// As an optional object attribute it must read optional(any, {}).
	obj := TFObject{Fields: []TFField{{Name: "inline_policies", Type: TFFreeFormMap{}, Optional: true}}}
	if got := obj.Format(0); !strings.Contains(got, "inline_policies = optional(any, {})") {
		t.Errorf("optional free-form map should render optional(any, {}), got:\n%s", got)
	}
}

func TestProtoToVariablesTF_CronJob_NamespaceIsString(t *testing.T) {
	msg := &kubernetescronjobv1alpha1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	// namespace should be flattened to string, not object({value = string, ...}).
	if strings.Contains(got, "namespace = object(") {
		t.Errorf("namespace should be 'string' (flattened from StringValueOrRef), not an object:\n%s", got)
	}
	// It is required (presence-constrained) so it stays bare, or optional with a
	// string default; either way the element type is string, never an object.
	if !strings.Contains(got, "namespace = string") && !strings.Contains(got, `namespace = optional(string, "")`) {
		t.Errorf("namespace should be a flattened string in the spec object, got:\n%s", got)
	}
}

func TestProtoToVariablesTF_RefMapVariablesIsMapString(t *testing.T) {
	// The _test torture kind carries ref_map (map<string, StringValueOrRef>)
	// specifically to pin the map-value flatten rule; live catalog specs drop
	// and rename their map fields over time, but this fixture cannot move.
	msg := &testkubernetesv1.TestCloudResourceKubernetes{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	// A map<string, StringValueOrRef> field flattens to map(string). Being
	// optional, it is wrapped with a {} zero default so a pruned tfvars
	// validates -- but the element type must remain map(string), never an
	// object, and maps must never be mis-rendered as object({key, value}).
	if !strings.Contains(got, "ref_map = optional(map(string), {})") {
		t.Errorf("ref_map should be 'optional(map(string), {})', got:\n%s", got)
	}
	if strings.Contains(got, "map(object(") {
		t.Errorf("a map field must not be modeled as map(object(...)), got:\n%s", got)
	}
}

// TestProtoToVariablesTF_ManifestOnlyFieldHasNoVariable asserts a field marked
// (dev.planton.shared.options.manifest_only) is not declared in variables.tf:
// the tfvars converter never sends it, so a declared attribute would be dead
// on every module. Its unmarked sibling is declared as usual.
func TestProtoToVariablesTF_ManifestOnlyFieldHasNoVariable(t *testing.T) {
	got, err := ProtoToVariablesTF(&testkubernetesv1.TestCloudResourceKubernetes{})
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}
	spec := extractBlock(got, `variable "spec"`)
	if strings.Contains(spec, "select_all") {
		t.Errorf("spec.select_all carries manifest_only and must not be declared:\n%s", spec)
	}
	if !strings.Contains(spec, "create_namespace = optional(bool, false)") {
		t.Errorf("spec.create_namespace has no marker and must be declared:\n%s", spec)
	}
}

func TestProtoToVariablesTF_CronJob_ApiVersionKindStatusSkipped(t *testing.T) {
	msg := &kubernetescronjobv1alpha1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	if strings.Contains(got, `"api_version"`) {
		t.Error("api_version should be skipped")
	}
	if strings.Contains(got, `"kind"`) {
		t.Error("kind should be skipped")
	}
	if strings.Contains(got, `"status"`) {
		t.Error("status should be skipped")
	}
}

func TestProtoToVariablesTF_CronJob_HasMetadataAndSpec(t *testing.T) {
	msg := &kubernetescronjobv1alpha1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	if !strings.Contains(got, `variable "metadata"`) {
		t.Error("should have metadata variable")
	}
	if !strings.Contains(got, `variable "spec"`) {
		t.Error("should have spec variable")
	}
}

func TestProtoToVariablesTF_CronJob_ValidHCL(t *testing.T) {
	msg := &kubernetescronjobv1alpha1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(got), "variables.tf")
	if diags.HasErrors() {
		t.Fatalf("generated variables.tf is not valid HCL: %s\n%s", diags.Error(), got)
	}

	blockSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variable", LabelNames: []string{"name"}},
		},
	}
	content, diags := file.Body.Content(blockSchema)
	if diags.HasErrors() {
		t.Fatalf("failed to get content: %s", diags.Error())
	}

	// Should have exactly 2 variable blocks: metadata and spec
	if len(content.Blocks) != 2 {
		t.Errorf("expected 2 variable blocks, got %d", len(content.Blocks))
	}
}

func TestProtoToVariablesTF_SimpleMessage_BackwardCompatible(t *testing.T) {
	// Test with CloudResourceMetadata directly to verify basic object generation.
	msg := &shared.CloudResourceMetadata{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	// Should have name and labels fields (version skipped in metadata).
	if !strings.Contains(got, `variable "name"`) {
		t.Errorf("should have name variable, got:\n%s", got)
	}
	if !strings.Contains(got, `variable "labels"`) {
		t.Errorf("should have labels variable, got:\n%s", got)
	}

	// version should be skipped (metadata convention).
	if strings.Contains(got, `"version"`) {
		t.Errorf("version should be skipped in metadata messages, got:\n%s", got)
	}
}

// --- Optional()/required schema coverage (the production schema-skew fix) ---

// canonicalMetadataBlock is the exact metadata variable every module must carry:
// name is required (always rendered), the rest are optional with proto zero
// defaults so a null-pruned tfvars validates. This is the contract the runtime
// renderer (ProtoToTFVars, EmitUnpopulated=false) depends on.
const canonicalMetadataBlock = `variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}`

func generateVariables(t *testing.T, msg proto.Message) string {
	t.Helper()
	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}
	parser := hclparse.NewParser()
	if _, diags := parser.ParseHCL([]byte(got), "variables.tf"); diags.HasErrors() {
		t.Fatalf("generated variables.tf is not valid HCL: %s\n%s", diags.Error(), got)
	}
	return got
}

// TestProtoToVariablesTF_CanonicalMetadata asserts the shared metadata envelope
// is emitted from the canonical block (name required, rest optional) and never
// leaks orchestrator-only envelope fields (slug/group/relationships) or version.
func TestProtoToVariablesTF_CanonicalMetadata(t *testing.T) {
	got := generateVariables(t, &awssubnetv1alpha1.AwsSubnet{})

	if !strings.Contains(got, canonicalMetadataBlock) {
		t.Errorf("metadata block is not the canonical form.\nwant:\n%s\n\ngot:\n%s", canonicalMetadataBlock, got)
	}
	metaBlock := extractBlock(got, `variable "metadata"`)
	for _, leaked := range []string{"slug", "group", "relationships", "version"} {
		if strings.Contains(metaBlock, leaked) {
			t.Errorf("metadata block leaked envelope/orchestrator field %q:\n%s", leaked, got)
		}
	}
}

// TestProtoToVariablesTF_RequiredVsOptional asserts the required-detection rule:
// buf.validate required OR a presence-implying constraint (string min_len) keeps
// an attribute bare; everything else is optional() with a zero default.
func TestProtoToVariablesTF_RequiredVsOptional(t *testing.T) {
	spec := extractBlock(generateVariables(t, &awssubnetv1alpha1.AwsSubnet{}), `variable "spec"`)

	// region (string.min_len) and vpc_id (required) stay bare. Assert only
	// fields whose requiredness is structural to the kind (a subnet cannot
	// exist without a VPC; region is the universal min_len field) — pinning
	// a live spec's OTHER fields here rots when depth work legally relaxes
	// their requiredness (cidr_block and availability_zone became optional
	// when the spec adopted the provider's conflict graph: IPAM-allocated
	// and IPv6-native subnets carry no IPv4 CIDR).
	for _, bare := range []string{"region = string", "vpc_id = string"} {
		if !strings.Contains(spec, bare) {
			t.Errorf("expected required field rendered bare %q in spec:\n%s", bare, spec)
		}
	}
	// optional scalars carry their proto zero default.
	for _, opt := range []string{
		"map_public_ip_on_launch = optional(bool, false)",
		"ipv6_cidr_block = optional(string, \"\")",
		"cidr_block = optional(string, \"\")",
	} {
		if !strings.Contains(spec, opt) {
			t.Errorf("expected optional field %q in spec:\n%s", opt, spec)
		}
	}
	// optional repeated message -> optional(list(object({...})), [])
	if !strings.Contains(spec, "routes = optional(list(object({") || !strings.Contains(spec, "})), [])") {
		t.Errorf("expected routes as optional list-of-object with [] default:\n%s", spec)
	}
}

// TestProtoToVariablesTF_OptionalScalarsAndMaps covers the kinds that failed in
// production (route53 records, ecr force_delete): all must be optional so a
// pruned tfvars validates instead of erroring "attribute X is required".
func TestProtoToVariablesTF_OptionalScalarsAndMaps(t *testing.T) {
	r53spec := extractBlock(generateVariables(t, &awsroute53zonev1alpha1.AwsRoute53Zone{}), `variable "spec"`)
	if !strings.Contains(r53spec, "vpc_associations = optional(list(object({") {
		t.Errorf("route53 spec.vpc_associations must be optional list-of-object:\n%s", r53spec)
	}

	ecrspec := extractBlock(generateVariables(t, &awsecrrepov1alpha1.AwsEcrRepo{}), `variable "spec"`)
	if !strings.Contains(ecrspec, "force_delete = optional(bool, false)") {
		t.Errorf("ecr spec.force_delete must be optional(bool, false):\n%s", ecrspec)
	}
}

// TestProtoToVariablesTF_RecursiveMessageCollapsesToAny asserts the descriptor
// walk terminates on a self-recursive message and collapses the recursive
// subtree to `any` (the google.protobuf.Struct treatment). Terraform's type
// language cannot express a recursive object constraint, and naturally
// recursive protos (boolean statement trees, nested expression grammars) are a
// legitimate spec shape -- before cycle detection this walk never returned.
// descriptorpb.DescriptorProto is used as the fixture because it is recursive
// by definition (nested_type is a repeated DescriptorProto).
func TestProtoToVariablesTF_RecursiveMessageCollapsesToAny(t *testing.T) {
	got := generateVariables(t, &descriptorpb.DescriptorProto{})

	// The direct self-recursion (nested_type, a repeated DescriptorProto at
	// the root) must collapse to bare `any` instead of recursing forever.
	// Not list(any): Terraform requires all list(any) elements to converge
	// to one type, which recursive elements cannot guarantee.
	nested := extractBlock(got, `variable "nested_type"`)
	if !strings.Contains(nested, "type = any") {
		t.Errorf("recursive repeated field must render bare any, got:\n%s", got)
	}
	// Non-recursive siblings keep their full typed shape (the cycle guard is
	// path-scoped, not a blanket flattening).
	field := extractBlock(got, `variable "field"`)
	if !strings.Contains(field, "type_name = optional(string)") {
		t.Errorf("non-recursive sibling fields must stay fully typed, got:\n%s", got)
	}
}

// extractBlock returns the substring from the first occurrence of header to the
// matching top-level closing brace (best-effort, for assertions/diagnostics).
func extractBlock(s, header string) string {
	i := strings.Index(s, header)
	if i < 0 {
		return ""
	}
	rest := s[i:]
	depth := 0
	for j := 0; j < len(rest); j++ {
		switch rest[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return rest[:j+1]
			}
		}
	}
	return rest
}
