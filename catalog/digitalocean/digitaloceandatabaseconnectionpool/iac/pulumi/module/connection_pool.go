package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// connectionPool provisions the PgBouncer pool and exports its outputs.
// EVERY argument is create-only upstream (the provider registers no update
// path), so any change replaces the pool and drops its live connections.
func connectionPool(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.DatabaseConnectionPool, error) {
	spec := locals.DigitalOceanDatabaseConnectionPool.Spec

	poolArgs := &digitalocean.DatabaseConnectionPoolArgs{
		// References are resolved to the literal cluster UUID before the
		// module runs.
		ClusterId: pulumi.String(spec.Cluster.GetValue()),
		Name:      pulumi.String(spec.PoolName),
		Mode:      pulumi.String(spec.Mode),
		Size:      pulumi.Int(int(spec.Size)),
		DbName:    pulumi.String(spec.DbName),
	}

	// Omitted user = inbound-user pool (clients bring their own
	// credentials) -- the safer default for shared pools.
	if spec.User != "" {
		poolArgs.User = pulumi.StringPtr(spec.User)
	}

	createdPool, err := digitalocean.NewDatabaseConnectionPool(
		ctx,
		"pool",
		poolArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean database connection pool")
	}

	ctx.Export(OpClusterId, createdPool.ClusterId)
	ctx.Export(OpPoolName, createdPool.Name)
	ctx.Export(OpHost, createdPool.Host)
	ctx.Export(OpPrivateHost, createdPool.PrivateHost)
	ctx.Export(OpPort, createdPool.Port)
	ctx.Export(OpUri, pulumi.ToSecret(createdPool.Uri))
	ctx.Export(OpPrivateUri, pulumi.ToSecret(createdPool.PrivateUri))
	ctx.Export(OpPassword, pulumi.ToSecret(createdPool.Password))

	return createdPool, nil
}
