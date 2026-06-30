package module

import (
	"github.com/pkg/errors"
	alicloudpolardbclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudpolardbcluster/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/polardb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func database(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	cluster *polardb.Cluster,
	dbType string,
	db *alicloudpolardbclusterv1.AliCloudPolardbDatabase,
) (*polardb.Database, error) {
	args := &polardb.DatabaseArgs{
		DbClusterId: cluster.ID(),
		DbName:      pulumi.String(db.DbName),
	}

	charset := db.CharacterSetName
	if charset == "" {
		charset = defaultCharacterSet(dbType)
	}
	args.CharacterSetName = pulumi.String(charset)

	if db.DbDescription != "" {
		args.DbDescription = pulumi.String(db.DbDescription)
	}

	if db.Collate != "" {
		args.Collate = pulumi.String(db.Collate)
	}

	if db.Ctype != "" {
		args.Ctype = pulumi.String(db.Ctype)
	}

	created, err := polardb.NewDatabase(ctx, db.DbName, args,
		pulumi.Provider(provider),
		pulumi.Parent(cluster),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create PolarDB database %s", db.DbName)
	}

	return created, nil
}

func defaultCharacterSet(dbType string) string {
	switch dbType {
	case "MySQL":
		return "utf8"
	case "PostgreSQL", "Oracle":
		return "UTF8"
	default:
		return "utf8"
	}
}
