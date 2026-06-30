// Package main provides the Pulumi program entrypoint for AWS ECS Service deployment.
// This module creates and manages ECS services with associated resources.
// Binary releases are gzip-compressed to reduce download size (~75% smaller).
// Auto-release test: Single Pulumi module change triggers v{semver}-pulumi-awsecsservice-{YYYYMMDD}.{N}
package main

import (
	"github.com/pkg/errors"
	awsecsservicev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsecsservice/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsecsservice/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &awsecsservicev1.AwsEcsServiceStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
