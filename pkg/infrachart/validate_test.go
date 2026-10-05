package infrachart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
)

// The end-to-end tests exercise the full gate against the permanent _test
// kind (TestCatalogKindGeneric), whose annotated_ref field carries the full
// FK annotation pair (default_kind = TestCatalogKindGeneric,
// default_kind_field_path = "status.outputs.id") — a hermetic fixture that
// never moves with production resource shapes. The two testdata fixture
// charts additionally exercise the pipeline against real provider kinds.

const testChartYaml = `apiVersion: infra-hub.planton.ai/v1alpha1
kind: InfraChart
metadata:
  name: Test Chart
spec:
  selector:
    kind: platform
  description: hermetic test chart
`

const testValuesYaml = `params:
  - name: base_name
    description: base resource name
    value: node
`

func writeChart(t *testing.T, templates map[string]string) string {
	t.Helper()
	return writeChartWithValues(t, testValuesYaml, templates)
}

func writeChartWithValues(t *testing.T, valuesYaml string, templates map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Chart.yaml"), testChartYaml)
	mustWrite(t, filepath.Join(dir, "values.yaml"), valuesYaml)
	for name, content := range templates {
		mustWrite(t, filepath.Join(dir, "templates", name), content)
	}
	return dir
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustValidate(t *testing.T, dir string, opts Options) *Report {
	t.Helper()
	report, err := Validate(dir, opts)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	return report
}

// issueMessages returns every issue message of the given severity across all
// variants.
func issueMessages(report *Report, severity Severity) []string {
	var out []string
	for _, v := range report.Variants {
		for _, issue := range v.Issues {
			if issue.Severity == severity {
				out = append(out, "["+v.Name+"] "+issue.Message)
			}
		}
	}
	return out
}

// defaultsVariant returns the "defaults" variant result.
func defaultsVariant(t *testing.T, report *Report) *VariantResult {
	t.Helper()
	for i := range report.Variants {
		if report.Variants[i].Name == "defaults" {
			return &report.Variants[i]
		}
	}
	t.Fatal("report has no defaults variant")
	return nil
}

func requireError(t *testing.T, report *Report, substring string) {
	t.Helper()
	for _, msg := range issueMessages(report, SeverityError) {
		if strings.Contains(msg, substring) {
			return
		}
	}
	t.Fatalf("expected an error containing %q, got errors: %v", substring, issueMessages(report, SeverityError))
}

const validNode = `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "{{ env }}-{{ values.base_name }}-b"
spec:
  requiredRef:
    value: literal
`

// A rendered document whose declared apiVersion conflicts with its kind must
// fail the build — the deploy boundary rejects the same conflict, and build
// and deploy must never disagree.
func TestValidateRejectsWrongEnvelopeApiVersion(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v99
kind: TestCatalogKindGeneric
metadata:
  name: "{{ env }}-{{ values.base_name }}-a"
spec:
  requiredRef:
    value: literal
`,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "apiVersion '_test.planton.dev/v99' does not match kind TestCatalogKindGeneric")
}

// A rendered document may omit apiVersion entirely: the platform stamps the
// authoritative value on write, so absence is not an authoring error.
func TestValidateAllowsMissingEnvelopeApiVersion(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
kind: TestCatalogKindGeneric
metadata:
  name: "{{ env }}-{{ values.base_name }}-a"
spec:
  requiredRef:
    value: literal
`,
	})
	report := mustValidate(t, dir, Options{})
	if report.HasErrors() {
		t.Fatalf("missing apiVersion must not be an error: %v", issueMessages(report, SeverityError))
	}
}

func TestValidateHappyPathWithInChartReference(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"b.yaml": validNode,
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "{{ env }}-{{ values.base_name }}-a"
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: "{{ env }}-{{ values.base_name }}-b"
`,
	})
	report := mustValidate(t, dir, Options{})
	if report.HasErrors() {
		t.Fatalf("expected no errors, got: %v", issueMessages(report, SeverityError))
	}
	docs := defaultsVariant(t, report).Docs
	if len(docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(docs))
	}
	if warnings := issueMessages(report, SeverityWarning); len(warnings) != 0 {
		t.Fatalf("expected no warnings (reference resolves in-chart), got: %v", warnings)
	}
	// The annotated default fieldPath was applied — the reference resolved
	// without an explicit fieldPath in the template.
	if docs[0].Name != "dev-node-a" && docs[1].Name != "dev-node-a" {
		t.Fatalf("rendered names unexpected: %v / %v", docs[0].Name, docs[1].Name)
	}
}

func TestValidateFkOverrideIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"b.yaml": validNode,
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: "dev-node-b"
      fieldPath: status.outputs.name
`,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "overrides the annotated composition key")
}

