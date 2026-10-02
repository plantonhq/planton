package module

import (
	gcptagvaluev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptagvalue/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the short name the value resource is built from. A tag value
// carries no labels, so there is no platform-label merge here.
type Locals struct {
	GcpTagValue *gcptagvaluev1alpha1.GcpTagValue

	// The value's short name: the spec's short_name, or metadata.name when
	// the spec leaves it empty -- the same naming basis every kind uses.
	ShortName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcptagvaluev1alpha1.GcpTagValueStackInput) *Locals {
	target := stackInput.Target

	shortName := target.Spec.ShortName
	if shortName == "" {
		shortName = target.Metadata.Name
	}

	return &Locals{
		GcpTagValue: target,
		ShortName:   shortName,
	}
}
