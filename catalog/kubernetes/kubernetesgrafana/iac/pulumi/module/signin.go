package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	kubernetesgrafanav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
)

// signIn is the rendering of auth.google and auth.generic_oauth: the
// grafana.ini sections Grafana reads its providers from, the environment
// variables that carry each client secret, and the data of the
// module-owned Secret those variables read.
//
// The client secrets arrive resolved from managed-secret references and
// never enter grafana.ini (the chart renders it into a ConfigMap, and its
// assertNoLeakedSecrets helper refuses a plaintext client_secret there).
// Grafana's GF_AUTH_<PROVIDER>_CLIENT_SECRET variables override the ini
// keys, so each secret rides envValueFrom -> secretKeyRef into the
// `<name>-sso` Secret this module creates before the release.
//
// PARITY: the Terraform module renders the same sections, keys and
// variable names (locals.tf, the sign-in block). Keep them in lockstep.
type signIn struct {
	Ini          map[string]interface{}
	EnvValueFrom map[string]interface{}
	SecretData   map[string]string
}

// buildSignIn returns nil when no provider is declared, so an install
// without sign-in renders nothing new and creates no Secret.
func buildSignIn(auth *kubernetesgrafanav1alpha1.KubernetesGrafanaAuth, secretName string) *signIn {
	google := auth.GetGoogle()
	generic := auth.GetGenericOauth()
	if google == nil && generic == nil {
		return nil
	}

	s := &signIn{
		Ini:          map[string]interface{}{},
		EnvValueFrom: map[string]interface{}{},
		SecretData:   map[string]string{},
	}

	if google != nil {
		section := map[string]interface{}{
			"enabled":       true,
			"client_id":     google.GetClientId(),
			"allow_sign_up": boolOrDefault(google.AllowSignUp, true),
		}
		if len(google.GetAllowedDomains()) > 0 {
			section["allowed_domains"] = strings.Join(google.GetAllowedDomains(), " ")
		}
		if google.GetHostedDomain() != "" {
			section["hosted_domain"] = google.GetHostedDomain()
		}
		if google.GetAutoLogin() {
			section["auto_login"] = true
		}
		if google.GetRoleAttributePath() != "" {
			section["role_attribute_path"] = google.GetRoleAttributePath()
			// Grafana ships [auth.google] with skip_org_role_sync = true, which
			// silently ignores role_attribute_path; a declared mapping must
			// apply, so sync is switched on with it.
			section["skip_org_role_sync"] = false
		}
		if google.GetRoleAttributeStrict() {
			section["role_attribute_strict"] = true
		}
		s.Ini["auth.google"] = section
		s.SecretData[vars.GoogleClientSecretKey] = google.GetClientSecret().GetValue()
		s.EnvValueFrom[vars.GoogleClientSecretEnv] = secretKeyRef(secretName, vars.GoogleClientSecretKey)
	}

	if generic != nil {
		name := generic.GetName()
		if name == "" {
			name = vars.DefaultGenericOAuthName
		}
		scopes := generic.GetScopes()
		if len(scopes) == 0 {
			scopes = vars.DefaultOAuthScopes
		}
		section := map[string]interface{}{
			"enabled":       true,
			"name":          name,
			"client_id":     generic.GetClientId(),
			"auth_url":      generic.GetAuthUrl(),
			"token_url":     generic.GetTokenUrl(),
			"api_url":       generic.GetApiUrl(),
			"scopes":        strings.Join(scopes, " "),
			"allow_sign_up": boolOrDefault(generic.AllowSignUp, true),
			"use_pkce":      boolOrDefault(generic.UsePkce, true),
		}
		optionalStrings := map[string]string{
			"email_attribute_path":  generic.GetEmailAttributePath(),
			"login_attribute_path":  generic.GetLoginAttributePath(),
			"name_attribute_path":   generic.GetNameAttributePath(),
			"role_attribute_path":   generic.GetRoleAttributePath(),
			"groups_attribute_path": generic.GetGroupsAttributePath(),
			"allowed_groups":        strings.Join(generic.GetAllowedGroups(), " "),
			"allowed_domains":       strings.Join(generic.GetAllowedDomains(), " "),
		}
		for key, value := range optionalStrings {
			if value != "" {
				section[key] = value
			}
		}
		if generic.GetRoleAttributePath() != "" {
			// Generic OAuth already syncs roles by default; stated anyway so
			// both providers read the same when a mapping is declared.
			section["skip_org_role_sync"] = false
		}
		if generic.GetRoleAttributeStrict() {
			section["role_attribute_strict"] = true
		}
		if generic.GetAutoLogin() {
			section["auto_login"] = true
		}
		s.Ini["auth.generic_oauth"] = section
		s.SecretData[vars.GenericOAuthClientSecretKey] = generic.GetClientSecret().GetValue()
		s.EnvValueFrom[vars.GenericOAuthClientSecretEnv] = secretKeyRef(secretName, vars.GenericOAuthClientSecretKey)
	}

	// Once the manifest declares sign-in it owns sign-in: Grafana stores
	// provider settings saved through Administration > Authentication (the
	// SSO settings API) in its database, and those override grafana.ini
	// and the environment. configurable_providers is the list of providers
	// that screen may edit; Grafana skips empty values when it layers
	// custom configuration over its defaults, so an empty list would leave
	// every provider editable. A list naming no provider is what replaces
	// the default, and no screen can re-point sign-in or add a second OAuth
	// way in beside the declared one.
	s.Ini["sso_settings"] = map[string]interface{}{"configurable_providers": vars.NoConfigurableProviders}

	return s
}

// credentialsChecksum fingerprints a module-owned Secret's data for a pod
// annotation. The workload reads the Secret only at start, so a changed
// value must change the pod template for the next apply to roll it. The
// fingerprint is the SHA-256 of the data's JSON encoding (keys sorted),
// which is exactly what OpenTofu's sha256(jsonencode(map)) computes, so
// both engines stamp the same annotation for the same data.
func credentialsChecksum(data map[string]string) string {
	encoded, _ := json.Marshal(data)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func secretKeyRef(name, key string) map[string]interface{} {
	return map[string]interface{}{
		"secretKeyRef": map[string]interface{}{
			"name": name,
			"key":  key,
		},
	}
}

func boolOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
