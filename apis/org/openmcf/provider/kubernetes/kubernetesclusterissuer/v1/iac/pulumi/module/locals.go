package module

import (
	"fmt"

	kubernetesclusterissuerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesclusterissuer/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesClusterIssuer  *kubernetesclusterissuerv1.KubernetesClusterIssuer
	CertManagerNamespace     string
	DnsDomain                string
	CloudflareSecretName     string
	AcmeAccountKeySecretName string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesclusterissuerv1.KubernetesClusterIssuerStackInput) *Locals {
	locals := &Locals{}
	locals.KubernetesClusterIssuer = stackInput.Target

	target := stackInput.Target

	locals.CertManagerNamespace = target.Spec.CertManagerNamespace.GetValue()
	locals.DnsDomain = target.Spec.DnsDomain
	locals.CloudflareSecretName = fmt.Sprintf("%s-cloudflare-credentials", target.Metadata.Name)
	locals.AcmeAccountKeySecretName = fmt.Sprintf("letsencrypt-%s-account-key", target.Spec.DnsDomain)

	return locals
}
