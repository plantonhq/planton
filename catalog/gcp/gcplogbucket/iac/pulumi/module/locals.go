package module

import (
	gcplogbucketv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcplogbucket/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The spec's default values, applied by the module (both engines apply the
// same ones so behavior is identical regardless of engine).
const (
	// The bucket location when the spec leaves it empty.
	defaultLocation = "global"

	// GCP's own default retention, sent explicitly so the spec default is
	// what the API applies rather than a silently different server-side
	// state.
	defaultRetentionDays = 30
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs. The bucket resources
// carry no labels argument, so there is no platform-label merge here.
type Locals struct {
	GcpLogBucket *gcplogbucketv1alpha1.GcpLogBucket

	// Exactly one of these is true — the scope arm the module creates
	// (empty scope means a project bucket in the provider's default
	// project, resolved through client config).
	IsProjectBucket bool
	IsFolderBucket  bool
	IsOrgBucket     bool
	IsBillingBucket bool

	// The resolved location and retention (spec defaults applied).
	Location      string
	RetentionDays int
}

func initializeLocals(_ *pulumi.Context, iacInput *gcplogbucketv1alpha1.GcpLogBucketIacInput) *Locals {
	target := iacInput.Target
	scope := target.Spec.Scope

	locals := &Locals{
		GcpLogBucket:  target,
		Location:      target.Spec.Location,
		RetentionDays: int(target.Spec.RetentionDays),
	}

	if locals.Location == "" {
		locals.Location = defaultLocation
	}
	if locals.RetentionDays == 0 {
		locals.RetentionDays = defaultRetentionDays
	}

	switch {
	case scope != nil && scope.FolderId.GetValue() != "":
		locals.IsFolderBucket = true
	case scope != nil && scope.OrganizationId != "":
		locals.IsOrgBucket = true
	case scope != nil && scope.BillingAccount != "":
		locals.IsBillingBucket = true
	default:
		locals.IsProjectBucket = true
	}

	return locals
}
