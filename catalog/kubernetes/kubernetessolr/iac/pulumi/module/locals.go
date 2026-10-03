package module

import (
	"fmt"
	"strconv"

	kubernetessolrv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessolr/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesSolr *kubernetessolrv1alpha1.KubernetesSolr
	Spec           *kubernetessolrv1alpha1.KubernetesSolrSpec

	// Resource-identity labels stamped on the module-created objects
	// (namespace, SolrCloud). The Solr operator derives ITS objects'
	// identity from the SolrCloud name; these labels tie the family back
	// to the Planton resource.
	Labels map[string]string

	// Namespace the cluster lives in (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// ClusterName is metadata.name — the naming root the operator derives
	// every object from: StatefulSet `<name>-solrcloud`, common Service
	// `<name>-solrcloud-common`, basic-auth Secret
	// `<name>-solrcloud-basic-auth`, provided ZooKeeper client service
	// `<name>-solrcloud-zookeeper-client`.
	ClusterName string

	// CommonServiceName is the operator-created Service fronting all Solr
	// nodes (`<name>-solrcloud-common`).
	CommonServiceName string

	// TlsEnabled drives the endpoint scheme/port: the common service
	// listens on 80 without TLS and 443 with TLS.
	TlsEnabled bool

	// InternalEndpoint is the in-cluster base URL of the cluster through
	// the common service (scheme per TlsEnabled; 80/443 are the scheme
	// defaults so no port suffix is rendered).
	InternalEndpoint string

	// BasicAuthSecretName is the operator-generated credential Secret
	// (`<name>-solrcloud-basic-auth`) — populated only when basic auth is
	// enabled AND no user-provided basic_auth_secret is in play (the
	// operator only generates credentials in that case).
	BasicAuthSecretName string

	// ZookeeperConnectionString is what the cluster actually uses:
	// the external ensemble's connection string, or the provided
	// ensemble's operator-named client service
	// (`<name>-solrcloud-zookeeper-client:2181`) — plus the chroot when
	// it diverges from "/".
	ZookeeperConnectionString string

	// PortForwardCommand reaches the common service from a workstation.
	PortForwardCommand string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetessolrv1alpha1.KubernetesSolrIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesSolr.String(),
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
	commonServiceName := clusterName + "-solrcloud-common"

	tlsEnabled := spec.GetTls() != nil
	scheme := "http"
	commonServicePort := 80
	if tlsEnabled {
		scheme = "https"
		commonServicePort = 443
	}

	// Basic auth is the only operator-managed authentication type; the
	// operator generates `<name>-solrcloud-basic-auth` only when the user
	// did not bring their own credential Secret.
	basicAuthSecretName := ""
	if spec.GetSecurity().GetAuthenticationType() == "basic" &&
		spec.GetSecurity().GetBasicAuthSecret().GetValue() == "" {
		basicAuthSecretName = clusterName + "-solrcloud-basic-auth"
	}

	// The connection string the cluster uses. The provided arm (and the
	// empty default, which the operator treats as provided-with-defaults)
	// lands on the operator-named client service.
	zkConnection := clusterName + "-solrcloud-zookeeper-client:2181"
	zkChroot := zookeeperChroot(spec.GetZookeeper())
	if external := spec.GetZookeeper().GetExternal(); external != nil {
		zkConnection = external.GetConnectionString()
	}
	if zkChroot != "" && zkChroot != "/" {
		zkConnection = zkConnection + zkChroot
	}

	return &Locals{
		KubernetesSolr:      target,
		Spec:                spec,
		Labels:              labels,
		Namespace:           namespace,
		ClusterName:         clusterName,
		CommonServiceName:   commonServiceName,
		TlsEnabled:          tlsEnabled,
		BasicAuthSecretName: basicAuthSecretName,
		InternalEndpoint: fmt.Sprintf("%s://%s.%s.svc.cluster.local",
			scheme, commonServiceName, namespace),
		ZookeeperConnectionString: zkConnection,
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s 8983:%d",
			commonServiceName, namespace, commonServicePort),
	}
}

// zookeeperChroot resolves the chroot across the zookeeper oneof — "/" is
// the shared default of both arms (and of the operator).
func zookeeperChroot(zookeeper *kubernetessolrv1alpha1.KubernetesSolrZookeeper) string {
	chroot := "/"
	if external := zookeeper.GetExternal(); external != nil && external.GetChroot() != "" {
		chroot = external.GetChroot()
	}
	if provided := zookeeper.GetProvided(); provided != nil && provided.GetChroot() != "" {
		chroot = provided.GetChroot()
	}
	return chroot
}