// A reference to a resource NO variant defines is a warning, not an error:
// charts compose onto resources owned elsewhere by design (a landing zone's
// network, a pre-existing DNS zone), resolved in the target environment at
// deploy time.
func TestValidateCrossChartReferenceIsAWarning(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: "ghost"
`,
	})
	report := mustValidate(t, dir, Options{})
	if report.HasErrors() {
		t.Fatalf("cross-chart reference must not be an error: %v", issueMessages(report, SeverityError))
	}
	warnings := issueMessages(report, SeverityWarning)
	if len(warnings) != 1 || !strings.Contains(warnings[0], "does not define") {
		t.Fatalf("expected one cross-chart-reference warning, got: %v", warnings)
	}
}

// A reference whose target another variant DOES define gets a sharper
// warning in the variant that lost it: either the toggle removed the
// resource while the reference still stands (a defect), or this is a
// bring-your-own arm redirecting the same reference to a resource owned
// outside the chart (a designed pattern). The two are offline-
// indistinguishable, so the diagnosis names both.
func TestValidateToggleRemovedTargetWarnsWithDiagnosis(t *testing.T) {
	values := testValuesYaml + `  - name: nodeBEnabled
    description: deploy node b
    type: bool
    value: true
`
	dir := writeChartWithValues(t, values, map[string]string{
		"b.yaml": `{% if values.nodeBEnabled | bool %}
` + validNode + `{% endif %}
`,
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: "dev-node-b"
`,
	})
	report := mustValidate(t, dir, Options{})
	if report.HasErrors() {
		t.Fatalf("variant-lost target must not be an error (bring-your-own arms are designed): %v", issueMessages(report, SeverityError))
	}
	// The defaults variant (toggle on) is clean; the flipped variant must
	// carry the sharper defined-in-another-variant diagnosis.
	found := false
	for _, v := range report.Variants {
		for _, issue := range v.Issues {
			if issue.Severity == SeverityWarning && strings.Contains(issue.Message, "defines in another variant") {
				if v.Name != "nodeBEnabled=false" {
					t.Fatalf("variant-lost-target warning attributed to variant %q", v.Name)
				}
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("expected a variant-lost-target warning, got: %+v", report.Variants)
	}
}

func TestValidateDependencyCycleIsAnError(t *testing.T) {
	node := func(name, refName string) string {
		return `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "` + name + `"
spec:
  requiredRef:
    value: literal
  annotatedRef:
    valueFrom:
      name: "` + refName + `"
`
	}
	dir := writeChart(t, map[string]string{
		"a.yaml": node("a", "b"),
		"b.yaml": node("b", "a"),
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "dependency cycle")
}

func TestValidateUnresolvableFieldPathIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
spec:
  requiredRef:
    value: literal
  optionalRef:
    valueFrom:
      kind: TestCatalogKindGeneric
      name: "a"
      fieldPath: status.outputs.no_such_output
`,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "no field named")
}

func TestValidateReferenceWithoutKindIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
spec:
  requiredRef:
    value: literal
  optionalRef:
    valueFrom:
      name: "something"
`,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "does not name a kind")
}

func TestValidateSpecViolationIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
spec:
  displayName: hello
`,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "schema validation failed")
}

func TestValidateMissingMetadataNameIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  id: not-a-name
spec:
  requiredRef:
    value: literal
`,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "metadata.name is required")
}

func TestValidateUnknownFieldIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
spec:
  requiredRef:
    value: literal
  fieldThatDoesNotExist: true
`,
	})
	report := mustValidate(t, dir, Options{})
	if !report.HasErrors() {
		t.Fatal("unknown spec field must fail strict loading")
	}
}

// Bool params are flipped automatically — one variant per toggle — so both
// arms of every conditional manifest validate without hand-written --set runs.
func TestValidateAutoFlipsBoolToggles(t *testing.T) {
	values := testValuesYaml + `  - name: extraEnabled
    description: deploy the extra node
    type: bool
    value: false
