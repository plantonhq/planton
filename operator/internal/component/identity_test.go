package component

import (
	"slices"
	"strings"
	"testing"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

func ingressPlatform(enabled bool) *v1.PlantonPlatform {
	p := &v1.PlantonPlatform{}
	p.Spec.Version = "v1.0.0"
	if enabled {
		p.Spec.Ingress = &v1.IngressSpec{Enabled: true, Hostname: "planton.example.com"}
	}
	return p
}

// Identity is not switchable: every install deploys the identity server, so
// "deploy Planton" always yields an authenticated platform -- with or without
// ingress. The gateway serves sign-in over port-forward when ingress is off.
func TestIdentity_IsEnabledUnconditionally(t *testing.T) {
	id := &Identity{}
	if !id.IsEnabled(ingressPlatform(false)) {
		t.Error("identity must be enabled without ingress (gateway front door)")
	}
	if !id.IsEnabled(ingressPlatform(true)) {
		t.Error("identity must be enabled with ingress")
	}
}

// Exactly one front door: the gateway serves the platform precisely when
// ingress does not.
func TestGateway_IsEnabledOppositeOfIngress(t *testing.T) {
	g := &Gateway{}
	if !g.IsEnabled(ingressPlatform(false)) {
		t.Error("gateway must be enabled without ingress")
	}
	if g.IsEnabled(ingressPlatform(true)) {
		t.Error("gateway must be disabled when ingress is the front door")
	}
}

func TestIdentityRealmAndAdminDefaults(t *testing.T) {
	p := ingressPlatform(true)
	if got := identityRealm(p); got != "planton" {
		t.Errorf("realm = %q, want the planton default", got)
	}
	// There is deliberately NO generic default admin identity: unset means no
	// sign-in user is seeded, and the component status says so plainly.
	if got := identityAdminEmail(p); got != "" {
		t.Errorf("adminEmail = %q, want empty (no pre-baked admin identity)", got)
	}

	p.Spec.Identity = &v1.IdentitySpec{Realm: "corp", AdminEmail: "boss@corp.example"}
	if got := identityRealm(p); got != "corp" {
		t.Errorf("realm = %q, want corp", got)
	}
	if got := identityAdminEmail(p); got != "boss@corp.example" {
		t.Errorf("adminEmail = %q, want boss@corp.example", got)
	}
}

// The control plane performs eager OIDC discovery at boot -- it is always an
// OIDC relying party now -- so it must wait for the identity server in both
// front-door modes.
func TestControlPlaneDependencies_IdentityAlways(t *testing.T) {
	cp := &ControlPlane{}

	for _, ingress := range []bool{false, true} {
		deps := cp.Dependencies(ingressPlatform(ingress))
		if !slices.Contains(deps, "identity") {
			t.Errorf("deps (ingress=%v) = %v: identity must gate the relying-party boot", ingress, deps)
		}
		for _, core := range []string{"postgresql", "redis", "temporal"} {
			if !slices.Contains(deps, core) {
				t.Errorf("core dependency %s missing (ingress=%v): %v", core, ingress, deps)
			}
		}
	}
}

// Both Ready-message arms must explain the one-time credential semantics and
// name the same recovery path -- the founder-hit failure was a consumed
// password meeting a bare "invalid username or password" with no surface
// saying why.
func TestIdentityReadyMessages(t *testing.T) {
	const publicURL = "http://planton.example.com"

	restored := identityRestoredRealmReadyMessage("planton", "planton-postgres-a4d0f96d", publicURL)
	if strings.Contains(restored, "first visitor") || strings.Contains(restored, "read it with") || !strings.Contains(restored, "planton-postgres-a4d0f96d") || !strings.Contains(restored, "sign in as before") {
		t.Errorf("a restored realm has its people already: no setup code, the source named, sign-in as before: %q", restored)
	}
	if !strings.Contains(restored, publicURL+resources.IdentityPathPrefix) || !strings.Contains(restored, resources.IdentityBootstrapAdminSecretName("planton")) {
		t.Errorf("the restored-realm message names the same recovery path as the other arms: %q", restored)
	}
	if identityRestoredFrom(&v1.PlantonPlatform{}) != "" || identityRestoredFrom(&v1.PlantonPlatform{Status: v1.PlantonPlatformStatus{Backup: &v1.BackupStatus{RestoredFrom: "x"}}}) != "x" {
		t.Error("identityRestoredFrom reads the database component's fact, empty when there is none")
	}

	setup := identitySetupModeReadyMessage("planton", "planton", publicURL)
	if !strings.Contains(setup, "setup code") {
		t.Errorf("setup message must name the setup code: %q", setup)
	}
	if !strings.Contains(setup, publicURL+resources.IdentityPathPrefix) {
		t.Errorf("setup message must name the identity admin console URL: %q", setup)
	}
	if !strings.Contains(setup, resources.IdentityBootstrapAdminSecretName("planton")) {
		t.Errorf("setup message must name the bootstrap-admin Secret: %q", setup)
	}

	declared := identityDeclaredAdminReadyMessage("planton", publicURL)
	if !strings.Contains(declared, resources.IdentityAdminUserSecretName("planton")) {
		t.Errorf("declared message must name the credentials Secret: %q", declared)
	}
	if !strings.Contains(declared, "one-time") {
		t.Errorf("declared message must state the password is one-time: %q", declared)
	}
	if !strings.Contains(declared, "invalid username or password") {
		t.Errorf("declared message must name the symptom a consumed password produces: %q", declared)
	}
	if !strings.Contains(declared, publicURL+resources.IdentityPathPrefix) ||
		!strings.Contains(declared, resources.IdentityBootstrapAdminSecretName("planton")) {
		t.Errorf("declared message must name the recovery path (admin console + bootstrap Secret): %q", declared)
	}
}

// The credentials Secret itself is where a person goes looking when sign-in
// fails, so its self-describing note must carry the same semantics.
func TestIdentityOneTimePasswordNote(t *testing.T) {
	note := resources.IdentityOneTimePasswordNote("planton", "http://planton.example.com")
	for _, want := range []string{
		"one-time",
		"invalid username or password",
		"http://planton.example.com" + resources.IdentityPathPrefix,
		resources.IdentityBootstrapAdminSecretName("planton"),
	} {
		if !strings.Contains(note, want) {
			t.Errorf("note missing %q: %q", want, note)
		}
	}
}

func TestPublicHostname(t *testing.T) {
	if got := publicHostname("https://planton.example.com"); got != "planton.example.com" {
		t.Errorf("https hostname = %q", got)
	}
	if got := publicHostname("http://planton.203-0-113-7.sslip.io"); got != "planton.203-0-113-7.sslip.io" {
		t.Errorf("http hostname = %q", got)
	}
}
