package module

import (
	"strconv"
	"strings"

	gcpgcsbucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpgcsbucket/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpLabels    map[string]string
	GcpGcsBucket *gcpgcsbucketv1.GcpGcsBucket
}

func initializeLocals(ctx *pulumi.Context, stackInput *gcpgcsbucketv1.GcpGcsBucketStackInput) *Locals {
	locals := &Locals{}

	locals.GcpGcsBucket = stackInput.Target

	target := stackInput.Target

	// Initialize standard GCP labels
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Spec.BucketName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpGcsBucket.String()),
	}

	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}

	// Merge user-provided labels from spec (user labels take precedence for conflicts)
	if target.Spec.GcpLabels != nil {
		for key, value := range target.Spec.GcpLabels {
			locals.GcpLabels[key] = value
		}
	}

	return locals
}
