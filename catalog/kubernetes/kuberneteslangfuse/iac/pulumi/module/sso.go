package module

import (
	"strings"

	kuberneteslangfusev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslangfuse/v1alpha1"
)

// ssoProvider is one configured single sign-on provider: Langfuse reads it
// from AUTH_<ID>_* variables on the web server.
type ssoProvider struct {
	// Id is the provider's identifier in variable names (upper-cased) and in
	// the instance Secret key of its client secret.
	Id     string
	Client *kuberneteslangfusev1alpha1.KubernetesLangfuseOAuthClient
	// Extra holds the provider-specific variables, keyed by the suffix after
	// AUTH_<ID>_ (or by a full name for Langfuse's claim variables).
	Extra map[string]string
}

// ssoProviders lists the configured providers in a fixed order (Terraform
// twin: local.sso_providers).
func ssoProviders(sso *kuberneteslangfusev1alpha1.KubernetesLangfuseSsoProviders) []ssoProvider {
	if sso == nil {
		return nil
	}
	var out []ssoProvider
	add := func(id string, client *kuberneteslangfusev1alpha1.KubernetesLangfuseOAuthClient, extra map[string]string) {
		if client != nil {
			out = append(out, ssoProvider{Id: id, Client: client, Extra: extra})
		}
	}
	if p := sso.GetGoogle(); p != nil {
		add("google", p.GetClient(), map[string]string{"ALLOWED_DOMAINS": csv(p.GetAllowedDomains())})
	}
	if p := sso.GetGithub(); p != nil {
		add("github", p.GetClient(), nil)
	}
	if p := sso.GetGithubEnterprise(); p != nil {
		add("github_enterprise", p.GetClient(), map[string]string{"BASE_URL": p.GetBaseUrl()})
	}
	if p := sso.GetGitlab(); p != nil {
		add("gitlab", p.GetClient(), map[string]string{"URL": p.GetUrl(), "ISSUER": p.GetIssuer()})
	}
	if p := sso.GetAzureAd(); p != nil {
		add("azure_ad", p.GetClient(), map[string]string{"TENANT_ID": p.GetTenantId()})
	}
	if p := sso.GetOkta(); p != nil {
		add("okta", p.GetClient(), map[string]string{"ISSUER": p.GetIssuer()})
	}
	if p := sso.GetAuthentik(); p != nil {
		add("authentik", p.GetClient(), map[string]string{"ISSUER": p.GetIssuer(), "AUTHORIZATION_URL": p.GetAuthorizationUrl()})
	}
	if p := sso.GetOnelogin(); p != nil {
		add("onelogin", p.GetClient(), map[string]string{"ISSUER": p.GetIssuer()})
	}
	if p := sso.GetAuth0(); p != nil {
		add("auth0", p.GetClient(), map[string]string{"ISSUER": p.GetIssuer()})
	}
	if p := sso.GetCognito(); p != nil {
		add("cognito", p.GetClient(), map[string]string{"ISSUER": p.GetIssuer()})
	}
	if p := sso.GetKeycloak(); p != nil {
		add("keycloak", p.GetClient(), map[string]string{
			"ISSUER": p.GetIssuer(), "NAME": p.GetName(), "SCOPE": p.GetScope(), "ID_TOKEN": optionalBool(p.UseIdToken),
		})
	}
	if p := sso.GetJumpcloud(); p != nil {
		add("jumpcloud", p.GetClient(), map[string]string{"ISSUER": p.GetIssuer(), "SCOPE": p.GetScope()})
	}
	if p := sso.GetCustom(); p != nil {
		add("custom", p.GetClient(), map[string]string{
			"ISSUER": p.GetIssuer(), "NAME": p.GetName(), "SCOPE": p.GetScope(),
			"ID_TOKEN": optionalBool(p.UseIdToken), "FETCH_USERINFO": trueOrEmpty(p.GetFetchUserinfo()),
		})
	}
	if p := sso.GetWorkos(); p != nil {
		add("workos", p.GetClient(), map[string]string{
			"ORGANIZATION_ID": p.GetOrganizationId(), "CONNECTION_ID": p.GetConnectionId(),
		})
	}
	if p := sso.GetWordpress(); p != nil {
		add("wordpress", p, nil)
	}
	return out
}

// env renders the provider's AUTH_<ID>_* variables into env.
func (p ssoProvider) env(env envSet, instanceSecretKey string) {
	prefix := "AUTH_" + strings.ToUpper(p.Id) + "_"
	env.plain(prefix+"CLIENT_ID", p.Client.GetClientId())
	env.instance(prefix+"CLIENT_SECRET", instanceSecretKey)
	env.plain(prefix+"ALLOW_ACCOUNT_LINKING", trueOrEmpty(p.Client.GetAllowAccountLinking()))
	if m := p.Client.GetClientAuthMethod(); m != kuberneteslangfusev1alpha1.KubernetesLangfuseOAuthClient_client_auth_method_unspecified {
		env.plain(prefix+"CLIENT_AUTH_METHOD", m.String())
	}
	var checks []string
	for _, c := range p.Client.GetChecks() {
		if c == kuberneteslangfusev1alpha1.KubernetesLangfuseOAuthClient_no_checks {
			checks = append(checks, "none")
		} else {
			checks = append(checks, c.String())
		}
	}
	env.plain(prefix+"CHECKS", csv(checks))
	if a := p.Client.GetIdTokenSignedResponseAlg(); a != kuberneteslangfusev1alpha1.KubernetesLangfuseOAuthClient_id_token_algorithm_unspecified {
		env.plain(prefix+"ID_TOKEN_SIGNED_RESPONSE_ALG", a.String())
	}
	for suffix, value := range p.Extra {
		env.plain(prefix+suffix, value)
	}
}
