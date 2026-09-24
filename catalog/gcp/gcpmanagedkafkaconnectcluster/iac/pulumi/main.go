package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkaconnectcluster/iac/pulumi/module"
	gcpmanagedkafkaconnectclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedkafkaconnectcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpmanagedkafkaconnectclusterv1alpha1.GcpManagedKafkaConnectClusterStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
