package vocabulary

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func scanner(t *testing.T) *Scanner {
	t.Helper()
	v, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	s, err := v.NewScanner()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func matches(t *testing.T, s *Scanner, path, line string) []string {
	t.Helper()
	var got []string
	for _, f := range s.ScanText(path, []byte(line)) {
		got = append(got, f.Match)
	}
	return got
}

func TestSpellingPatternMatchesEveryCaseForm(t *testing.T) {
	re := regexp.MustCompile(SpellingPattern("CloudResource"))
	for _, form := range []string{
		"CloudResource", "cloudResource", "cloud_resource", "CLOUD_RESOURCE",
		"cloud-resource", "cloudresource", "Cloud Resource", "cloud resource",
		"AwsCloudResourceKind", "dev.planton.shared.cloudresourcekind",
	} {
		if !re.MatchString(form) {
			t.Errorf("%q not matched by %s", form, re)
		}
	}
	if got := SpellingPattern("crkreflect"); got != "(?i)crkreflect" {
		t.Errorf("one-word spelling: got %s", got)
	}
}

func TestRetiredSpellingsAreFound(t *testing.T) {
	s := scanner(t)
	for _, line := range []string{
		"kind: CloudResourceKind",
		"message AwsS3BucketStackInput {",
		"stackInput := &AwsS3BucketStackInput{}",
		"value: stack-job",
		"TEMPORAL_TASK_QUEUE_STACK_JOB",
		"an infra project in the environment",
		"last_applied_cloud_object",
		"import github.com/plantonhq/planton/pkg/crkreflect",
		"kind: ComponentCostProfile",
		"labels: {e2e-component: awsvpc}",
		"see _rules/component/forge",
		"a catalog of 700+ components",
		"id: cr_awsvpc_01jabcdefghjkmnpqrstvwxyz0",
		"id: sj_01jabcdefghjkmnpqrstvwxyz0",
		"IDs look like `cr_`",
		`strings.HasPrefix(id, "cr_")`,
		"run planton cloud-resource:apply",
		"Stack Outputs",
		"projectId: infpj_01ABC…",
		"InfraStack (infproj_…)",
		"a working copy of a deployed project",
		"cat .planton/project.yaml",
		"planton project create --from-chart gcp/cloud-run-environment",
		"- name: STACK_EXECUTION_LOGS_GCS_BUCKET",
	} {
		if len(matches(t, s, "README.md", line)) == 0 {
			t.Errorf("no finding in %q", line)
		}
	}
}

func TestWrappedNamesAreFound(t *testing.T) {
	s := scanner(t)
	for _, text := range []string{
		"Pulumi implementation for the AwsVpc deployment\ncomponent.",
		"// the value a\n// cloud resource carries",
		"# then a\n# catalog\n# component",
	} {
		if len(s.ScanText("README.md", []byte(text))) == 0 {
			t.Errorf("no finding in wrapped %q", text)
		}
	}
	if got := s.ScanText("README.md", []byte("the Google\nCloud resource hierarchy")); len(got) != 0 {
		t.Errorf("a vendor phrase wrapped across lines was flagged: %v", got)
	}
	if got := s.ScanText("catalog/x/e2e/manifest.yaml", []byte("    release: kube-prometheus-stack\n  job_label: cnpg.io/cluster")); len(got) != 0 {
		t.Errorf("a YAML key line was joined to the value above it: %v", got)
	}
	if got := s.ScanText("README.md", []byte("one deployment\nper environment")); len(got) != 0 {
		t.Errorf("ordinary wrapped prose was flagged: %v", got)
	}
}

func TestOtherPeoplesWordsAreAllowed(t *testing.T) {
	s := scanner(t)
	for _, c := range []struct{ path, line string }{
		{"catalog/gcp/gcpproject/iac/tf/main.tf", `service = "cloudresourcemanager.googleapis.com"`},
		{"docs/gcp.md", "Google Cloud Resource Manager"},
		{"catalog/gcp/gcpbigqueryconnection/v1alpha1/spec.proto", "GcpBigQueryConnectionCloudResource cloud_resource = 5;"},
		{"pkg/x.go", "service account of a google_cloud_resource"},
		{"catalog/gcp/gcpbigqueryconnection/iac/pulumi/module/connection.go", "args.CloudResource = &bigquery.ConnectionCloudResourceArgs{}"},
		{"catalog/kubernetes/kubernetespostgres/iac/x.go", "BarmanCloudObjectStore"},
		{"docs/x.md", "the multi-cloud-catalog skill"},
		{"docs/x.md", "every Google Cloud resource in the project"},
		{"docs/x.md", "the InfraStack Input tab and every InfraStack job"},
		{"pkg/x.go", "est := InfraComponentCostEstimate{}"},
		{"docs/x.md", "a contact's OpenStack job at ARM"},
		{"_changelog/x.md", "added to the Hetzner Cloud catalog"},
		{"docs/x.md", "run `pulumi stack output --json`"},
		{"catalog/kubernetes/kuberneteskubeprometheusstack/v1alpha1/outputs.proto", "message KubernetesKubePrometheusStackOutputs {"},
		{"docs/x.md", "a Kubernetes Deployment of three replicas"},
		{"site/src/components/Hero.tsx", "import { Hero } from '@/components/Hero'"},
		{"catalog/aws/awsvpc/cost.yaml", "# what drives this component's bill"},
		{"pkg/x.go", "cr_terms := local.terms"},
		{"catalog/gcp/gcpproject/v1/spec.proto", "string project_id = 1; // projectId of the deployed GCP project's parent"},
		{"docs/x.md", "Pulumi stack execution is orchestrated by the runner"},
		{"docs/x.md", "planton-os project create --name q4-launch"},
		{"docs/x.md", "the Dockerfile in the service's project root"},
	} {
		if got := matches(t, s, c.path, c.line); len(got) != 0 {
			t.Errorf("%s: %q flagged %v", c.path, c.line, got)
		}
	}
}

func TestPathAllowancesStayInTheirPaths(t *testing.T) {
	s := scanner(t)
	if got := matches(t, s, "pkg/x.go", "cloud_resource := spec.CloudResource"); len(got) == 0 {
		t.Error("BigQuery's allowance leaked outside gcpbigqueryconnection")
	}
}

func TestGlobDoubleStarSlashMatchesNoDirectory(t *testing.T) {
	re := globRegexp("**/go.mod")
	for _, p := range []string{"go.mod", "a/go.mod", "a/b/go.mod"} {
		if !re.MatchString(p) {
			t.Errorf("**/go.mod should match %s", p)
		}
	}
	if re.MatchString("go.mod.bak") || globRegexp("site/*.ts").MatchString("site/a/x.ts") {
		t.Error("glob matched too much")
	}
}

func TestExcludedPaths(t *testing.T) {
	s := scanner(t)
	for _, p := range []string{"pkg/vocabulary/vocabulary.yaml", "pkg/kubernetes/kubernetestypes/x/y.go"} {
		if !s.Excluded(p) {
			t.Errorf("%s should be excluded", p)
		}
	}
	if s.Excluded("pkg/vocabularyx/a.go") || s.Excluded("catalog/aws/awsvpc/README.md") {
		t.Error("exclusion matched a path it should not")
	}
}

// TestCurrentWordsContainNoRetiredSpelling keeps re-translation safe: a
// current word that contained a retired spelling would be flagged forever.
func TestCurrentWordsContainNoRetiredSpelling(t *testing.T) {
	s := scanner(t)
	v, _ := Load()
	for _, w := range v.Words {
		if got := matches(t, s, "README.md", w.Name); len(got) != 0 {
			t.Errorf("current word %s contains retired spelling %v", w.Name, got)
		}
	}
}

// repoRoot resolves the repository root from this file's location so the
// gate works from any test working directory (including the Bazel sandbox,
// where the source tree is absent -- the gate skips there).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller location")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestRetiredVocabularyGate is the CI guardrail: no tracked file carries a
// spelling vocabulary.yaml retires, outside its allowances. On failure, write
// the word the finding names; never add an allowance for a Planton word.
func TestRetiredVocabularyGate(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		t.Skip("repository tree not present (bazel sandbox); runs under go test and the lint.vocabulary lane")
	}
	findings, err := scanner(t).ScanTree(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		return
	}
	counts := CountByUse(findings)
	uses := make([]string, 0, len(counts))
	for u := range counts {
		uses = append(uses, u)
	}
	sort.Slice(uses, func(i, j int) bool { return counts[uses[i]] > counts[uses[j]] })
	var b strings.Builder
	fmt.Fprintf(&b, "%d retired spellings in tracked files:\n", len(findings))
	for _, u := range uses {
		fmt.Fprintf(&b, "  %7d  use %s\n", counts[u], u)
	}
	const shown = 50
	for i, f := range findings {
		if i == shown {
			fmt.Fprintf(&b, "  ... and %d more\n", len(findings)-shown)
			break
		}
		fmt.Fprintf(&b, "  %s\n", f)
	}
	t.Fatal(b.String())
}
