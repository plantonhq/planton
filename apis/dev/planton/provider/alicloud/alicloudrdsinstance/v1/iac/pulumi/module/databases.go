package module

import (
	"github.com/pkg/errors"
	alicloudrdsinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudrdsinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func database(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	instance *rds.Instance,
	engine string,
	db *alicloudrdsinstancev1.AliCloudRdsDatabase,
) (*rds.Database, error) {
	args := &rds.DatabaseArgs{
		InstanceId:   instance.ID(),
		DataBaseName: pulumi.String(db.Name),
	}

	charset := db.CharacterSet
	if charset == "" {
		charset = defaultCharacterSet(engine)
	}
	args.CharacterSet = pulumi.String(charset)

	if db.Description != "" {
		args.Description = pulumi.String(db.Description)
	}

	created, err := rds.NewDatabase(ctx, db.Name, args,
		pulumi.Provider(provider),
		pulumi.Parent(instance),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create database %s", db.Name)
	}

	return created, nil
}

func defaultCharacterSet(engine string) string {
	switch engine {
	case "MySQL", "MariaDB":
		return "utf8mb4"
	case "PostgreSQL", "PPAS":
		return "UTF8"
	case "SQLServer":
		return "Chinese_PRC_CI_AS"
	default:
		return "utf8"
	}
}
