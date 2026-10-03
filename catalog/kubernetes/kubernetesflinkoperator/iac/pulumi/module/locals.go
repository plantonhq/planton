package module

import (
	"sort"
	"strconv"
	"strings"

	kubernetesflinkoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflinkoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesflinkoperatorv1alpha1.KubernetesFlinkOperatorSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the keystore-password Secret — never injected
	// into the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name. The module pins the chart's
	// nameOverride (and fullnameOverride) to it — the Deployment
	// renders from `flink-operator.name`, which honors nameOverride.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// HelmChartRepo is the versioned ASF downloads directory serving
	// the chart — the version is part of the repository URL itself.
	HelmChartRepo string

	// WebhookEnabled resolves spec.webhook_enabled with unset-or-true =
	// enabled (the upstream default this spec keeps). With the webhook
	// enabled the chart renders cert-manager Issuer/Certificate
	// UNCONDITIONALLY (cert-manager is this kind's registry
	// prerequisite) and both webhook configurations are failurePolicy
	// Fail.
	WebhookEnabled bool

	// WebhookKeystoreSecretName is the module-owned keystore-password
	// Secret ("<name>-webhook-keystore") — the replacement for the
	// chart's HARDCODED PUBLIC PASSWORD default ("password1234", base64
	// in templates/webhook/secret.yaml behind keystore.
	// useDefaultPassword=true), which must never ship.
	WebhookKeystoreSecretName string

	// WebhookService is the operator's webhook Service name — CHART-
	// FIXED, not fullname-derived. Empty when the webhook is disabled;
	// matches outputs.proto.
	WebhookService string

	// JobServiceAccount is the service account name Flink job pods run
	// as, resolved to the chart default "flink" when unset. The chart
	// marks it helm.sh/resource-policy: keep — it survives uninstall so
	// running jobs never lose their identity.
	JobServiceAccount string

	// OperatorConfig is spec.operator_config merged with the module-
	// owned leader-election keys. Leader election is module-owned: any
	// replica count beyond 1 REQUIRES it (the chart's own contract — it
	// refuses multi-replica installs without it), so the two keys
	// render exactly when replicas > 1 — never a spec knob that could
	// drift from the replica count. Key spelling verified against the
	// operator docs (kubernetes.operator.leader-election.enabled /
	// .lease-name).
	OperatorConfig map[string]string

	// FlinkConfFile is OperatorConfig rendered as Flink configuration
	// lines, sorted by key. NOTE the Flink conf format is "key: value"
	// (colon-space), NOT "key=value" — the file is YAML-flavored Flink
	// configuration.
	FlinkConfFile string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesflinkoperatorv1alpha1.KubernetesFlinkOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesFlinkOperator.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	releaseName := target.Metadata.Name

	// Unset-or-true = enabled (proto3 optional bool presence).
	webhookEnabled := spec.WebhookEnabled == nil || spec.GetWebhookEnabled()

	webhookService := ""
	if webhookEnabled {
		webhookService = vars.WebhookServiceName
	}

	jobServiceAccount := spec.GetJobServiceAccount()
	if jobServiceAccount == "" {
		jobServiceAccount = "flink"
	}

	operatorConfig := map[string]string{}
	for k, v := range spec.GetOperatorConfig() {
		operatorConfig[k] = v
	}
	if spec.GetReplicas() > 1 {
		operatorConfig["kubernetes.operator.leader-election.enabled"] = "true"
		operatorConfig["kubernetes.operator.leader-election.lease-name"] = releaseName + "-lease"
	}

	keys := make([]string, 0, len(operatorConfig))
	for k := range operatorConfig {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, k+": "+operatorConfig[k])
	}
	flinkConfFile := strings.Join(lines, "\n")

	return &Locals{
		Spec:                      spec,
		Labels:                    labels,
		Namespace:                 spec.Namespace.GetValue(),
		ReleaseName:               releaseName,
		ChartVersion:              chartVersion,
		HelmChartRepo:             helmChartRepo(chartVersion),
		WebhookEnabled:            webhookEnabled,
		WebhookKeystoreSecretName: releaseName + "-webhook-keystore",
		WebhookService:            webhookService,
		JobServiceAccount:         jobServiceAccount,
		OperatorConfig:            operatorConfig,
		FlinkConfFile:             flinkConfFile,
	}
}
