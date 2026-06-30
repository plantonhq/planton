package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// function provisions a serverless Function via DigitalOcean App Platform
// and exports its ID + public HTTPS endpoint.
func function(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.App, error) {

	// 1. Translate the env maps into the provider‑specific structure.
	var functionEnvs digitalocean.AppSpecFunctionEnvArray

	// Add regular environment variables
	for k, v := range locals.DigitalOceanFunction.Spec.EnvironmentVariables {
		functionEnvs = append(functionEnvs, digitalocean.AppSpecFunctionEnvArgs{
			Key:   pulumi.String(k),
			Value: pulumi.String(v),
		})
	}

	// Add secret environment variables
	for k, v := range locals.DigitalOceanFunction.Spec.SecretEnvironmentVariables {
		functionEnvs = append(functionEnvs, digitalocean.AppSpecFunctionEnvArgs{
			Key:   pulumi.String(k),
			Value: pulumi.String(v),
			Type:  pulumi.String("SECRET"),
		})
	}

	// 2. Build a single Function component inside an AppSpec.
	functionDef := digitalocean.AppSpecFunctionArgs{
		Name: pulumi.String(locals.DigitalOceanFunction.Spec.FunctionName),
		Envs: functionEnvs,
	}

	appSpec := digitalocean.AppSpecArgs{
		Name:      pulumi.String(locals.DigitalOceanFunction.Metadata.Name),
		Region:    pulumi.String(locals.DigitalOceanFunction.Spec.Region.String()),
		Functions: digitalocean.AppSpecFunctionArray{functionDef},
	}

	appArgs := &digitalocean.AppArgs{
		Spec: appSpec,
	}

	// 3. Create the App (which deploys the Function).
	createdApp, err := digitalocean.NewApp(
		ctx,
		"function",
		appArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean app for function")
	}

	// 4. Export stack outputs.
	ctx.Export(OpFunctionId, createdApp.ID())
	ctx.Export(OpHttpsEndpoint, createdApp.LiveUrl)

	return createdApp, nil
}
