package module

import (
	civocertificatev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civocertificate/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed local variables for the certificate module.
type Locals struct {
	// CivoCertificate is the input resource specification
	CivoCertificate *civocertificatev1.CivoCertificate
	// Metadata is the cloud resource metadata (name, labels, description)
	Metadata *shared.CloudResourceMetadata
	// Labels are key-value pairs for resource tagging
	Labels map[string]string
}

// initializeLocals prepares local variables from the stack input.
func initializeLocals(ctx *pulumi.Context, stackInput *civocertificatev1.CivoCertificateStackInput) *Locals {
	locals := &Locals{
		CivoCertificate: stackInput.Target,
		Metadata:        stackInput.Target.Metadata,
		Labels:          make(map[string]string),
	}

	// Populate labels from metadata
	if locals.Metadata != nil && locals.Metadata.Labels != nil {
		for k, v := range locals.Metadata.Labels {
			locals.Labels[k] = v
		}
	}

	// Add tags from spec to labels
	if locals.CivoCertificate.Spec != nil && len(locals.CivoCertificate.Spec.Tags) > 0 {
		for _, tag := range locals.CivoCertificate.Spec.Tags {
			locals.Labels[tag] = "true"
		}
	}

	return locals
}
