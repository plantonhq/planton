package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudlogprojectv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudlogproject/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/log"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudlogprojectv1.AliCloudLogProjectStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudLogProject.Spec

	// Credentials are injected via environment variables by the runner
	// (ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET, etc.).
	// The Pulumi alicloud provider reads these automatically.
	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	project, err := log.NewProject(ctx, spec.ProjectName, &log.ProjectArgs{
		ProjectName:     pulumi.String(spec.ProjectName),
		Description:     pulumi.String(spec.Description),
		ResourceGroupId: optionalString(spec.ResourceGroupId),
		Tags:            pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create SLS project %s", spec.ProjectName)
	}

	logStoreNames := pulumi.StringMap{}

	for _, ls := range spec.LogStores {
		store, err := logStore(ctx, alicloudProvider, project, spec.ProjectName, ls)
		if err != nil {
			return err
		}

		if logStoreEnableIndex(ls) {
			if err := logStoreIndex(ctx, alicloudProvider, store, spec.ProjectName, ls.Name); err != nil {
				return err
			}
		}

		logStoreNames[ls.Name] = pulumi.String(ls.Name)
	}

	ctx.Export(OpProjectName, project.ProjectName)
	ctx.Export(OpProjectId, project.ID())
	ctx.Export(OpLogStoreNames, logStoreNames)

	return nil
}

func logStore(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	project *log.Project,
	projectName string,
	ls *alicloudlogprojectv1.AliCloudLogStore,
) (*log.Store, error) {
	resourceName := fmt.Sprintf("%s-%s", projectName, ls.Name)

	store, err := log.NewStore(ctx, resourceName, &log.StoreArgs{
		ProjectName:        pulumi.String(projectName),
		LogstoreName:       pulumi.String(ls.Name),
		RetentionPeriod:    pulumi.Int(logStoreRetentionDays(ls)),
		ShardCount:         pulumi.Int(logStoreShardCount(ls)),
		AutoSplit:          pulumi.Bool(logStoreAutoSplit(ls)),
		MaxSplitShardCount: pulumi.Int(logStoreMaxSplitShardCount(ls)),
		AppendMeta:         pulumi.Bool(logStoreAppendMeta(ls)),
	}, pulumi.Provider(provider), pulumi.Parent(project))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create log store %s", ls.Name)
	}

	return store, nil
}

func logStoreIndex(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	store *log.Store,
	projectName string,
	logstoreName string,
) error {
	resourceName := fmt.Sprintf("%s-%s-index", projectName, logstoreName)

	_, err := log.NewStoreIndex(ctx, resourceName, &log.StoreIndexArgs{
		Project:  pulumi.String(projectName),
		Logstore: pulumi.String(logstoreName),
		FullText: &log.StoreIndexFullTextArgs{
			CaseSensitive:  pulumi.Bool(false),
			IncludeChinese: pulumi.Bool(false),
			Token:          pulumi.String(`, '";=()[]{}?@&<>/:\n\t\r`),
		},
	}, pulumi.Provider(provider), pulumi.Parent(store))
	if err != nil {
		return errors.Wrapf(err, "failed to create index for log store %s", logstoreName)
	}

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
