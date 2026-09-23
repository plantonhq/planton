package module

import (
	"fmt"
	"strconv"

	kubernetesneo4jv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesneo4j/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesneo4jv1alpha1.KubernetesNeo4JSpec

	// Resource-identity labels stamped on the module-created satellites
	// (namespace, the auth Secret — never injected into the chart's own
	// resources; Helm owns those).
	Labels map[string]string

	// Namespace Neo4j installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// Neo4j servers coexist in one cluster (and Enterprise cluster members
	// are each their own release). The chart names its always-created
	// ClusterIP Service after the release (neo4j.fullname = Release.Name
	// when no name overrides are set), which is what makes ServiceName
	// below deterministic.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// The chart's neo4j.name — REQUIRED by the chart (its neo4j.name
	// helper fails the install when empty; nothing defaults it to the
	// release name). cluster_name when set (Enterprise members sharing it
	// form one cluster), else metadata.name. Controls the pod selector
	// labels, anti-affinity matching, the shared `<neo4j.name>-lb-neo4j`
	// LoadBalancer Service name, and clustering identity.
	Neo4JName string

	// Edition resolved to the chart/spec default (community) when unset.
	Edition string

	// Whether the module materializes the auth Secret: every arm but
	// existing_secret, which references a Secret the user owns. The chart
	// is never left to mint a credential nobody can find afterwards.
	CreateAuthSecret bool

	// Whether the module generates the admin password (auth absent). With
	// the password arm the declared value is used instead.
	GenerateAdminPassword bool

	// The admin password when the password arm is declared; empty when the
	// module generates it. Lands ONLY in the module-materialized Secret —
	// never in rendered chart values.
	AdminPassword string

	// Name of the Secret the chart's neo4j.passwordFromSecret points at:
	// the module-materialized "<metadata.name>-auth" (declared or
	// generated password), or the referenced existing Secret. Never empty.
	AuthSecretName string

	// Name of the main ClusterIP Service the chart always creates —
	// neo4j.fullname = the release name (templates/neo4j-svc.yaml names it
	// `{{ include "neo4j.fullname" . }}`; without fullnameOverride /
	// nameOverride that helper renders .Release.Name).
	ServiceName string

	// In-cluster bolt endpoint drivers connect to.
	BoltEndpoint string

	// In-cluster HTTP API / Browser endpoint.
	HttpEndpoint string

	// kubectl one-liner for reaching bolt from a workstation.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetesneo4jv1alpha1.KubernetesNeo4JStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesNeo4j.String(),
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

	edition := spec.GetEdition()
	if edition == "" {
		edition = "community"
	}

	neo4jName := spec.GetClusterName()
	if neo4jName == "" {
		neo4jName = target.Metadata.Name
	}

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name

	adminPassword := spec.GetAuth().GetPassword()
	existingSecret := spec.GetAuth().GetExistingSecret()
	createAuthSecret := existingSecret == ""
	generateAdminPassword := createAuthSecret && adminPassword == ""

	authSecretName := existingSecret
	if createAuthSecret {
		authSecretName = releaseName + "-auth"
	}

	// The chart's always-created ClusterIP Service is named after
	// neo4j.fullname, which is the release name here (no name overrides
	// are rendered).
	serviceName := releaseName

	return &Locals{
		Spec:                  spec,
		Labels:                labels,
		Namespace:             namespace,
		ReleaseName:           releaseName,
		ChartVersion:          chartVersion,
		Neo4JName:             neo4jName,
		Edition:               edition,
		CreateAuthSecret:      createAuthSecret,
		GenerateAdminPassword: generateAdminPassword,
		AdminPassword:         adminPassword,
		AuthSecretName:        authSecretName,
		ServiceName:           serviceName,
		BoltEndpoint: fmt.Sprintf("neo4j://%s.%s.svc.cluster.local:7687",
			serviceName, namespace),
		HttpEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:7474",
			serviceName, namespace),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s 7687:7687",
			serviceName, namespace),
	}
}
