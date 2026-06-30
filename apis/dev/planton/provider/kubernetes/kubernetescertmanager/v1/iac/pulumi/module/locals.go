package module

import (
	kubernetescertmanagerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetescertmanager/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesCertManager *kubernetescertmanagerv1.KubernetesCertManager
	Namespace             string
	ServiceAccountName    string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetescertmanagerv1.KubernetesCertManagerStackInput) *Locals {
	locals := &Locals{}
	locals.KubernetesCertManager = stackInput.Target

	target := stackInput.Target

	locals.Namespace = target.Spec.Namespace.GetValue()
	locals.ServiceAccountName = target.Metadata.Name

	return locals
}
