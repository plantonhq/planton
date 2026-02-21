package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudrdsinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func account(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	instance *rds.Instance,
	acct *alicloudrdsinstancev1.AliCloudRdsAccount,
) error {
	args := &rds.RdsAccountArgs{
		DbInstanceId:    instance.ID(),
		AccountName:     pulumi.String(acct.AccountName),
		AccountPassword: pulumi.String(acct.AccountPassword),
		AccountType:     pulumi.String(accountType(acct)),
	}

	if acct.AccountDescription != "" {
		args.AccountDescription = pulumi.String(acct.AccountDescription)
	}

	createdAccount, err := rds.NewRdsAccount(ctx, acct.AccountName, args,
		pulumi.Provider(provider),
		pulumi.Parent(instance),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create RDS account %s", acct.AccountName)
	}

	for i, priv := range acct.Privileges {
		privName := fmt.Sprintf("%s-priv-%d", acct.AccountName, i)

		_, err := rds.NewAccountPrivilege(ctx, privName, &rds.AccountPrivilegeArgs{
			InstanceId:  instance.ID(),
			AccountName: createdAccount.AccountName,
			Privilege:   pulumi.String(privilege(priv)),
			DbNames:     pulumi.ToStringArray(priv.DatabaseNames),
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
