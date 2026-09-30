package module

import (
	"fmt"
	"strconv"

	kuberneteskubeprometheusstackv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskubeprometheusstack/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the remote-write auth Secret — never injected
	// into the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace the stack installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name. fullnameOverride is pinned to
	// it (values.go), and the chart derives every child name from the
	// fullname: `<name>-prometheus`, `<name>-alertmanager`,
	// `<name>-grafana`, `<name>-operator`, `<name>-kube-state-metrics`.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Whether the alertmanager / bundled-grafana halves deploy (proto
	// optional-bool defaults resolve to true).
	AlertmanagerEnabled bool
	GrafanaEnabled      bool

	// Child service names derived from the pinned fullname.
	PrometheusService   string
	AlertmanagerService string
	GrafanaService      string

	// In-cluster endpoints (empty when the half is disabled).
	PrometheusEndpoint   string
	AlertmanagerEndpoint string
	GrafanaEndpoint      string

	// Name of the Secret holding the bundled Grafana's admin
	// credentials: the referenced existing Secret when declared, else
	// the grafana subchart's own `<name>-grafana` Secret (generated once,
	// lookup-stable across upgrades). Empty when grafana is disabled.
	GrafanaAdminSecretName string

	// Name of the module-owned Secret carrying declared remote-write
	// basic-auth usernames (see vars.RemoteWriteAuthSecretSuffix). Empty
	// when no remote-write entry declares basic auth.
	RemoteWriteAuthSecretName string

	// Name of the module-owned Secret carrying the notification
	// credentials (see vars.NotificationsSecretSuffix). Empty when no
	// notifications are declared or Alertmanager is disabled.
	NotificationsSecretName string

	// The rendered notifications (notifications.go); nil when none are
	// declared.
	Notifications *notifications

	// kubectl one-liners for reaching the UIs from a workstation.
	PrometheusPortForwardCommand string
	GrafanaPortForwardCommand    string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesKubePrometheusStack.String(),
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

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name

	alertmanagerEnabled := true
	if am := spec.GetAlertmanager(); am != nil && am.Enabled != nil {
		alertmanagerEnabled = am.GetEnabled()
	}
	grafanaEnabled := true
	if g := spec.GetGrafana(); g != nil && g.Enabled != nil {
		grafanaEnabled = g.GetEnabled()
	}

	prometheusService := releaseName + "-prometheus"
	alertmanagerService := ""
	grafanaService := ""
	alertmanagerEndpoint := ""
	grafanaEndpoint := ""
	grafanaAdminSecretName := ""
	grafanaPortForwardCommand := ""

	if alertmanagerEnabled {
		alertmanagerService = releaseName + "-alertmanager"
		alertmanagerEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:9093",
			alertmanagerService, namespace)
	}
	if grafanaEnabled {
		grafanaService = releaseName + "-grafana"
		grafanaEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local",
			grafanaService, namespace)
		// The grafana subchart's generated admin Secret is named after
		// its own fullname (`<name>-grafana`); an existing Secret's own
		// name wins when declared.
		grafanaAdminSecretName = grafanaService
		if existing := spec.GetGrafana().GetAdminSecret(); existing != nil {
			grafanaAdminSecretName = existing.GetName()
		}
		grafanaPortForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s 3000:80",
			grafanaService, namespace)
	}

	remoteWriteAuthSecretName := ""
	for _, rw := range spec.GetPrometheus().GetRemoteWrite() {
		if rw.GetBasicAuth() != nil {
			remoteWriteAuthSecretName = releaseName + vars.RemoteWriteAuthSecretSuffix
			break
		}
	}

	notificationsSecretName := ""
	if alertmanagerEnabled && spec.GetAlertmanager().GetNotifications() != nil {
		notificationsSecretName = releaseName + vars.NotificationsSecretSuffix
	}

	locals := &Locals{
		Spec:                spec,
		Labels:              labels,
		Namespace:           namespace,
		ReleaseName:         releaseName,
		ChartVersion:        chartVersion,
		AlertmanagerEnabled: alertmanagerEnabled,
		GrafanaEnabled:      grafanaEnabled,
		PrometheusService:   prometheusService,
		AlertmanagerService: alertmanagerService,
		GrafanaService:      grafanaService,
		PrometheusEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:9090",
			prometheusService, namespace),
		AlertmanagerEndpoint:      alertmanagerEndpoint,
		GrafanaEndpoint:           grafanaEndpoint,
		GrafanaAdminSecretName:    grafanaAdminSecretName,
		RemoteWriteAuthSecretName: remoteWriteAuthSecretName,
		NotificationsSecretName:   notificationsSecretName,
		PrometheusPortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s 9090:9090",
			prometheusService, namespace),
		GrafanaPortForwardCommand: grafanaPortForwardCommand,
	}
	locals.Notifications = buildNotifications(locals)
	return locals
}
