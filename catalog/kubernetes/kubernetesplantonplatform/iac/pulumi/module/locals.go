package module

import (
	"fmt"
	"strconv"

	kubernetesplantonplatformv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonplatform/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec

	// PlatformName is the PlantonPlatform CR name — this resource's
	// metadata.name, and the prefix of every object the operator creates
	// for the platform.
	PlatformName string

	// Resource-identity labels stamped on the module-created objects (the
	// namespace and the CR itself).
	Labels map[string]string

	// Namespace the platform lives in (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// GatewayLocalPort resolved to the operator default when unset — the
	// port the port_forward_command output advertises.
	GatewayLocalPort int

	// The names of the Secrets this module materializes for the database's
	// backup and recovery object stores (see object_store_secrets.go).
	// Deterministic from the platform name, so both engines and an import
	// recipe derive them blind; the CR names them to the operator, which
	// never creates them itself.
	BackupCredentialsSecretName   string
	RecoveryCredentialsSecretName string
	BackupEndpointCaSecretName    string
	RecoveryEndpointCaSecretName  string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesPlantonPlatform.String(),
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

	gatewayLocalPort := vars.GatewayDefaultPort
	if spec.GetGateway() != nil && spec.GetGateway().LocalPort != nil {
		gatewayLocalPort = int(spec.GetGateway().GetLocalPort())
	}

	postgresClusterName := target.Metadata.Name + vars.PostgresClusterSuffix

	return &Locals{
		Spec:             spec,
		PlatformName:     target.Metadata.Name,
		Labels:           labels,
		Namespace:        spec.Namespace.GetValue(),
		GatewayLocalPort: gatewayLocalPort,

		BackupCredentialsSecretName:   postgresClusterName + vars.BackupCredentialsSecretSuffix,
		RecoveryCredentialsSecretName: postgresClusterName + vars.RecoveryCredentialsSecretSuffix,
		BackupEndpointCaSecretName:    postgresClusterName + vars.BackupEndpointCaSecretSuffix,
		RecoveryEndpointCaSecretName:  postgresClusterName + vars.RecoveryEndpointCaSecretSuffix,
	}
}

// portForwardCommand renders the exact command that opens the platform's
// door on this machine — twin of the Terraform module's output expression.
func (l *Locals) portForwardCommand() string {
	return fmt.Sprintf("kubectl port-forward -n %s svc/%s%s %d:%d",
		l.Namespace, l.PlatformName, vars.GatewayServiceSuffix, l.GatewayLocalPort, vars.GatewayServicePort)
}

// setupCodeCommand renders the exact command that reads the first-run
// setup code — twin of the Terraform module's output expression.
func (l *Locals) setupCodeCommand() string {
	return fmt.Sprintf("kubectl -n %s get secret %s%s -o jsonpath='{.data.%s}' | base64 -d",
		l.Namespace, l.PlatformName, vars.SetupCodeSecretSuffix, vars.SetupCodeSecretKey)
}
