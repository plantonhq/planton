package module

import (
	"fmt"
	"strconv"

	kubernetesneo4jv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesneo4j/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
)

// Locals struct mirrors the "locals" concept from Terraform,
// providing a consolidated place for derived values.
type Locals struct {
	// The top-level resource from the user’s manifest (apiVersion, kind, metadata, spec).
	KubernetesNeo4J *kubernetesneo4jv1.KubernetesNeo4J

	// Standard labels used for all created resources.
	Labels map[string]string

	// The namespace in which everything will be deployed.
	Namespace string

	// Pod selector labels for identifying the pods that run Neo4j.
	Neo4jPodSelectorLabels map[string]string

	// Calculated hostname for external Ingress usage.
	IngressExternalHostname string

	// The name of the Kubernetes Service for the Neo4j instance.
	KubeServiceName string

	// The fully qualified domain name (in-cluster) for the Neo4j Service.
	KubeServiceFqdn string

	// A convenient port-forwarding command for local development.
	KubePortForwardCommand string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// The Neo4j Helm chart creates a secret named "<release>-auth" for the password
	PasswordSecretName string
}

// initializeLocals populates Locals from the stack input and
// exports some fields to the Pulumi stack outputs.
func initializeLocals(
	ctx *pulumi.Context,
	stackInput *kubernetesneo4jv1.KubernetesNeo4JStackInput,
) *Locals {
	// Initialize the Locals struct.
	locals := &Locals{
		KubernetesNeo4J: stackInput.Target,
		Labels:          map[string]string{},
	}

	target := stackInput.Target

	// Basic labels that identify this resource in the cluster.
	locals.Labels[kuberneteslabelkeys.Resource] = strconv.FormatBool(true)
	locals.Labels[kuberneteslabelkeys.ResourceName] = target.Metadata.Name
	locals.Labels[kuberneteslabelkeys.ResourceKind] = cloudresourcekind.CloudResourceKind_KubernetesNeo4j.String()

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Define pod selector labels that Helm or raw K8s resources might use.
	locals.Neo4jPodSelectorLabels = map[string]string{
		"app.kubernetes.io/name":      "neo4j",
		"app.kubernetes.io/instance":  target.Metadata.Name,
		"app.kubernetes.io/component": "primary",
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// The Neo4j Helm chart creates a secret named "<release>-auth" for the password
	locals.PasswordSecretName = fmt.Sprintf("%s-auth", target.Metadata.Name)

	// Construct a service name and FQDN.
	locals.KubeServiceName = fmt.Sprintf("%s-neo4j", target.Metadata.Name)
	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))
	ctx.Export(OpBoltUriKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	// Provide a default port-forward command for local debugging (web interface is 7474).
	// For Bolt (7687), you could adapt this as needed.
	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 7474:7474",
		locals.Namespace, locals.KubeServiceName)
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	// If ingress is enabled, use the hostname directly
	if target.Spec.Ingress != nil &&
		target.Spec.Ingress.Enabled &&
		target.Spec.Ingress.Hostname != "" {
		locals.IngressExternalHostname = target.Spec.Ingress.Hostname
		ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))
	}

	return locals
}
