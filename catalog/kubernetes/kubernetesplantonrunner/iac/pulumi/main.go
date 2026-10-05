// Package main provides the Pulumi program entrypoint for the Kubernetes
// Planton Runner: a standing, outbound-only runner installed from the
// official planton-runner Helm chart that executes deploy and cloud
// operations from inside the cluster's network.
package main

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonrunner/iac/pulumi/module"
	kubernetesplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &kubernetesplantonrunnerv1alpha1.KubernetesPlantonRunnerIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return errors.Wrap(err, "failed to load iac-input")
		}

		return module.Resources(ctx, iacInput)
	})
}
