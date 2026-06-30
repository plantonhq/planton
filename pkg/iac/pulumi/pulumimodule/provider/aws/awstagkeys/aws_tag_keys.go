package awstagkeys

import (
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/labels/labelkeys"
)

var (
	Name         = "Name"
	Resource     = labelkeys.WithDomainPrefix("resource")
	Organization = labelkeys.WithDomainPrefix("organization")
	Environment  = labelkeys.WithDomainPrefix("environment")
	ResourceKind = labelkeys.WithDomainPrefix("resource-kind")
	ResourceId   = labelkeys.WithDomainPrefix("resource-id")
)
