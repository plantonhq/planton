package module

import (
	"fmt"
	"strconv"

	kubernetessignozv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessignoz/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetessignozv1alpha1.KubernetesSignozSpec

	// Resource-identity labels stamped on the module-created namespace —
	// never injected into the chart's own resources; Helm owns those.
	Labels map[string]string

	// Namespace SigNoz installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// SigNoz instances can coexist. fullnameOverride pins every chart
	// child name to it.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// ---- exported composition handles (see outputs.go) -----------------
	// The ClickHouse* handles are passthroughs of the DECLARED connection
	// — this component installs no ClickHouse; downstream kinds
	// referencing them compose against the same store SigNoz uses.
	Service                      string
	KubeEndpoint                 string
	PortForwardCommand           string
	OtelCollectorService         string
	OtlpGrpcEndpoint             string
	OtlpHttpEndpoint             string
	ClickHouseEndpoint           string
	ClickHouseUsername           string
	ClickHousePasswordSecretName string
	ClickHousePasswordSecretKey  string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetessignozv1alpha1.KubernetesSignozIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesSignoz.String(),
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

	// The signoz Service carries the fullname itself; the collector
	// Service appends the component name — both pinned via
	// fullnameOverride.
	service := releaseName
	kubeEndpoint := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service, namespace, vars.ServerHttpPort)
	portForward := fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
		service, namespace, vars.ServerHttpPort, vars.ServerHttpPort)

	collectorService := fmt.Sprintf("%s-otel-collector", releaseName)
	otlpGrpc := fmt.Sprintf("%s.%s.svc.cluster.local:%d", collectorService, namespace, vars.OtlpGrpcPort)
	otlpHttp := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", collectorService, namespace, vars.OtlpHttpPort)

	// The declared connection, mirrored into the exported handles.
	clickhouse := spec.GetClickhouse()
	tcpPort := int32(vars.ClickHouseTcpPort)
	if clickhouse.TcpPort != nil {
		tcpPort = clickhouse.GetTcpPort()
	}
	clickHouseEndpoint := fmt.Sprintf("%s:%d", clickhouse.Host.GetValue(), tcpPort)

	return &Locals{
		Spec:                         spec,
		Labels:                       labels,
		Namespace:                    namespace,
		ReleaseName:                  releaseName,
		ChartVersion:                 chartVersion,
		Service:                      service,
		KubeEndpoint:                 kubeEndpoint,
		PortForwardCommand:           portForward,
		OtelCollectorService:         collectorService,
		OtlpGrpcEndpoint:             otlpGrpc,
		OtlpHttpEndpoint:             otlpHttp,
		ClickHouseEndpoint:           clickHouseEndpoint,
		ClickHouseUsername:           clickhouse.Username,
		ClickHousePasswordSecretName: clickhouse.PasswordSecret.SecretName.GetValue(),
		ClickHousePasswordSecretKey:  clickhouse.PasswordSecret.SecretKey,
	}
}
