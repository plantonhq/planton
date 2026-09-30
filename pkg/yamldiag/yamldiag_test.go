package yamldiag

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/explain"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func kindDescriptor(t *testing.T, name string) protoreflect.MessageDescriptor {
	t.Helper()
	res, err := explain.ResolveKindName(name)
	if err != nil {
		t.Fatalf("resolve %s: %v", name, err)
	}
	return res.Message
}

func one(t *testing.T, mismatches []Mismatch) Mismatch {
	t.Helper()
	if len(mismatches) != 1 {
		t.Fatalf("want exactly 1 mismatch, got %d: %+v", len(mismatches), mismatches)
	}
	return mismatches[0]
}

func TestUnknownFieldWithSuggestion(t *testing.T) {
	manifest := `apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpc
metadata:
  name: probe
spec:
  region: us-east-1
  cidrBlok: 10.0.0.0/16
`
	m := one(t, Diagnose([]byte(manifest), kindDescriptor(t, "AwsVpc")))
	if m.Path != "spec.cidrBlok" || m.Line != 7 {
		t.Errorf("path/line = %q/%d", m.Path, m.Line)
	}
	if m.Suggestion != "cidrBlock" {
		t.Errorf("suggestion = %q", m.Suggestion)
	}
	// The reference pointer must target the PARENT: the unknown field itself
	// cannot be explained, its legitimate siblings can.
	if rendered := m.Format("AwsVpc"); !strings.Contains(rendered, "see: planton explain AwsVpc.spec\n") &&
		!strings.HasSuffix(rendered, "see: planton explain AwsVpc.spec") {
		t.Errorf("unknown-field pointer should target the parent:\n%s", rendered)
	}
}

func TestUnknownFieldCasingVariant(t *testing.T) {
	manifest := `kind: AwsVpc
spec:
  cidr_blok: 10.0.0.0/16
`
	m := one(t, Diagnose([]byte(manifest), kindDescriptor(t, "AwsVpc")))
	if m.Suggestion != "cidrBlock" {
		t.Errorf("suggestion = %q", m.Suggestion)
	}
}

// TestBareStringForForeignKeyWrapper is the probe's exact failure class: the
// diagnosis must name the path, the REAL line, and both authorable shapes
// with the field's declared reference target.
func TestBareStringForForeignKeyWrapper(t *testing.T) {
	manifest := `apiVersion: aws.planton.dev/v1alpha1
kind: AwsSecurityGroup
metadata:
  name: probe
spec:
  region: us-east-1
  vpcId: vpc-12345
  description: probe group
`
	m := one(t, Diagnose([]byte(manifest), kindDescriptor(t, "aws-security-group")))
	if m.Path != "spec.vpcId" || m.Line != 7 {
		t.Fatalf("path/line = %q/%d", m.Path, m.Line)
	}
	for _, want := range []string{"write as {value:", "valueFrom:", "kind: AwsVpc", "status.outputs.vpc_id", "bare"} {
		if !strings.Contains(m.Problem, want) {
			t.Errorf("problem missing %q: %s", want, m.Problem)
		}
	}
	rendered := m.Format("AwsSecurityGroup")
	if !strings.Contains(rendered, "planton explain AwsSecurityGroup.spec.vpcId") {
		t.Errorf("explain pointer missing: %s", rendered)
	}
}

func TestObjectWhereListExpected(t *testing.T) {
	manifest := `kind: AwsVpc
spec:
  secondaryIpv4Cidrs:
    cidrBlock: 10.1.0.0/16
`
	m := one(t, Diagnose([]byte(manifest), kindDescriptor(t, "AwsVpc")))
	if m.Path != "spec.secondaryIpv4Cidrs" {
		t.Errorf("path = %q", m.Path)
	}
	if !strings.Contains(m.Problem, "expects a list") || !strings.Contains(m.Problem, "got an object") {
		t.Errorf("problem = %q", m.Problem)
	}
}

func TestQuotingErrorOnStringField(t *testing.T) {
	manifest := `kind: AwsVpc
spec:
  region: 1.29
`
	m := one(t, Diagnose([]byte(manifest), kindDescriptor(t, "AwsVpc")))
	if !strings.Contains(m.Problem, "unquoted float") || !strings.Contains(m.Problem, `quote it: "1.29"`) {
		t.Errorf("problem = %q", m.Problem)
	}

	// The inverse must NEVER fire: numeric fields legally accept strings.
	numeric := `kind: AwsVpc
spec:
  ipv4NetmaskLength: "24"
`
	if got := Diagnose([]byte(numeric), kindDescriptor(t, "AwsVpc")); len(got) != 0 {
		t.Errorf("quoted number on int field must not be diagnosed: %+v", got)
	}
}