`
	templates := map[string]string{
		"a.yaml": validNode,
		"extra.yaml": `{% if values.extraEnabled | bool %}
---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "extra"
spec:
  requiredRef:
    value: literal
{% endif %}
`,
	}
	dir := writeChartWithValues(t, values, templates)

	report := mustValidate(t, dir, Options{})
	if report.HasErrors() {
		t.Fatalf("errors: %v", issueMessages(report, SeverityError))
	}
	if len(report.Variants) != 2 {
		t.Fatalf("expected 2 variants (defaults + auto-flip), got %d", len(report.Variants))
	}
	if got := len(report.Variants[0].Docs); got != 1 {
		t.Fatalf("defaults (toggle off): expected 1 doc, got %d", got)
	}
	if report.Variants[1].Name != "extraEnabled=true" {
		t.Fatalf("unexpected variant name %q", report.Variants[1].Name)
	}
	if got := len(report.Variants[1].Docs); got != 2 {
		t.Fatalf("flipped (toggle on): expected 2 docs, got %d", got)
	}

	// An explicit --set pins the toggle: it is not auto-flipped again.
	report = mustValidate(t, dir, Options{Set: map[string]string{"extraEnabled": "true"}})
	if report.HasErrors() || len(report.Variants) != 1 {
		t.Fatalf("explicit set: errors=%v variants=%d", issueMessages(report, SeverityError), len(report.Variants))
	}
	if got := len(report.Variants[0].Docs); got != 2 {
		t.Fatalf("explicit toggle-on: expected 2 docs, got %d", got)
	}
}

func TestValidateBannedTagIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": "{% set x = 1 %}\n" + validNode,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "disabled by the platform's sandboxed renderer")
}

func TestValidateDuplicateIdentityIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": validNode,
		"b.yaml": validNode,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "duplicate resource identity")
}

func TestValidateValuelessParamWarnsWhenUsed(t *testing.T) {
	values := `params:
  - name: base_name
    description: has no default
`
	dir := writeChartWithValues(t, values, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "{{ values.base_name }}"
spec:
  requiredRef:
    value: literal
`,
	})
	report := mustValidate(t, dir, Options{})
	found := false
	for _, msg := range issueMessages(report, SeverityWarning) {
		if strings.Contains(msg, "has no default value") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a no-default warning, got: %+v", report.Variants)
	}
}

// The fixture charts exercise the pipeline against real provider kinds.

func TestValidateValidFixture(t *testing.T) {
	report := mustValidate(t, "testdata/valid-chart", Options{})
	if report.HasErrors() {
		t.Fatalf("expected the valid fixture chart to pass, got: %v", issueMessages(report, SeverityError))
	}
	// defaults + the subnetEnabled toggle flipped.
	if len(report.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(report.Variants))
	}
	if got := len(report.Variants[0].Docs); got != 2 {
		t.Fatalf("defaults variant: expected 2 docs (vpc + subnet), got %d", got)
	}
	if report.Variants[1].Name != "subnetEnabled=false" {
		t.Fatalf("unexpected variant name %q", report.Variants[1].Name)
	}
	if got := len(report.Variants[1].Docs); got != 1 {
		t.Fatalf("flipped variant: expected 1 doc (vpc only), got %d", got)
	}
}

func TestValidateBrokenFixture(t *testing.T) {
	report := mustValidate(t, "testdata/broken-chart", Options{})
	if !report.HasErrors() {
		t.Fatal("expected the broken fixture chart to fail")
	}
	if len(report.Variants) != 1 {
		t.Fatalf("expected 1 variant (no bool params), got %d", len(report.Variants))
	}

	// One error per defect class — the deliberately wrong fieldPath trips two
	// checks at once (it overrides the field's annotated composition key AND
	// does not resolve on the target kind), both correct...
	errs := issueMessages(report, SeverityError)
	if len(errs) != 4 {
		t.Fatalf("expected 4 errors, got %d: %v", len(errs), errs)
	}
	assertContains(t, errs[0], "cidrBlok")                            // unknown spec field fails strict load
	assertContains(t, errs[1], "region")                              // required field missing fails protovalidate
	assertContains(t, errs[2], "overrides the annotated composition") // explicit path fights the FK annotation
	assertContains(t, errs[3], "does_not_exist")                      // fieldPath does not resolve on the target kind

	// ...and the undefined targets surface as cross-chart warnings: the
	// phantom-vpc reference is well-formed but targets a resource no variant
	// defines, and the dangling-ref doc's target failed to load so it is not
	// a defined identity either.
	warnings := issueMessages(report, SeverityWarning)
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %v", len(warnings), warnings)
	}
	assertContains(t, warnings[0]+warnings[1], "phantom-vpc")
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("%q does not mention %q", got, want)
	}
}

