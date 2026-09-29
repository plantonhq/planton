package failure

import (
	"errors"
	"strings"
	"testing"
)

// The raw texts below were recorded from real engine runs (OpenTofu 1.12 with
// the helm, http, and kubectl providers; the Helm SDK; a Kind API server) and
// are kept verbatim, wrapping and box characters included, so the explainer
// is tested against what a terminal shows and not against a tidied copy.

const tofuUnreachableRepository = `
╷
│ Error: Error making request
│ 
│   with data.http.helm_crds_index[0],
│   on helm_crds.tf line 100, in data "http" "helm_crds_index":
│  100: data "http" "helm_crds_index" {
│ 
│ Error making request: GET
│ https://charts.does-not-resolve.invalid/index.yaml giving up after 1
│ attempt(s): Get "https://charts.does-not-resolve.invalid/index.yaml": dial
│ tcp: lookup charts.does-not-resolve.invalid: no such host
╵
╷
│ Error: Error locating chart
│ 
│   with data.helm_template.helm_crds[0],
│   on helm_crds.tf line 127, in data "helm_template" "helm_crds":
│  127: data "helm_template" "helm_crds" {
│ 
│ Unable to locate chart podinfo: looks like
│ "https://charts.does-not-resolve.invalid" is not a valid chart repository
│ or cannot be reached: Get
│ "https://charts.does-not-resolve.invalid/index.yaml": dial tcp: lookup
│ charts.does-not-resolve.invalid: no such host
╵
`

const tofuForbiddenCRDs = `
╷
│ Error: compositecontrollers.metacontroller.k8s.io failed to run apply: customresourcedefinitions.apiextensions.k8s.io "compositecontrollers.metacontroller.k8s.io" is forbidden: User "system:serviceaccount:gate-identity:gate-restricted" cannot patch resource "customresourcedefinitions" in API group "apiextensions.k8s.io" at the cluster scope
│ 
│   with kubectl_manifest.helm_crds["compositecontrollers.metacontroller.k8s.io"],
│   on helm_crds.tf line 263, in resource "kubectl_manifest" "helm_crds":
│  263: resource "kubectl_manifest" "helm_crds" {
│ 
╵
╷
│ Error: decoratorcontrollers.metacontroller.k8s.io failed to run apply: customresourcedefinitions.apiextensions.k8s.io "decoratorcontrollers.metacontroller.k8s.io" is forbidden: User "system:serviceaccount:gate-identity:gate-restricted" cannot patch resource "customresourcedefinitions" in API group "apiextensions.k8s.io" at the cluster scope
│ 
│   with kubectl_manifest.helm_crds["decoratorcontrollers.metacontroller.k8s.io"],
│   on helm_crds.tf line 263, in resource "kubectl_manifest" "helm_crds":
│  263: resource "kubectl_manifest" "helm_crds" {
│ 
╵
`

const helmVersionNotFound = `Unable to locate chart opensearch-operator: chart "opensearch-operator" version "99.99.99" not found in https://opensearch-project.github.io/opensearch-k8s-operator/ repository`

const helmOCINotFound = `Unable to locate chart oci://ghcr.io/plantonhq/charts/planton-operator: failed to perform "FetchReference" on source: ghcr.io/plantonhq/charts/planton-operator:99.99.99: not found`

const helmStaleCache = `Unable to locate chart podinfo: no cached repo found. (try 'helm repo update'): open /Users/me/Library/Caches/helm/repository/neo4j-index.yaml: no such file or directory`

func requireParts(t *testing.T, f *Failure) {
	t.Helper()
	for name, part := range map[string]string{"observed": f.Observed, "meaning": f.Meaning, "next step": f.NextStep} {
		if strings.TrimSpace(part) == "" {
			t.Fatalf("failure lacks its %s part: %+v", name, f)
		}
	}
}