func TestUnknownEnumValue(t *testing.T) {
	manifest := `kind: KubernetesMetricsServer
spec:
  tls:
    type: cert_manger
`
	m := one(t, Diagnose([]byte(manifest), kindDescriptor(t, "KubernetesMetricsServer")))
	if !strings.Contains(m.Problem, `"cert_manger" is not a value`) {
		t.Errorf("problem = %q", m.Problem)
	}
	if m.Suggestion != "cert_manager" {
		t.Errorf("suggestion = %q", m.Suggestion)
	}
}

func TestOneofDoubleSet(t *testing.T) {
	manifest := `kind: KubernetesNamespace
spec:
  name: probe
  resourceProfile:
    preset: small
    custom:
      cpu:
        requests: "1"
        limits: "2"
`
	found := false
	for _, m := range Diagnose([]byte(manifest), kindDescriptor(t, "kubernetes-namespace")) {
		if strings.Contains(m.Problem, "alternatives") && strings.Contains(m.Problem, "set exactly one") {
			found = true
		}
	}
	if !found {
		t.Error("double-set oneof not diagnosed")
	}
}

// TestAllMismatchesInOnePass locks the agent-loop property: a manifest with
// several independent defects reports them all at once.
func TestAllMismatchesInOnePass(t *testing.T) {
	manifest := `kind: AwsVpc
spec:
  region: 1.29
  cidrBlok: 10.0.0.0/16
  secondaryIpv4Cidrs:
    nested: wrong
`
	got := Diagnose([]byte(manifest), kindDescriptor(t, "AwsVpc"))
	if len(got) != 3 {
		t.Fatalf("want 3 mismatches in one pass, got %d: %+v", len(got), got)
	}
}

func TestValidManifestNoMismatches(t *testing.T) {
	manifest := `apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpc
metadata:
  name: probe
  labels:
    team: platform
spec:
  region: us-east-1
  cidrBlock: 10.0.0.0/16
  enableDnsSupport: true
  secondaryIpv4Cidrs:
    - cidrBlock: 10.1.0.0/16
`
	if got := Diagnose([]byte(manifest), kindDescriptor(t, "AwsVpc")); len(got) != 0 {
		t.Errorf("valid manifest diagnosed: %+v", got)
	}
}

// TestUnmodeledConstructsSkipInTheSafeDirection: aliased values and
// merge-key expansions are not walked -- their content is checked where the
// anchor is defined -- so they can never produce a false diagnosis, only
// under-reporting that the original parser error still covers.
func TestUnmodeledConstructsSkipInTheSafeDirection(t *testing.T) {
	t.Run("valid alias reuse diagnoses nothing", func(t *testing.T) {
		manifest := `kind: KubernetesNamespace
metadata:
  labels: &common
    team: platform
spec:
  name: probe
  annotations: *common
`
		if got := Diagnose([]byte(manifest), kindDescriptor(t, "kubernetes-namespace")); len(got) != 0 {
			t.Errorf("aliased manifest wrongly diagnosed: %+v", got)
		}
	})

	t.Run("merge keys are skipped, explicit keys still checked", func(t *testing.T) {
		// The merged content would fail protojson (team is not a field of
		// networkConfig) -- the diagnoser cannot see through the merge, so
		// it must stay silent about it while still checking explicit keys.
		manifest := `kind: KubernetesNamespace
metadata:
  labels: &base
    team: platform
spec:
  name: probe
  networkConfig:
    <<: *base
    isolateIngres: true
`
		m := one(t, Diagnose([]byte(manifest), kindDescriptor(t, "kubernetes-namespace")))
		if m.Suggestion != "isolateIngress" {
			t.Errorf("explicit key beside a merge not checked: %+v", m)
		}
	})
}

// TestSilentOnProtojsonOwnedFailures: value-format problems inside legal
// structure belong to the parser; the diagnoser must not guess.
func TestSilentOnProtojsonOwnedFailures(t *testing.T) {
	// ipv4NetmaskLength as a non-numeric string is structurally a legal
	// scalar; whether "abc" coerces is protojson's call.
	manifest := `kind: AwsVpc
spec:
  region: us-east-1
  ipv4NetmaskLength: abc
`
	if got := Diagnose([]byte(manifest), kindDescriptor(t, "AwsVpc")); len(got) != 0 {
		t.Errorf("scalar coercion wrongly diagnosed: %+v", got)
	}
}

func TestFormatAll(t *testing.T) {
	text := FormatAll([]Mismatch{
		{Path: "spec.a", Line: 3, Problem: "p1"},
		{Path: "spec.b", Line: 5, Problem: "p2", Suggestion: "c"},
	}, "AwsVpc")
	for _, want := range []string{"spec.a (line 3): p1", "did you mean: c", "planton explain AwsVpc.spec.b"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
}
