package module

// Chart identity — must stay byte-identical with the Terraform module's
// locals (helm_chart_name / helm_chart_repo): cross-engine chart-name drift
// deploys two different products from one manifest.
//
// KNOW THIS about the repo URL: the grafana chart's canonical home is the
// grafana-community index — the old https://grafana.github.io/helm-charts
// stopped serving new versions at chart 10.5.x, and kube-prometheus-stack's
// own dependency block points at the community repo. Never "fix" this back
// to the grafana.github.io URL.
var vars = struct {
	HelmChartName string
	HelmChartRepo string
	// Fallback when spec.chart_version is unset — mirror of the proto
	// field's default option and the Terraform module's coalesce. Chart
	// 12.8.0 ships Grafana 13.1.1; the chart pin governs.
	DefaultChartVersion string
	// Fallback for spec.storage.size — mirror of the proto default.
	DefaultStorageSize string
	// Grafana's admin-credential Secret keys — the chart's own key names,
	// used both for the chart-generated Secret and as the defaults for an
	// existing Secret's key overrides.
	AdminUserKey     string
	AdminPasswordKey string
	// Service port the chart exposes (targets container 3000).
	ServicePort int
	// Sign-in: the module-owned Secret is `<name>` plus this suffix; its
	// keys and the Grafana variables that read them are fixed, and the
	// Terraform module uses the same literals (sso_secret_name and the
	// sign-in env map in locals.tf).
	SsoSecretSuffix             string
	GoogleClientSecretKey       string
	GoogleClientSecretEnv       string
	GenericOAuthClientSecretKey string
	GenericOAuthClientSecretEnv string
	// Mirrors of the proto defaults for generic OAuth (the button label)
	// and of Grafana's own default scope list.
	DefaultGenericOAuthName string
	DefaultOAuthScopes      []string
	// sso_settings.configurable_providers once sign-in is declared: a
	// value naming no provider (Grafana ignores an empty one).
	NoConfigurableProviders string
	// Pod annotation carrying the fingerprint of every module-owned
	// Secret Grafana reads only at start; the same key on every catalog
	// module that uses the pattern.
	CredentialsChecksumAnnotation string
}{
	HelmChartName:                 "grafana",
	HelmChartRepo:                 "https://grafana-community.github.io/helm-charts",
	DefaultChartVersion:           "12.8.0",
	DefaultStorageSize:            "10Gi",
	AdminUserKey:                  "admin-user",
	AdminPasswordKey:              "admin-password",
	ServicePort:                   80,
	SsoSecretSuffix:               "-sso",
	GoogleClientSecretKey:         "google-client-secret",
	GoogleClientSecretEnv:         "GF_AUTH_GOOGLE_CLIENT_SECRET",
	GenericOAuthClientSecretKey:   "generic-oauth-client-secret",
	GenericOAuthClientSecretEnv:   "GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET",
	DefaultGenericOAuthName:       "OAuth",
	DefaultOAuthScopes:            []string{"openid", "email", "profile"},
	NoConfigurableProviders:       "none",
	CredentialsChecksumAnnotation: "checksum/credentials",
}