func TestExplainUnreachableRepositoryOnceForBothDataSources(t *testing.T) {
	failures := Explain(tofuUnreachableRepository)
	if len(failures) != 1 {
		t.Fatalf("expected one explanation for one root cause (two data sources failed on the same host), got %d: %v", len(failures), failures)
	}
	f := failures[0]
	requireParts(t, f)
	if !strings.Contains(f.Observed, "https://charts.does-not-resolve.invalid ") || !strings.Contains(f.Observed, "no such host") {
		t.Fatalf("observation must name the repository and keep the raw text: %s", f.Observed)
	}
	if !strings.Contains(f.NextStep, "curl -I https://charts.does-not-resolve.invalid/index.yaml") {
		t.Fatalf("next step must be a runnable check: %s", f.NextStep)
	}
}

func TestExplainForbiddenCRDsOnceAcrossThreeResources(t *testing.T) {
	failures := Explain(tofuForbiddenCRDs)
	if len(failures) != 1 {
		t.Fatalf("expected one explanation, got %d", len(failures))
	}
	f := failures[0]
	requireParts(t, f)
	for _, want := range []string{
		"system:serviceaccount:gate-identity:gate-restricted",
		"may not patch customresourcedefinitions.apiextensions.k8s.io at the cluster scope",
		"is forbidden",
	} {
		if !strings.Contains(f.Observed, want) {
			t.Errorf("observation lacks %q: %s", want, f.Observed)
		}
	}
	if !strings.Contains(f.Meaning, "outside the Helm release") {
		t.Errorf("the CRD meaning must say why the module needs the right: %s", f.Meaning)
	}
	for _, want := range []string{"iac/permissions.yaml", "spec.crds.install", "helm template --include-crds"} {
		if !strings.Contains(f.NextStep, want) {
			t.Errorf("next step lacks %q: %s", want, f.NextStep)
		}
	}
}

func TestExplainForbiddenOnAnyResource(t *testing.T) {
	raw := `deployments.apps "web" is forbidden: User "alice" cannot create resource "deployments" in API group "apps" in the namespace "shop"`
	failures := Explain(raw)
	if len(failures) != 1 {
		t.Fatalf("expected one explanation, got %d", len(failures))
	}
	f := failures[0]
	requireParts(t, f)
	if !strings.Contains(f.Observed, `alice) may not create deployments.apps in the namespace "shop"`) {
		t.Errorf("observation: %s", f.Observed)
	}
	if !strings.Contains(f.NextStep, "grant create on deployments.apps") {
		t.Errorf("next step: %s", f.NextStep)
	}
}

// Recorded from a Kind API server holding the PlantonPlatform definition of
// planton-operator chart 0.8.1 while a declaration named the Gateway API
// front door the 0.9.0 definition introduced. The first is the kubectl
// provider's server-side apply (the shape both engines produce); the second
// is a strict create through kubectl.
const tofuFieldNotInSchema = `
╷
│ Error: planton failed to run apply: failed to create typed patch object (planton-gw/planton; planton.ai/v1, Kind=PlantonPlatform): .spec.ingress.gatewayRef: field not declared in schema
│ 
│   with kubectl_manifest.planton_platform,
│   on main.tf line 31, in resource "kubectl_manifest" "planton_platform":
│   31: resource "kubectl_manifest" "planton_platform" {
│ 
╵
`

const kubectlStrictUnknownField = `Error from server (BadRequest): error when creating "STDIN": PlantonPlatform in version "v1" cannot be handled as a PlantonPlatform: strict decoding error: unknown field "spec.ingress.gatewayRef"`

func TestExplainFieldNotInDefinitionUnderServerSideApply(t *testing.T) {
	failures := Explain(tofuFieldNotInSchema)
	if len(failures) != 1 {
		t.Fatalf("expected one explanation, got %d", len(failures))
	}
	f := failures[0]
	requireParts(t, f)
	for _, want := range []string{"spec.ingress.gatewayRef", "PlantonPlatform", "field not declared in schema"} {
		if !strings.Contains(f.Observed, want) {
			t.Errorf("observation lacks %q: %s", want, f.Observed)
		}
	}
	if !strings.Contains(f.Meaning, "predates spec.ingress.gatewayRef") {
		t.Errorf("meaning must name the field the definition predates: %s", f.Meaning)
	}
	if !strings.Contains(f.NextStep, "KubernetesPlantonOperator (spec.chart_version)") {
		t.Errorf("next step must name the resource that installs the PlantonPlatform definition: %s", f.NextStep)
	}
}

