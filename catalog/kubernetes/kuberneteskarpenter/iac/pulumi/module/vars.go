package module

var vars = struct {
	// HelmOciRepo is the OCI registry path holding BOTH charts
	// (oci://public.ecr.aws/karpenter/karpenter and .../karpenter-crd).
	// Pulumi's helm.v3.Release does not resolve oci:// through
	// RepositoryOpts the way the Terraform provider does — the chart
	// reference must be the JOINED "<repo>/<chart>" string (see main.go).
	HelmOciRepo string
	// HelmChartName is the controller chart ("karpenter").
	HelmChartName string
	// CrdChartName is the companion CRD chart ("karpenter-crd") — upstream's
	// supported mechanism for keeping CRDs upgradable (Helm installs the
	// copies bundled inside the main chart once and NEVER upgrades them).
	CrdChartName string
	// DefaultChartVersion is the fallback when spec.chart_version is unset
	// AND the platform's defaulting middleware did not run. Keep aligned
	// with the spec default. The karpenter and karpenter-crd charts version
	// together with the controller (both 1.14.0 = Karpenter 1.14.0), so ONE
	// version pins BOTH releases.
	DefaultChartVersion string
	// ReleaseName is FIXED: Karpenter owns the cluster-wide karpenter.sh
	// label domain, its CRDs, and node lifecycle — one installation per
	// cluster is an upstream constraint, so the release name never derives
	// from metadata.name.
	ReleaseName string
	// CrdReleaseName is the FIXED name of the CRD release, for the same
	// singleton reason.
	CrdReleaseName string
	// ServiceAccountName is the controller's service-account name — the
	// subject IRSA trust policies and EKS Pod Identity associations are
	// written against, so it is surfaced as an output. Derivation
	// (verified in the served chart's _helpers.tpl + serviceaccount.yaml):
	// serviceAccount.create defaults true and serviceAccount.name defaults
	// to the fullname template; with no fullnameOverride and the release
	// name ("karpenter") containing the chart name ("karpenter"), fullname
	// IS the release name — so the SA is named "karpenter".
	ServiceAccountName string
	// HelmTimeoutSeconds bounds the atomic install/upgrade of both
	// releases. 600s covers image pulls on cold clusters; atomic rolls
	// back on expiry so a wedged install never lingers half-deployed.
	HelmTimeoutSeconds int

	// Chart-value defaults resolved module-side so both engines render the
	// SAME values whether or not the platform's defaulting middleware ran.
	// Every one mirrors the served chart's values.yaml — drift here would
	// silently change what "unset" means.
	DefaultReplicas                int
	DefaultLogLevel                string
	DefaultBatchMaxDuration        string
	DefaultBatchIdleDuration       string
	DefaultPreferencePolicy        string
	DefaultMinValuesPolicy         string
	DefaultPriorityClassName       string
	DefaultReservedEnis            int
	DefaultVmMemoryOverheadPercent string
}{
	// Chart identity — MUST be identical in the Terraform module's locals
	// (cross-engine chart drift installs different software per engine).
	HelmOciRepo:         "oci://public.ecr.aws/karpenter",
	HelmChartName:       "karpenter",
	CrdChartName:        "karpenter-crd",
	DefaultChartVersion: "1.14.0",
	ReleaseName:         "karpenter",
	CrdReleaseName:      "karpenter-crd",
	ServiceAccountName:  "karpenter",
	HelmTimeoutSeconds:  600,

	DefaultReplicas:                2,
	DefaultLogLevel:                "info",
	DefaultBatchMaxDuration:        "10s",
	DefaultBatchIdleDuration:       "1s",
	DefaultPreferencePolicy:        "Respect",
	DefaultMinValuesPolicy:         "Strict",
	DefaultPriorityClassName:       "system-cluster-critical",
	DefaultReservedEnis:            0,
	DefaultVmMemoryOverheadPercent: "0.075",
}