// The whole document is validated, the way the install does, not only its
// spec: a metadata rule (here the slug alphabet) that the spec-only check
// never ran passed offline and was refused by the server.
func TestValidateMetadataViolationIsAnError(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"a.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "a"
  slug: Not_A_Slug
spec:
  requiredRef:
    value: literal
`,
	})
	report := mustValidate(t, dir, Options{})
	requireError(t, report, "A slug is lowercase letters and digits joined by single hyphens")
}

// The reported case, byte for byte: a lifecycle policy written as a YAML
// block scalar (`|`) at the end of the file keeps its trailing newline on the
// server, and the catalog's rule refuses a policy that does not end with its
// closing brace. The offline render used to drop the file's final newline, so
// the scalar lost its own and the chart passed here.
func TestValidateBlockScalarPolicyFailsExactlyAsInstallDoes(t *testing.T) {
	ecr := func(indicator string) string {
		return `apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcrRegistrySettings
metadata:
  name: planton-e2e-service-path
spec:
  region: us-east-1
  repositoryCreationTemplates:
    - prefix: planton-e2e
      appliedFor:
        - CREATE_ON_PUSH
      lifecyclePolicy: ` + indicator + `
        {"rules": [{"rulePriority": 1, "selection": {"tagStatus": "any", "countType": "sinceImagePushed", "countUnit": "days", "countNumber": 1}, "action": {"type": "expire"}}]}
`
	}

	kept := mustValidate(t, writeChart(t, map[string]string{"ecr.yaml": ecr("|")}), Options{})
	requireError(t, kept, "lifecycle_policy must be a valid JSON lifecycle policy document")

	stripped := mustValidate(t, writeChart(t, map[string]string{"ecr.yaml": ecr("|-")}), Options{})
	if stripped.HasErrors() {
		t.Fatalf("a block-stripped policy is accepted by install and must pass here: %v", issueMessages(stripped, SeverityError))
	}
}

// A host adds its own rules over each rendered document (the Planton CLI adds
// where secrets may go): a host finding is an error on that document, and a
// host check runs only on a document whose schema passed, the order the
// platform applies them in.
func TestValidateRunsHostDocumentChecksAfterTheSchema(t *testing.T) {
	var seen []string
	check := func(doc proto.Message) []string {
		name := metadatareflect.ExtractMetadata(doc).GetName()
		seen = append(seen, name)
		return []string{"host finding on " + name}
	}
	dir := writeChart(t, map[string]string{
		"ok.yaml": validNode,
		"bad.yaml": `---
apiVersion: _test.planton.dev/v1alpha2
kind: TestCatalogKindGeneric
metadata:
  name: "bad"
spec:
  displayName: hello
`,
	})
	report := mustValidate(t, dir, Options{DocumentChecks: []DocumentCheck{check}})
	requireError(t, report, "host finding on dev-node-b")
	for _, name := range seen {
		if name == "bad" {
			t.Fatalf("a host check must not run on a document whose schema failed")
		}
	}
}

// A host that resolves reference tokens before anything deploys (the Planton
// CLI and its `$secret/` references) names them, and a schema rule written
// about the literal -- here the binary data's base64 pattern -- is set aside
// for the token instead of failing the chart. Without the host's grammar the
// same chart fails the rule, because the backendless deploy takes literals.
func TestValidateDefersARuleOnAHostToken(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"secret.yaml": `---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSecret
metadata:
  name: runner-ca
spec:
  name: runner-ca
  namespace:
    value: builds
  opaque:
    binaryData:
      ca.crt: $secret/runner-ca
`,
	})
	isToken := func(value string) bool { return strings.HasPrefix(value, "$secret/") }

	requireError(t, mustValidate(t, dir, Options{}), "schema validation failed")

	report := mustValidate(t, dir, Options{IsDeferredToken: isToken})
	if errs := issueMessages(report, SeverityError); len(errs) > 0 {
		t.Fatalf("a rule on a host token must be set aside, got: %v", errs)
	}
}
