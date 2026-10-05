package module

import (
	"fmt"
	"strconv"

	kubernetesmysqlv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmysql/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesMysql *kubernetesmysqlv1alpha1.KubernetesMysql
	Spec            *kubernetesmysqlv1alpha1.KubernetesMysqlSpec

	// Resource-identity labels stamped on every module-created object
	// (namespace, the PerconaXtraDBCluster resource, credential Secrets).
	// The operator derives ITS objects' identity from the cluster name;
	// these labels tie the whole family back to the Planton resource.
	Labels map[string]string

	// Namespace the cluster lives in (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// ClusterName is metadata.name — the naming root the operator derives
	// every object from: pods `<name>-pxc-0..N`, services `<name>-pxc` /
	// `<name>-haproxy` / `<name>-proxysql`, the system-users Secret
	// `<name>-secrets`.
	ClusterName string

	// IsHaproxy is true unless the spec's proxy oneof chose ProxySQL —
	// an absent proxy block means the HAProxy default posture.
	IsHaproxy bool

	// Traffic service names (operator-created; exported as outputs).
	// ReplicasServiceName is empty unless HAProxy is in play with its
	// replicas Service enabled.
	PrimaryServiceName  string
	ReplicasServiceName string

	// KubeEndpoint is the in-cluster FQDN of the write path — the
	// connection host applications use.
	KubeEndpoint string

	// RootPasswordSecretName is the operator-managed system-users Secret
	// (`<name>-secrets`); its `root` key is the root password.
	RootPasswordSecretName string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesmysqlv1alpha1.KubernetesMysqlIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesMysql.String(),
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

	// The proxy oneof: ProxySQL only when explicitly chosen; anything
	// else (haproxy arm or an absent proxy block) is the HAProxy posture.
	isHaproxy := spec.GetProxy().GetProxysql() == nil

	primaryServiceName := clusterName + "-proxysql"
	if isHaproxy {
		primaryServiceName = clusterName + "-haproxy"
	}

	// The replicas (read) Service exists only for HAProxy, and only while
	// enabled — the upstream default is enabled, so only an explicit
	// false in the spec turns it off.
	replicasServiceName := ""
	if isHaproxy {
		replicasServiceName = clusterName + "-haproxy-replicas"
		if exposeReplicas := spec.GetProxy().GetHaproxy().GetExposeReplicas(); exposeReplicas != nil &&
			exposeReplicas.Enabled != nil && !*exposeReplicas.Enabled {
			replicasServiceName = ""
		}
	}

	return &Locals{
		KubernetesMysql:     target,
		Spec:                spec,
		Labels:              labels,
		Namespace:           namespace,
		ClusterName:         clusterName,
		IsHaproxy:           isHaproxy,
		PrimaryServiceName:  primaryServiceName,
		ReplicasServiceName: replicasServiceName,
		KubeEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:3306",
			primaryServiceName, namespace),
		RootPasswordSecretName: clusterName + "-secrets",
	}
}
