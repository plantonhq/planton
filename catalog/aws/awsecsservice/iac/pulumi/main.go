// Package main provides the Pulumi program entrypoint for AWS ECS Service deployment.
// This module creates and manages ECS services with associated resources.
// Binary releases are gzip-compressed to reduce download size (~75% smaller).
// Auto-release test: Single Pulumi module change triggers v{semver}-pulumi-awsecsservice-{YYYYMMDD}.{N}
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/aws/awsecsservice/iac/pulumi/module"
	awsecsservicev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecsservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &awsecsservicev1alpha1.AwsEcsServiceIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
