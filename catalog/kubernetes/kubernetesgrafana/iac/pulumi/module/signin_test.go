package module

import (
	"fmt"
	"strings"
	"testing"

	kubernetesgrafanav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin what typed sign-in renders into the chart values: each
// client secret leaves the configuration for the module-owned Secret and
// reaches Grafana only through an environment variable, the admin screen
// is locked once sign-in is declared, a rotated secret changes the pod
// template, and an install without sign-in renders none of it.
// The OpenTofu twin is held to the same shapes by the live sign-in
// scenario, which signs in through both engines' installs.

const (
	googleSecret  = "GOCSPX-test-google-secret-value"
	genericSecret = "okta-test-client-secret-value"
)

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func localsWithAuth(auth *kubernetesgrafanav1alpha1.KubernetesGrafanaAuth) *Locals {
	return initializeLocals(nil, &kubernetesgrafanav1alpha1.KubernetesGrafanaStackInput{
		Target: &kubernetesgrafanav1alpha1.KubernetesGrafana{
			Metadata: &shared.CloudResourceMetadata{Name: "hub"},
			Spec: &kubernetesgrafanav1alpha1.KubernetesGrafanaSpec{
				Namespace: literal("observability"),
				Server:    &kubernetesgrafanav1alpha1.KubernetesGrafanaServer{RootUrl: "https://grafana.example.com"},
				Auth:      auth,
			},
		},
	})
}

func googleAuth(secret string) *kubernetesgrafanav1alpha1.KubernetesGrafanaAuth {
	return &kubernetesgrafanav1alpha1.KubernetesGrafanaAuth{
		Google: &kubernetesgrafanav1alpha1.KubernetesGrafanaGoogleSignIn{
			ClientId:          "123-abc.apps.googleusercontent.com",
			ClientSecret:      literal(secret),
			AllowedDomains:    []string{"example.com", "example.org"},
			HostedDomain:      "example.com",
			RoleAttributePath: "email == 'lead@example.com' && 'Admin' || 'Viewer'",
		},
	}
}

func renderedValues(t *testing.T, locals *Locals) map[string]interface{} {
	t.Helper()
	values, err := buildHelmValues(locals)
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	return values
}

func iniSection(t *testing.T, values map[string]interface{}, section string) map[string]interface{} {
	t.Helper()
	ini, ok := values["grafana.ini"].(map[string]interface{})
	if !ok {
		t.Fatalf("grafana.ini not rendered")
	}
	s, ok := ini[section].(map[string]interface{})
	if !ok {
		t.Fatalf("grafana.ini section %q not rendered; sections: %v", section, ini)
	}
	return s
}

func TestGoogleSignInRendersSettingsAndKeepsTheSecretOutOfValues(t *testing.T) {
	locals := localsWithAuth(googleAuth(googleSecret))
	values := renderedValues(t, locals)

	google := iniSection(t, values, "auth.google")
	want := map[string]interface{}{
		"enabled":             true,
		"client_id":           "123-abc.apps.googleusercontent.com",
		"allowed_domains":     "example.com example.org",
		"hosted_domain":       "example.com",
		"allow_sign_up":       true,
		"role_attribute_path": "email == 'lead@example.com' && 'Admin' || 'Viewer'",
		"skip_org_role_sync":  false,
	}
	for key, value := range want {
		if google[key] != value {
			t.Errorf("auth.google %s = %v, want %v", key, google[key], value)
		}
	}
	if _, present := google["client_secret"]; present {
		t.Error("auth.google carries client_secret; the chart would refuse it and the ConfigMap would expose it")
	}
	if rendered := fmt.Sprintf("%v", values); strings.Contains(rendered, googleSecret) {
		t.Error("the client secret appears in the rendered chart values")
	}

	envValueFrom := values["envValueFrom"].(map[string]interface{})
	ref := envValueFrom["GF_AUTH_GOOGLE_CLIENT_SECRET"].(map[string]interface{})["secretKeyRef"].(map[string]interface{})
	if ref["name"] != "hub-sso" || ref["key"] != "google-client-secret" {
		t.Errorf("GF_AUTH_GOOGLE_CLIENT_SECRET reads %v, want hub-sso/google-client-secret", ref)
	}
	if got := locals.SignIn.SecretData["google-client-secret"]; got != googleSecret {
		t.Errorf("module-owned Secret carries %q under google-client-secret, want the resolved secret", got)
	}
}

func TestGenericOAuthRendersDefaultsAndItsOwnVariable(t *testing.T) {
	locals := localsWithAuth(&kubernetesgrafanav1alpha1.KubernetesGrafanaAuth{
		GenericOauth: &kubernetesgrafanav1alpha1.KubernetesGrafanaGenericOAuthSignIn{
			ClientId:            "0oa1example",
			ClientSecret:        literal(genericSecret),
			AuthUrl:             "https://id.example.com/authorize",
			TokenUrl:            "https://id.example.com/token",
			ApiUrl:              "https://id.example.com/userinfo",
			GroupsAttributePath: "groups",
			AllowedGroups:       []string{"platform", "sre"},
		},
	})
	values := renderedValues(t, locals)

	generic := iniSection(t, values, "auth.generic_oauth")
	want := map[string]interface{}{
		"name":                  "OAuth",
		"scopes":                "openid email profile",
		"allow_sign_up":         true,
		"use_pkce":              true,
		"groups_attribute_path": "groups",
		"allowed_groups":        "platform sre",
		"token_url":             "https://id.example.com/token",
	}
	for key, value := range want {
		if generic[key] != value {
			t.Errorf("auth.generic_oauth %s = %v, want %v", key, generic[key], value)
		}
	}
	for _, absent := range []string{"client_secret", "allowed_domains", "role_attribute_path", "auto_login"} {
		if _, present := generic[absent]; present {
			t.Errorf("auth.generic_oauth renders %s, which was not declared", absent)
		}
	}
	if _, ok := values["envValueFrom"].(map[string]interface{})["GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET"]; !ok {
		t.Error("GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET is not wired")
	}
	if locals.SignIn.SecretData["generic-oauth-client-secret"] != genericSecret {
		t.Error("module-owned Secret does not carry the generic OAuth client secret")
	}
}

func TestDeclaredSignInLocksTheAdminScreen(t *testing.T) {
	values := renderedValues(t, localsWithAuth(googleAuth(googleSecret)))
	lock := iniSection(t, values, "sso_settings")
	// Grafana skips an empty value when it layers custom configuration over
	// its defaults, which would leave every provider editable; the lock is a
	// non-empty list naming no provider.
	if providers, ok := lock["configurable_providers"]; !ok || providers != "none" {
		t.Errorf("sso_settings.configurable_providers = %v, want \"none\"", providers)
	}
}

func TestRotatedSecretChangesThePodTemplate(t *testing.T) {
	before := renderedValues(t, localsWithAuth(googleAuth(googleSecret)))
	after := renderedValues(t, localsWithAuth(googleAuth(googleSecret+"-rotated")))
	checksum := func(values map[string]interface{}) string {
		annotations, ok := values["podAnnotations"].(map[string]interface{})
		if !ok {
			t.Fatal("podAnnotations not rendered")
		}
		return annotations["checksum/credentials"].(string)
	}
	if checksum(before) == checksum(after) {
		t.Error("a rotated client secret left the pod template unchanged; Grafana would keep the old secret until someone restarts it")
	}
	if checksum(before) != checksum(renderedValues(t, localsWithAuth(googleAuth(googleSecret)))) {
		t.Error("the checksum is not deterministic; every apply would roll Grafana")
	}
}

func TestChecksumMatchesOpenTofusEncoding(t *testing.T) {
	// OpenTofu stamps sha256(jsonencode(local.sso_secret_data)); jsonencode
	// sorts keys and escapes <, > and & exactly as encoding/json does. The
	// expected digest is what `tofu console` prints for
	// sha256(jsonencode({b="2", a="<&>"})), so both engines stamp the same
	// annotation and a switch of engine does not roll Grafana.
	got := credentialsChecksum(map[string]string{"b": "2", "a": "<&>"})
	const want = "915a6e5101546601602b070e7477e0a9dca0cb5925ab66555ae1e0ddce06427d"
	if got != want {
		t.Errorf("checksum = %s, want OpenTofu's %s", got, want)
	}
}

func TestNoSignInRendersNothingNew(t *testing.T) {
	locals := localsWithAuth(&kubernetesgrafanav1alpha1.KubernetesGrafanaAuth{AnonymousEnabled: true})
	if locals.SignIn != nil {
		t.Fatal("sign-in rendered without a declared provider")
	}
	values := renderedValues(t, locals)
	if _, present := values["podAnnotations"]; present {
		t.Error("podAnnotations rendered without sign-in")
	}
	ini := values["grafana.ini"].(map[string]interface{})
	for _, section := range []string{"auth.google", "auth.generic_oauth", "sso_settings"} {
		if _, present := ini[section]; present {
			t.Errorf("grafana.ini renders %s without sign-in", section)
		}
	}
}
