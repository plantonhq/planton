package module

import (
	"fmt"
	"strconv"

	kubernetesopensearchv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopensearch/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesOpenSearch *kubernetesopensearchv1alpha1.KubernetesOpenSearch
	Spec                 *kubernetesopensearchv1alpha1.KubernetesOpenSearchSpec

	// Resource-identity labels stamped on the module-created objects
	// (namespace, OpenSearchCluster). The operator derives ITS objects'
	// identity from the cluster name; these labels tie the family back to
	// the Planton resource.
	Labels map[string]string

	// Namespace the cluster lives in (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// ClusterName is metadata.name — the naming root the operator derives
	// every object from: StatefulSets `<name>-<pool>`, the main Service
	// `<name>` (the module pins general.serviceName to it), the discovery
	// Service `<name>-discovery`, the admin credentials Secret
	// `<name>-admin-password`, the Dashboards deployment/Service
	// `<name>-dashboards`.
	ClusterName string

	// HttpPort is the resolved HTTP API port (spec default 9200 — the
	// same default the CRD carries).
	HttpPort int

	// HttpEndpoint is the in-cluster URL of the HTTP API. Scheme is
	// ALWAYS https: see the comment on initializeLocals.
	HttpEndpoint string

	// AdminCredentialsSecretName is the operator-generated
	// `<name>-admin-password` Secret (fields username/password) — or
	// empty when spec.security.config replaces the operator bootstrap
	// (the operator then only mirrors user-provided credentials).
	AdminCredentialsSecretName string

	// Dashboards handles — empty when dashboards are not enabled.
	DashboardsServiceName string
	DashboardsEndpoint    string

	PortForwardCommand string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesopensearchv1alpha1.KubernetesOpenSearchIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesOpenSearch.String(),
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

	namespace := spec.Namespace.GetValue()
	clusterName := target.Metadata.Name

	httpPort := 9200
	if spec.HttpPort != nil && spec.GetHttpPort() > 0 {
		httpPort = int(spec.GetHttpPort())
	}

	// The HTTP endpoint scheme is https REGARDLESS of the security block:
	// the operator itself always talks https to the cluster
	// (pkg/builders/cluster.go URLForCluster returns
	// "https://<svc>.svc.<dns-base>:<port>" unconditionally, and the
	// node readiness probe curls https://localhost:<port>). With
	// spec.security absent the TLS reconciler generates nothing
	// (pkg/reconcilers/tls.go Reconcile returns early) and the
	// opensearchproject image's demo security configuration serves the
	// HTTP layer over TLS instead (pkg/reconcilers/securityconfig.go:
	// "Cluster is running with demo certificates").
	httpEndpoint := fmt.Sprintf("https://%s.%s.svc.cluster.local:%d",
		clusterName, namespace, httpPort)

	// The operator creates `<name>-admin-password` unconditionally
	// (pkg/reconcilers/cluster.go → builders.PasswordSecret), but with a
	// custom security config the operator's bootstrapped credentials do
	// not exist — the user's admin_credentials_secret is authoritative —
	// so the handle is exported empty.
	adminCredentialsSecretName := clusterName + "-admin-password"
	if spec.GetSecurity().GetConfig() != nil {
		adminCredentialsSecretName = ""
	}

	dashboardsServiceName := ""
	dashboardsEndpoint := ""
	if spec.GetDashboards().GetEnabled() {
		// The operator names the Dashboards Service
		// `<general.serviceName>-dashboards` (pkg/builders/dashboards.go)
		// and the module pins serviceName to the cluster name.
		dashboardsServiceName = clusterName + "-dashboards"
		dashboardsScheme := "http"
		if spec.GetDashboards().GetTls().GetEnable() {
			dashboardsScheme = "https"
		}
		dashboardsEndpoint = fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d",
			dashboardsScheme, dashboardsServiceName, namespace, vars.DashboardsPort)
	}

	return &Locals{
		KubernetesOpenSearch:       target,
		Spec:                       spec,
		Labels:                     labels,
		Namespace:                  namespace,
		ClusterName:                clusterName,
		HttpPort:                   httpPort,
		HttpEndpoint:               httpEndpoint,
		AdminCredentialsSecretName: adminCredentialsSecretName,
		DashboardsServiceName:      dashboardsServiceName,
		DashboardsEndpoint:         dashboardsEndpoint,
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
			clusterName, namespace, httpPort, httpPort),
	}
}