func TestExplainFieldNotInDefinitionUnderStrictDecoding(t *testing.T) {
	failures := Explain(kubectlStrictUnknownField)
	if len(failures) != 1 {
		t.Fatalf("expected one explanation, got %d", len(failures))
	}
	f := failures[0]
	requireParts(t, f)
	if !strings.Contains(f.Observed, "the API server rejected spec.ingress.gatewayRef on PlantonPlatform") {
		t.Errorf("observation: %s", f.Observed)
	}
}

func TestExplainFieldNotInDefinitionGenericOwner(t *testing.T) {
	raw := `failed to create typed patch object (shop/web; example.com/v1, Kind=Widget): .spec.color: field not declared in schema`
	failures := Explain(raw)
	if len(failures) != 1 {
		t.Fatalf("expected one explanation, got %d", len(failures))
	}
	f := failures[0]
	requireParts(t, f)
	if !strings.Contains(f.NextStep, "upgrade the resource that installs the Widget definition") {
		t.Errorf("a kind with no known owner gets the generic next step: %s", f.NextStep)
	}
}

func TestExplainHelmTexts(t *testing.T) {
	cases := []struct {
		name, raw, observedHas, nextStepHas string
	}{
		{"version not published", helmVersionNotFound, "chart opensearch-operator 99.99.99 is not in the index of https://opensearch-project.github.io/opensearch-k8s-operator/", "index.yaml"},
		{"oci version not published", helmOCINotFound, "registry has no chart planton-operator 99.99.99 at oci://ghcr.io/plantonhq/charts", "helm show chart oci://ghcr.io/plantonhq/charts/planton-operator"},
		{"stale cache", helmStaleCache, "could not read a repository index on this machine while locating chart podinfo", "helm repo update"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			failures := Explain(c.raw)
			if len(failures) != 1 {
				t.Fatalf("expected one explanation, got %d: %v", len(failures), failures)
			}
			requireParts(t, failures[0])
			if !strings.Contains(failures[0].Observed, c.observedHas) {
				t.Errorf("observation lacks %q: %s", c.observedHas, failures[0].Observed)
			}
			if !strings.Contains(failures[0].NextStep, c.nextStepHas) {
				t.Errorf("next step lacks %q: %s", c.nextStepHas, failures[0].NextStep)
			}
		})
	}
}

// A Terraform module's own postcondition already explains an unpublished
// version; Helm's raw text sits beside it. The explainer must not say it
// again.
func TestExplainStaysSilentWhenTheModuleSpokeFirst(t *testing.T) {
	output := `
│ Error: Resource postcondition failed
│ observed: chart podinfo 99.99.99 is not in the index of https://stefanprodan.github.io/podinfo
│ meaning: the pinned chart_version has not been published to that repository, or the version string is misspelled
│ next step: set spec.chart_version to a version listed at https://stefanprodan.github.io/podinfo/index.yaml, then re-run
╵
╷
│ Error: Error locating chart
│ Unable to locate chart podinfo: chart "podinfo" version "99.99.99" not found in https://stefanprodan.github.io/podinfo repository
╵
`
	if failures := Explain(output); len(failures) != 0 {
		t.Fatalf("the module already explained this; the explainer repeated it: %v", failures)
	}
}

func TestExplainRecognizesNothingInOrdinaryOutput(t *testing.T) {
	if failures := Explain("Apply complete! Resources: 5 added, 0 changed, 0 destroyed."); len(failures) != 0 {
		t.Fatalf("unexpected explanations: %v", failures)
	}
}

