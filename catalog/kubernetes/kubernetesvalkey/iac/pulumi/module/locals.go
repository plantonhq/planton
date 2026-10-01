package module

import (
	"fmt"
	"strconv"

	kubernetesvalkeyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesvalkey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesvalkeyv1alpha1.KubernetesValkeySpec

	// Resource-identity labels: stamped on the module-created satellites
	// (namespace, the auth Secret) and on every Valkey pod through the
	// chart's own podLabels -- so a log line, a metric or an alert from it
	// names its organization and environment. The chart's selector is its
	// own fixed labels (valkey.selectorLabels); these never reach it.
	Labels map[string]string

	// Namespace Valkey installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// Valkey instances coexist in one cluster (one cache per team, one
	// store per workload), so each manifest gets its own release. The
	// chart's fullname is pinned to the same value (values.go), which is
	// what makes the rendered Service names below deterministic.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Whether ACL authentication was declared. Gates the auth Secret, the
	// chart's auth values, and the credential outputs.
	AuthEnabled bool

	// Deterministic name of the module-materialized password Secret
	// ("<metadata.name>-auth", one key per ACL username). The chart
	// consumes it via auth.usersExistingSecret, so passwords never appear
	// in rendered chart values.
	AuthSecretName string

	// Whether the replication block was declared (primary/replica
	// topology). Gates the chart's replica values, the read/headless
	// Service outputs, and the PodDisruptionBudget rendering.
	ReplicationEnabled bool

	// Whether the replication-mode read Service renders — replication
	// declared AND read_service.enabled not explicitly false (the chart
	// default is true).
	ReadServiceEnabled bool

	// The write Service port, resolved to the chart default (6379) when
	// the service block or its port is unset — feeds the endpoint outputs.
	ServicePort int32

	// In-cluster endpoint of the write Service.
	KubeEndpoint string

	// kubectl one-liner for reaching the store from a workstation.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetesvalkeyv1alpha1.KubernetesValkeyStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesValkey.String(),
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

	replicationEnabled := spec.GetReplication() != nil

	// The chart renders the read Service by default in replication mode;
	// only an explicit enabled=false suppresses it.
	readServiceEnabled := replicationEnabled
	if rs := spec.GetReplication().GetReadService(); rs != nil && rs.Enabled != nil {
		readServiceEnabled = rs.GetEnabled()
	}

	servicePort := int32(6379)
	if spec.GetService() != nil && spec.GetService().Port != nil {
		servicePort = spec.GetService().GetPort()
	}

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name

	return &Locals{
		Spec:               spec,
		Labels:             labels,
		Namespace:          namespace,
		ReleaseName:        releaseName,
		ChartVersion:       chartVersion,
		AuthEnabled:        spec.GetAuth() != nil,
		AuthSecretName:     releaseName + "-auth",
		ReplicationEnabled: replicationEnabled,
		ReadServiceEnabled: readServiceEnabled,
		ServicePort:        servicePort,
		KubeEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			releaseName, namespace, servicePort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
			releaseName, namespace, servicePort, servicePort),
	}
}
