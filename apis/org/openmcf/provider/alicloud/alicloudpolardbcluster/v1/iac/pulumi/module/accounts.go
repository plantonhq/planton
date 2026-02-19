package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudpolardbclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/polardb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func account(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	cluster *polardb.Cluster,
	acct *alicloudpolardbclusterv1.AlicloudPolardbAccount,
) error {
	args := &polardb.AccountArgs{
		DbClusterId:     cluster.ID(),
		AccountName:     pulumi.String(acct.AccountName),
		AccountPassword: pulumi.String(acct.AccountPassword),
		AccountType:     pulumi.String(accountType(acct)),
	}

	if acct.AccountDescription != "" {
		args.AccountDescription = pulumi.String(acct.AccountDescription)
	}

	createdAccount, err := polardb.NewAccount(ctx, acct.AccountName, args,
		pulumi.Provider(provider),
		pulumi.Parent(cluster),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create PolarDB account %s", acct.AccountName)
	}

	for i, priv := range acct.Privileges {
		privName := fmt.Sprintf("%s-priv-%d", acct.AccountName, i)

		_, err := polardb.NewAccountPrivilege(ctx, privName, &polardb.AccountPrivilegeArgs{
			DbClusterId:      cluster.ID(),
			AccountName:      createdAccount.AccountName,
			AccountPrivilege: pulumi.String(accountPrivilege(priv)),
			DbNames:          pulumi.ToStringArray(priv.DbNames),
		},
			pulumi.Provider(provider),
			pulumi.Parent(createdAccount),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create privilege %s for account %s", privName, acct.AccountName)
		}
	}

	return nil
}