// Auth0's refusal of a client grant on a tenant whose plan lacks machine-to-
// machine access to Organizations. The runner line is recorded from a
// Free-plan tenant; the OpenTofu diagnostic is that same Auth0 text in the
// engine's standard frame (summary, then the failing resource's address),
// wrapped at terminal width.
const runnerAuth0OrganizationsEntitlement = `Auth0Client/user-lifecycle failed (16s): 403 Forbidden: Please upgrade your subscription to use Machine to Machine access to Organizations.`

const tofuAuth0OrganizationsEntitlement = `
╷
│ Error: 403 Forbidden: Please upgrade your subscription to use Machine to
│ Machine access to Organizations.
│ 
│   with auth0_client_grant.api_grants["0"],
│   on main.tf line 126, in resource "auth0_client_grant" "api_grants":
│  126: resource "auth0_client_grant" "api_grants" {
│ 
╵
`

func TestExplainAuth0OrganizationsEntitlement(t *testing.T) {
	cases := []struct {
		name, raw, observedHas string
	}{
		{"runner line", runnerAuth0OrganizationsEntitlement, "Auth0 refused a client grant"},
		{"tofu diagnostic", tofuAuth0OrganizationsEntitlement, `Auth0 refused the client grant auth0_client_grant.api_grants["0"]`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			failures := Explain(c.raw)
			if len(failures) != 1 {
				t.Fatalf("expected one explanation, got %d: %v", len(failures), failures)
			}
			f := failures[0]
			requireParts(t, f)
			for _, want := range []string{c.observedHas, "does not include machine-to-machine access to Organizations", "Please upgrade your subscription"} {
				if !strings.Contains(f.Observed, want) {
					t.Errorf("observation lacks %q: %s", want, f.Observed)
				}
			}
			for _, want := range []string{"allowAnyOrganization", "organizationUsage"} {
				if !strings.Contains(f.Meaning, want) {
					t.Errorf("meaning must name the field that asks for organization access (%q): %s", want, f.Meaning)
				}
				if !strings.Contains(f.NextStep, want) {
					t.Errorf("next step must name the field to remove (%q): %s", want, f.NextStep)
				}
			}
			if !strings.Contains(f.NextStep, "upgrade the Auth0 tenant's plan") {
				t.Errorf("next step must offer the plan upgrade: %s", f.NextStep)
			}
		})
	}
}

func TestExplainAuth0StaysSilent(t *testing.T) {
	spoken := Auth0OrganizationsEntitlement("", "403 Forbidden: Please upgrade your subscription to use Machine to Machine access to Organizations.").Error() +
		"\n" + runnerAuth0OrganizationsEntitlement
	cases := map[string]string{
		"the module spoke first": spoken,
		"an unrelated 403":       `Auth0Client/cli failed (3s): 403 Forbidden: Insufficient scope, expected any of: create:client_grants`,
	}
	for name, output := range cases {
		t.Run(name, func(t *testing.T) {
			if failures := Explain(output); len(failures) != 0 {
				t.Fatalf("expected no explanation, got %v", failures)
			}
		})
	}
}

func TestAnnotateCarriesCauseAndExplanation(t *testing.T) {
	cause := errors.New("failed to execute tofu command tofu apply")
	err := Annotate(cause, tofuForbiddenCRDs)
	if !errors.Is(err, cause) {
		t.Fatal("the annotated error must still be the cause for errors.Is")
	}
	var f *Failure
	if !errors.As(err, &f) {
		t.Fatal("the annotated error must expose its Failure to errors.As")
	}
	if !strings.HasPrefix(err.Error(), cause.Error()) || !strings.Contains(err.Error(), "next step:") {
		t.Fatalf("the text must lead with the engine's error and carry the explanation: %s", err.Error())
	}
	if same := Annotate(cause, "nothing here"); same != cause {
		t.Fatal("an output nothing recognizes must return the error unchanged")
	}
	if Annotate(nil, tofuForbiddenCRDs) != nil {
		t.Fatal("a nil error stays nil")
	}
}
