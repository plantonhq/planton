package runner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/plantonhq/planton/e2e/framework/provider"
)

// The scenario-declared skip annotation exists for configs that
// STRUCTURALLY cannot round-trip to zero changes (write-only ForceNew
// secrets, where the only post-import plan is a replace): the gate must
// exclude a scenario ONLY when the annotation carries a reason, and an
// annotation-free scenario must stay enrolled.
func TestImportRoundTripEnabled_ScenarioSkipAnnotation(t *testing.T) {
	repoRoot := t.TempDir()
	mapDir := filepath.Join(repoRoot, "catalog", "aws", "awsecscluster", "iac")
	if err := os.MkdirAll(mapDir, 0o755); err != nil {
		t.Fatalf("mkdir import-map dir: %v", err)
	}
	mapYaml := "apiVersion: iac.planton.dev/v1\nkind: CatalogKindImportMap\n"
	if err := os.WriteFile(filepath.Join(mapDir, "import-map.yaml"), []byte(mapYaml), 0o600); err != nil {
		t.Fatalf("write import map: %v", err)
	}
	t.Setenv(ImportRoundTripEnvVar, "1")

	tc := &provider.KindTestContext{
		Kind:     "awsecscluster",
		Provider: "aws",
		Engine:   "terraform",
		RepoRoot: repoRoot,
	}

	tc.ManifestPath = writeScenarioManifest(t, t.TempDir(), "")
	if !importRoundTripEnabled(tc) {
		t.Fatal("annotation-free scenario must stay enrolled in the round-trip")
	}

	tc.ManifestPath = writeScenarioManifest(t, t.TempDir(),
		"  annotations:\n    planton.dev/e2e-import-roundtrip-skip: \"write-only ForceNew secrets cannot round-trip\"\n")
	if importRoundTripEnabled(tc) {
		t.Fatal("a reason-carrying skip annotation must exclude the scenario")
	}
}

// The round-trip oracle's tolerance hinges on this differ naming EXACTLY the
// attributes an in-place update touches: name too few and a real drift slips
// through as "tolerated"; degrade on unexpected shapes and the check must
// fail closed, never tolerate blindly.
func TestChangedTopLevelAttributes(t *testing.T) {
	cases := []struct {
		name   string
		before interface{}
		after  interface{}
		want   []string
	}{
		{
			name:   "single config-only flip (the post-import force_destroy case)",
			before: map[string]interface{}{"force_destroy": false, "bucket": "b", "tags": map[string]interface{}{"a": "1"}},
			after:  map[string]interface{}{"force_destroy": true, "bucket": "b", "tags": map[string]interface{}{"a": "1"}},
			want:   []string{"force_destroy"},
		},
		{
			name:   "nested change surfaces as its top-level attribute",
			before: map[string]interface{}{"tags": map[string]interface{}{"a": "1"}},
			after:  map[string]interface{}{"tags": map[string]interface{}{"a": "2"}},
			want:   []string{"tags"},
		},
		{
			name:   "attribute present on one side only",
			before: map[string]interface{}{"bucket": "b"},
			after:  map[string]interface{}{"bucket": "b", "acl": "private"},
			want:   []string{"acl"},
		},
		{
			name:   "no change",
			before: map[string]interface{}{"bucket": "b"},
			after:  map[string]interface{}{"bucket": "b"},
			want:   nil,
		},
		{
			name:   "non-object shapes fail closed",
			before: nil,
			after:  map[string]interface{}{"bucket": "b"},
			want:   []string{"<non-object change>"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := changedTopLevelAttributes(tc.before, tc.after)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("changedTopLevelAttributes = %v, want %v", got, tc.want)
			}
		})
	}
}

// The sub-path tolerance backs BOTH declaration vocabularies (the provider
// catalog's dotted config-only/write-normalized entries and the kind
// map's resource-scoped import-normalized entries): only the declared leaf
// may differ; any sibling drift under the same attribute fails closed. The
// canonical import-normalized case is a Secret data key wiring a salted-hash
// computed attribute the provider re-salts on import.
func TestChangeCoveredBySubPaths(t *testing.T) {
	before := map[string]interface{}{"data": map[string]interface{}{
		"REGISTRY_PASSWD":   "plain",
		"REGISTRY_HTPASSWD": "user:$2a$10$oldsalt",
	}}
	afterHashOnly := map[string]interface{}{"data": map[string]interface{}{
		"REGISTRY_PASSWD":   "plain",
		"REGISTRY_HTPASSWD": "user:$2a$10$newsalt",
	}}
	afterSiblingDrift := map[string]interface{}{"data": map[string]interface{}{
		"REGISTRY_PASSWD":   "DRIFTED",
		"REGISTRY_HTPASSWD": "user:$2a$10$newsalt",
	}}
	declared := [][]string{{"data", "REGISTRY_HTPASSWD"}}

	if !changeCoveredBySubPaths(before, afterHashOnly, "data", declared) {
		t.Fatal("the declared salted-hash sub-path alone must be tolerated")
	}
	if changeCoveredBySubPaths(before, afterSiblingDrift, "data", declared) {
		t.Fatal("sibling drift under the same attribute must fail closed")
	}
	if changeCoveredBySubPaths(before, afterHashOnly, "data", nil) {
		t.Fatal("no declarations must tolerate nothing")
	}
	if changeCoveredBySubPaths(before, afterHashOnly, "metadata", declared) {
		t.Fatal("a declaration scoped to another attribute must not cover this one")
	}
}

// The blind round-trip holds no user-pasted ARN, but account_id/region are
// deployment-level facts every ARN-shaped output carries; per-resource parts
// must NEVER be synthesized (a wrong recipe could then pass on a fake id).
func TestAccountLevelArnParts(t *testing.T) {
	cases := []struct {
		name    string
		outputs map[string]string
		want    map[string]string
	}{
		{
			name: "regional ARN carries both parts",
			outputs: map[string]string{
				"table_arn": "arn:aws:dynamodb:us-west-2:123456789012:table/tbl",
			},
			want: map[string]string{"account_id": "123456789012", "region": "us-west-2"},
		},
		{
			name: "global-service ARNs fill what they carry (IAM: account only)",
			outputs: map[string]string{
				"role_arn": "arn:aws:iam::123456789012:role/my-role",
			},
			want: map[string]string{"account_id": "123456789012"},
		},
		{
			name: "parts merge across outputs (S3 carries neither, IAM the account, regional the region)",
			outputs: map[string]string{
				"bucket_arn": "arn:aws:s3:::my-bucket",
				"role_arn":   "arn:aws:iam::123456789012:role/my-role",
			},
			want: map[string]string{"account_id": "123456789012"},
		},
		{
			name:    "no ARN-shaped outputs yield nil, never fabricated parts",
			outputs: map[string]string{"bucket_id": "my-bucket"},
			want:    nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := accountLevelArnParts(tc.outputs)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("accountLevelArnParts = %v, want %v", got, tc.want)
			}
		})
	}
}

// Optional ("{name?}") placeholders never block a blind import; required ones
// always do.
func TestRequiredUnresolved(t *testing.T) {
	missing := requiredUnresolved(
		"name:{table_name}/index:{index_name?}/{account_id}",
		[]string{"index_name", "account_id"},
	)
	if !reflect.DeepEqual(missing, []string{"account_id"}) {
		t.Fatalf("requiredUnresolved = %v, want [account_id]", missing)
	}
}
