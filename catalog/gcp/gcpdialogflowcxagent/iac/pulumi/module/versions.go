package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/diagflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// versionsAndEnvironments creates one flow version per spec.versions[]
// entry, keyed "{flow_id}/{display name}" with the empty flow_id resolved to
// the start flow, then one environment per spec.environments[] entry whose
// version configs resolve to those versions or compose an outside version's
// path under this agent. Flow paths are composed from the agent's name
// because a same-manifest agent's id does not exist before the apply.
func versionsAndEnvironments(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, createdAgent *diagflow.CxAgent) error {
	spec := locals.GcpDialogflowCxAgent.Spec
	resourceName := locals.GcpDialogflowCxAgent.Metadata.Name
	agentName := createdAgent.ID().ToStringOutput()

	createdVersions := map[string]*diagflow.CxVersion{}
	versionNames := pulumi.StringArray{}
	for _, version := range spec.Versions {
		flowId := flowIdOrStart(version.FlowId)
		key := fmt.Sprintf("%s/%s", flowId, version.DisplayName)

		// A version snapshots its flow at creation and waits for the flow's
		// model to train.
		created, err := diagflow.NewCxVersion(ctx,
			fmt.Sprintf("%s-version-%s", resourceName, key),
			&diagflow.CxVersionArgs{
				Parent:         pulumi.Sprintf("%s/flows/%s", agentName, flowId),
				DisplayName:    pulumi.String(version.DisplayName),
				Description:    optionalString(version.Description),
				DeletionPolicy: deletionPolicy(locals),
			},
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdAgent))
		if err != nil {
			return errors.Wrapf(err, "failed to create flow version %s", key)
		}
		createdVersions[key] = created
		versionNames = append(versionNames, created.ID().ToStringOutput())
	}

	environmentNames := pulumi.StringArray{}
	for _, environment := range spec.Environments {
		configs := diagflow.CxEnvironmentVersionConfigArray{}
		// Resources an environment waits on: the declared versions it pins.
		pinned := []pulumi.Resource{}
		for _, config := range environment.VersionConfigs {
			flowId := flowIdOrStart(config.FlowId)
			var version pulumi.StringInput
			if config.Version != "" {
				declared, ok := createdVersions[fmt.Sprintf("%s/%s", flowId, config.Version)]
				if !ok {
					return errors.Errorf("environment %s names version %q of flow %s, which is not declared in versions",
						environment.DisplayName, config.Version, flowId)
				}
				version = declared.ID().ToStringOutput()
				pinned = append(pinned, declared)
			} else {
				version = pulumi.Sprintf("%s/flows/%s/versions/%s", agentName, flowId, config.VersionId)
			}
			configs = append(configs, &diagflow.CxEnvironmentVersionConfigArgs{Version: version})
		}

		created, err := diagflow.NewCxEnvironment(ctx,
			fmt.Sprintf("%s-environment-%s", resourceName, environment.DisplayName),
			&diagflow.CxEnvironmentArgs{
				Parent:         agentName,
				DisplayName:    pulumi.String(environment.DisplayName),
				Description:    optionalString(environment.Description),
				VersionConfigs: configs,
				DeletionPolicy: deletionPolicy(locals),
			},
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdAgent),
			pulumi.DependsOn(pinned))
		if err != nil {
			return errors.Wrapf(err, "failed to create environment %s", environment.DisplayName)
		}
		environmentNames = append(environmentNames, created.ID().ToStringOutput())
	}

	ctx.Export(OpVersionNames, versionNames.ToStringArrayOutput())
	ctx.Export(OpEnvironmentNames, environmentNames.ToStringArrayOutput())
	return nil
}
