package module

import (
	"fmt"
	"strconv"

	kuberneteszalandopostgresoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteszalandopostgresoperator/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals collect computed values that are reused across resources.
type Locals struct {
	KubernetesZalandoPostgresOperator *kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperator
	KubernetesLabels                  map[string]string

	// Namespace is the Kubernetes namespace to deploy to
	Namespace string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	BackupSecretName    string
	BackupConfigMapName string
}

// initializeLocals builds the Locals struct once and re‑uses it elsewhere.
func initializeLocals(_ *pulumi.Context, stackInput *kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorStackInput) *Locals {
	target := stackInput.Target

	kubeLabels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: "KubernetesZalandoPostgresOperator",
	}

	if target.Metadata.Id != "" {
		kubeLabels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		kubeLabels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		kubeLabels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Get namespace from spec (required field)
	namespace := target.Spec.Namespace.GetValue()
	if namespace == "" {
		namespace = vars.Namespace
	}

	return &Locals{
		KubernetesZalandoPostgresOperator: target,
		KubernetesLabels:                  kubeLabels,
		Namespace:                         namespace,

		// Computed resource names to avoid conflicts when multiple instances share a namespace
		// Format: {metadata.name}-{purpose}
		BackupSecretName:    fmt.Sprintf("%s-backup-credentials", target.Metadata.Name),
		BackupConfigMapName: fmt.Sprintf("%s-backup-config", target.Metadata.Name),
	}
}
