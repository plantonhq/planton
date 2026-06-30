package module

import (
	kubernetesissuerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesissuer/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesIssuer *kubernetesissuerv1.KubernetesIssuer
	Namespace        string
	IssuerName       string
	Labels           map[string]string

	// Issuer-type flags -- exactly one is true (enforced by proto CEL validation).
	IsCa         bool
	IsSelfSigned bool

	// Only populated when IsCa == true.
	CaSecretName string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesissuerv1.KubernetesIssuerStackInput) *Locals {
	locals := &Locals{}

	target := stackInput.Target
	spec := target.Spec

	locals.KubernetesIssuer = target
	locals.Namespace = spec.Namespace.GetValue()
	locals.IssuerName = target.Metadata.Name
	locals.Labels = target.Metadata.Labels

	// Detect which issuer_type oneof branch is set.
	switch spec.IssuerType.(type) {
	case *kubernetesissuerv1.KubernetesIssuerSpec_Ca:
		locals.IsCa = true
		locals.CaSecretName = spec.GetCa().GetCaSecretName().GetValue()
	case *kubernetesissuerv1.KubernetesIssuerSpec_SelfSigned:
		locals.IsSelfSigned = true
	}

	return locals
}
