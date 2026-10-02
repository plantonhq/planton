package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dataConnector enables the Discovery Engine API and sets up the connector
// (`google_discovery_engine_data_connector`), which creates the collection
// and one data store per entity. The collection ids, the source, the
// location, the KMS key, the static-IP switch, and each entity's name are
// immutable; the schedule, the parameters, the modes, and the action side
// update in place.
//
// Send posture (parity with the Terraform module): optional strings,
// lists, and blocks are sent only when set so Google's defaults stay in
// charge (sync mode PERIODIC, the incremental interval three hours). The
// parameter maps carry Secret Manager resource names, never secret
// material, by Google's design.
func dataConnector(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiSearchDataConnector.Spec

	// Enable the Discovery Engine API first so a fresh project works on the
	// first deploy. DisableOnDestroy stays false: tearing down one
	// connector must never disable the API for everything else in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("discoveryengine.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpvsdc-discoveryengine.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable discoveryengine.googleapis.com api")
	}

	args := &discoveryengine.DataConnectorArgs{
		CollectionId:            pulumi.String(locals.CollectionId),
		CollectionDisplayName:   pulumi.String(locals.CollectionDisplayName),
		DataSource:              pulumi.String(spec.DataSource),
		Location:                pulumi.String(spec.Location),
		RefreshInterval:         pulumi.String(spec.RefreshInterval),
		IncrementalSyncDisabled: pulumi.BoolPtr(spec.IncrementalSyncDisabled),
		AutoRunDisabled:         pulumi.BoolPtr(spec.AutoRunDisabled),
		StaticIpEnabled:         pulumi.BoolPtr(spec.StaticIpEnabled),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.DataSourceVersion != nil {
		args.DataSourceVersion = pulumi.Int(int(spec.GetDataSourceVersion()))
	}
	// Exactly one of the two parameter forms (a spec rule).
	if len(spec.Params) > 0 {
		args.Params = pulumi.ToStringMap(spec.Params)
	}
	if spec.JsonParams != "" {
		args.JsonParams = pulumi.String(spec.JsonParams)
	}
	if spec.IncrementalRefreshInterval != "" {
		args.IncrementalRefreshInterval = pulumi.String(spec.IncrementalRefreshInterval)
	}
	if spec.SyncMode != "" {
		args.SyncMode = pulumi.String(spec.SyncMode)
	}
	if len(spec.ConnectorModes) > 0 {
		args.ConnectorModes = pulumi.ToStringArray(spec.ConnectorModes)
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.String(spec.KmsKeyName.GetValue())
	}

	if len(spec.Entities) > 0 {
		entities := discoveryengine.DataConnectorEntityArray{}
		for _, entity := range spec.Entities {
			entityArgs := &discoveryengine.DataConnectorEntityArgs{
				EntityName: pulumi.String(entity.EntityName),
			}
			if entity.Params != "" {
				entityArgs.Params = pulumi.String(entity.Params)
			}
			if len(entity.KeyPropertyMappings) > 0 {
				entityArgs.KeyPropertyMappings = pulumi.ToStringMap(entity.KeyPropertyMappings)
			}
			entities = append(entities, entityArgs)
		}
		args.Entities = entities
	}

	if len(spec.DestinationConfigs) > 0 {
		destinationConfigs := discoveryengine.DataConnectorDestinationConfigArray{}
		for _, cfg := range spec.DestinationConfigs {
			cfgArgs := &discoveryengine.DataConnectorDestinationConfigArgs{}
			if cfg.Key != "" {
				cfgArgs.Key = pulumi.String(cfg.Key)
			}
			if cfg.Params != "" {
				cfgArgs.Params = pulumi.String(cfg.Params)
			}
			if len(cfg.Destinations) > 0 {
				destinations := discoveryengine.DataConnectorDestinationConfigDestinationArray{}
				for _, destination := range cfg.Destinations {
					destinationArgs := &discoveryengine.DataConnectorDestinationConfigDestinationArgs{}
					if destination.Host != "" {
						destinationArgs.Host = pulumi.String(destination.Host)
					}
					if destination.Port != nil {
						destinationArgs.Port = pulumi.Int(int(destination.GetPort()))
					}
					destinations = append(destinations, destinationArgs)
				}
				cfgArgs.Destinations = destinations
			}
			destinationConfigs = append(destinationConfigs, cfgArgs)
		}
		args.DestinationConfigs = destinationConfigs
	}

	if action := spec.ActionConfig; action != nil {
		actionArgs := &discoveryengine.DataConnectorActionConfigArgs{
			CreateBapConnection: pulumi.BoolPtr(action.CreateBapConnection),
		}
		if len(action.ActionParams) > 0 {
			actionArgs.ActionParams = pulumi.ToStringMap(action.ActionParams)
		}
		args.ActionConfig = actionArgs
	}
	if bap := spec.BapConfig; bap != nil {
		bapArgs := &discoveryengine.DataConnectorBapConfigArgs{}
		if len(bap.SupportedConnectorModes) > 0 {
			bapArgs.SupportedConnectorModes = pulumi.ToStringArray(bap.SupportedConnectorModes)
		}
		if len(bap.EnabledActions) > 0 {
			bapArgs.EnabledActions = pulumi.ToStringArray(bap.EnabledActions)
		}
		args.BapConfig = bapArgs
	}

	// Client-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := discoveryengine.NewDataConnector(ctx,
		locals.GcpVertexAiSearchDataConnector.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create discovery engine data connector")
	}

	// One data store per entity, in manifest order; Google fills each
	// entity's data_store after setup.
	entityDataStores := created.Entities.ApplyT(func(entities []discoveryengine.DataConnectorEntity) []string {
		names := []string{}
		for _, entity := range entities {
			if entity.DataStore != nil {
				names = append(names, *entity.DataStore)
			}
		}
		return names
	}).(pulumi.StringArrayOutput)

	ctx.Export(OpName, created.Name)
	ctx.Export(OpCollectionId, created.CollectionId)
	ctx.Export(OpLocation, created.Location)
	ctx.Export(OpState, created.State)
	ctx.Export(OpEntityDataStores, entityDataStores)
	ctx.Export(OpStaticIpAddresses, created.StaticIpAddresses)
	ctx.Export(OpPrivateConnectivityProjectId, created.PrivateConnectivityProjectId)
	return nil
}
