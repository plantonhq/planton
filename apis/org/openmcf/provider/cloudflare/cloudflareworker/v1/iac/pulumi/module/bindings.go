package module

import (
	cloudflareworkerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareworker/v1"
	cloudfl "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildBindings flattens the spec's grouped, type-specific binding lists into the
// provider's single discriminated bindings array. Sensitive values (secret_text)
// arrive resolved via StringValueOrRef.GetValue().
func buildBindings(spec *cloudflareworkerv1.CloudflareWorkerSpec) cloudfl.WorkersScriptBindingArray {
	var bindings cloudfl.WorkersScriptBindingArray

	for k, v := range spec.Vars {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name: pulumi.String(k),
			Type: pulumi.String("plain_text"),
			Text: pulumi.String(v),
		})
	}
	for _, b := range spec.Secrets {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name: pulumi.String(b.Name),
			Type: pulumi.String("secret_text"),
			Text: pulumi.String(b.Value.GetValue()),
		})
	}
	for _, b := range spec.KvNamespaces {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name:        pulumi.String(b.Name),
			Type:        pulumi.String("kv_namespace"),
			NamespaceId: pulumi.String(b.NamespaceId.GetValue()),
		})
	}
	for _, b := range spec.R2Buckets {
		args := cloudfl.WorkersScriptBindingArgs{
			Name:       pulumi.String(b.Name),
			Type:       pulumi.String("r2_bucket"),
			BucketName: pulumi.String(b.BucketName.GetValue()),
		}
		if b.Jurisdiction != "" {
			args.Jurisdiction = pulumi.String(b.Jurisdiction)
		}
		bindings = append(bindings, args)
	}
	for _, b := range spec.D1Databases {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name: pulumi.String(b.Name),
			Type: pulumi.String("d1"),
			Id:   pulumi.String(b.DatabaseId.GetValue()),
		})
	}
	for _, b := range spec.HyperdriveConfigs {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name: pulumi.String(b.Name),
			Type: pulumi.String("hyperdrive"),
			Id:   pulumi.String(b.ConfigId.GetValue()),
		})
	}
	for _, b := range spec.Services {
		args := cloudfl.WorkersScriptBindingArgs{
			Name:    pulumi.String(b.Name),
			Type:    pulumi.String("service"),
			Service: pulumi.String(b.Service.GetValue()),
		}
		if b.Environment != "" {
			args.Environment = pulumi.String(b.Environment)
		}
		if b.Entrypoint != "" {
			args.Entrypoint = pulumi.String(b.Entrypoint)
		}
		bindings = append(bindings, args)
	}
	for _, b := range spec.Queues {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name:      pulumi.String(b.Name),
			Type:      pulumi.String("queue"),
			QueueName: pulumi.String(b.QueueName.GetValue()),
		})
	}
	for _, b := range spec.DurableObjects {
		args := cloudfl.WorkersScriptBindingArgs{
			Name:      pulumi.String(b.Name),
			Type:      pulumi.String("durable_object_namespace"),
			ClassName: pulumi.String(b.ClassName),
		}
		if b.ScriptName != "" {
			args.ScriptName = pulumi.String(b.ScriptName)
		}
		if b.Environment != "" {
			args.Environment = pulumi.String(b.Environment)
		}
		bindings = append(bindings, args)
	}
	for _, b := range spec.AnalyticsEngineDatasets {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name:    pulumi.String(b.Name),
			Type:    pulumi.String("analytics_engine"),
			Dataset: pulumi.String(b.Dataset),
		})
	}
	for _, b := range spec.VectorizeIndexes {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name:      pulumi.String(b.Name),
			Type:      pulumi.String("vectorize"),
			IndexName: pulumi.String(b.IndexName),
		})
	}
	for _, b := range spec.Ai {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name: pulumi.String(b.Name),
			Type: pulumi.String("ai"),
		})
	}
	for _, b := range spec.VersionMetadata {
		bindings = append(bindings, cloudfl.WorkersScriptBindingArgs{
			Name: pulumi.String(b.Name),
			Type: pulumi.String("version_metadata"),
		})
	}

	return bindings
}
