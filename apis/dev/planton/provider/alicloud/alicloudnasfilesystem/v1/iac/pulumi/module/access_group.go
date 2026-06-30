package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudnasfilesystemv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudnasfilesystem/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/nas"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// accessGroup creates a NAS access group with the specified access rules.
// Returns the access group name for use in mount target configuration.
func accessGroup(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	resourceName string,
	fsType string,
	rules []*alicloudnasfilesystemv1.AliCloudNasAccessRule,
) (pulumi.StringOutput, error) {
	agName := fmt.Sprintf("%s-ag", resourceName)

	ag, err := nas.NewAccessGroup(ctx, agName, &nas.AccessGroupArgs{
		AccessGroupName: pulumi.String(agName),
		AccessGroupType: pulumi.String("Vpc"),
		FileSystemType:  pulumi.String(fsType),
		Description:     pulumi.Sprintf("Access group for NAS file system %s", resourceName),
	}, pulumi.Provider(provider))
	if err != nil {
		return pulumi.StringOutput{}, errors.Wrapf(err, "failed to create access group %s", agName)
	}

	for i, rule := range rules {
		if err := accessRule(ctx, provider, ag, agName, fsType, i, rule); err != nil {
			return pulumi.StringOutput{}, err
		}
	}

	return ag.AccessGroupName, nil
}

func accessRule(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	ag *nas.AccessGroup,
	agName string,
	fsType string,
	index int,
	rule *alicloudnasfilesystemv1.AliCloudNasAccessRule,
) error {
	ruleName := fmt.Sprintf("%s-rule-%d", agName, index)

	args := &nas.AccessRuleArgs{
		AccessGroupName: ag.AccessGroupName,
		SourceCidrIp:    pulumi.String(rule.SourceCidrIp),
		FileSystemType:  pulumi.String(fsType),
	}

	if rule.RwAccessType != nil && *rule.RwAccessType != "" {
		args.RwAccessType = pulumi.String(*rule.RwAccessType)
	} else {
		args.RwAccessType = pulumi.String("RDWR")
	}

	if rule.UserAccessType != nil && *rule.UserAccessType != "" {
		args.UserAccessType = pulumi.String(*rule.UserAccessType)
	} else {
		args.UserAccessType = pulumi.String("no_squash")
	}

	if rule.Priority != nil && *rule.Priority > 0 {
		args.Priority = pulumi.Int(int(*rule.Priority))
	} else {
		args.Priority = pulumi.Int(1)
	}

	_, err := nas.NewAccessRule(ctx, ruleName, args,
		pulumi.Provider(provider),
		pulumi.Parent(ag),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create access rule %s", ruleName)
	}

	return nil
}
