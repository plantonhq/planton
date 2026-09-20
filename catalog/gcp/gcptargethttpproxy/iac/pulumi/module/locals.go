package module

import (
	gcptargethttpproxyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptargethttpproxy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpTargetHttpProxy *gcptargethttpproxyv1alpha1.GcpTargetHttpProxy

	// The cloud-side name defaults to metadata.name when the spec leaves
	// proxy_name empty — the same naming basis every kind uses.
	ProxyName string

	// The scope selector: a set spec.region builds the regional proxy, an
	// empty one the global proxy — the same switch the Terraform module's
	// count guards make.
	IsRegional bool
}

func initializeLocals(ctx *pulumi.Context, stackInput *gcptargethttpproxyv1alpha1.GcpTargetHttpProxyStackInput) *Locals {
	target := stackInput.Target

	proxyName := target.Spec.ProxyName
	if proxyName == "" {
		proxyName = target.Metadata.Name
	}

	return &Locals{
		GcpTargetHttpProxy: target,
		ProxyName:          proxyName,
		IsRegional:         target.Spec.Region != "",
	}
}
