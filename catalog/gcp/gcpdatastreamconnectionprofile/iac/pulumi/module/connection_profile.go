package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/datastream"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// secret marks a password, key, or certificate secret in Pulumi state and
// returns nil for an empty value so the provider omits it.
func secret(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.ToSecret(pulumi.String(value)).(pulumi.StringOutput)
}

// optionalString returns nil for an empty value so the provider omits it.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.String(value)
}

// optionalPort returns nil for zero so the provider's engine default port
// applies -- the Terraform module's null-for-zero rule.
func optionalPort(port int32) pulumi.IntPtrInput {
	if port <= 0 {
		return nil
	}
	return pulumi.Int(int(port))
}

// connectionProfile creates the Datastream connection profile with exactly
// one profile type (the spec's rule) and at most one connectivity option.
func connectionProfile(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDatastreamConnectionProfile.Spec
	resourceName := locals.GcpDatastreamConnectionProfile.Metadata.Name

	// Enable the Datastream API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one profile must
	// never disable the API for every other profile and stream in the
	// project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("datastream.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpdscp-datastream.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable datastream.googleapis.com api")
	}

	args := &datastream.ConnectionProfileArgs{
		Location:            pulumi.String(spec.Location),
		ConnectionProfileId: pulumi.String(locals.ConnectionProfileId),
		DisplayName:         pulumi.String(locals.DisplayName),
		Labels:              pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.CreateWithoutValidation {
		args.CreateWithoutValidation = pulumi.BoolPtr(true)
	}

	// Google's BigQuery profile has no settings; the spec's bool emits the
	// empty marker block.
	if spec.BigqueryProfile {
		args.BigqueryProfile = &datastream.ConnectionProfileBigqueryProfileArgs{}
	}

	if gcs := spec.GcsProfile; gcs != nil {
		args.GcsProfile = &datastream.ConnectionProfileGcsProfileArgs{
			Bucket:   pulumi.String(gcs.Bucket.GetValue()),
			RootPath: optionalString(gcs.RootPath),
		}
	}

	if mysql := spec.MysqlProfile; mysql != nil {
		mysqlArgs := &datastream.ConnectionProfileMysqlProfileArgs{
			Hostname:                    pulumi.String(mysql.Hostname.GetValue()),
			Port:                        optionalPort(mysql.Port),
			Username:                    pulumi.String(mysql.Username),
			Password:                    secret(mysql.Password),
			SecretManagerStoredPassword: optionalString(mysql.SecretManagerStoredPassword.GetValue()),
		}
		// Declared (even empty) means sent: an empty ssl_config turns TLS on
		// without certificate verification -- the Terraform module's rule.
		if ssl := mysql.SslConfig; ssl != nil {
			mysqlArgs.SslConfig = &datastream.ConnectionProfileMysqlProfileSslConfigArgs{
				CaCertificate:     secret(ssl.CaCertificate),
				ClientCertificate: secret(ssl.ClientCertificate),
				ClientKey:         secret(ssl.ClientKey),
			}
		}
		args.MysqlProfile = mysqlArgs
	}

	if postgresql := spec.PostgresqlProfile; postgresql != nil {
		postgresqlArgs := &datastream.ConnectionProfilePostgresqlProfileArgs{
			Hostname:                    pulumi.String(postgresql.Hostname.GetValue()),
			Port:                        optionalPort(postgresql.Port),
			Username:                    pulumi.String(postgresql.Username),
			Password:                    secret(postgresql.Password),
			SecretManagerStoredPassword: optionalString(postgresql.SecretManagerStoredPassword.GetValue()),
			Database:                    pulumi.String(postgresql.Database),
		}
		if ssl := postgresql.SslConfig; ssl != nil {
			sslArgs := &datastream.ConnectionProfilePostgresqlProfileSslConfigArgs{}
			if verification := ssl.ServerVerification; verification != nil {
				sslArgs.ServerVerification = &datastream.ConnectionProfilePostgresqlProfileSslConfigServerVerificationArgs{
					CaCertificate: pulumi.ToSecret(pulumi.String(verification.CaCertificate)).(pulumi.StringOutput),
				}
			}
			if verification := ssl.ServerAndClientVerification; verification != nil {
				sslArgs.ServerAndClientVerification = &datastream.ConnectionProfilePostgresqlProfileSslConfigServerAndClientVerificationArgs{
					CaCertificate:     pulumi.ToSecret(pulumi.String(verification.CaCertificate)).(pulumi.StringOutput),
					ClientCertificate: pulumi.ToSecret(pulumi.String(verification.ClientCertificate)).(pulumi.StringOutput),
					ClientKey:         pulumi.ToSecret(pulumi.String(verification.ClientKey)).(pulumi.StringOutput),
				}
			}
			postgresqlArgs.SslConfig = sslArgs
		}
		args.PostgresqlProfile = postgresqlArgs
	}

	if oracle := spec.OracleProfile; oracle != nil {
		oracleArgs := &datastream.ConnectionProfileOracleProfileArgs{
			Hostname:                    pulumi.String(oracle.Hostname),
			Port:                        optionalPort(oracle.Port),
			Username:                    pulumi.String(oracle.Username),
			Password:                    secret(oracle.Password),
			SecretManagerStoredPassword: optionalString(oracle.SecretManagerStoredPassword.GetValue()),
			DatabaseService:             pulumi.String(oracle.DatabaseService),
		}
		if len(oracle.ConnectionAttributes) > 0 {
			oracleArgs.ConnectionAttributes = pulumi.ToStringMap(oracle.ConnectionAttributes)
		}
		args.OracleProfile = oracleArgs
	}

	if sqlServer := spec.SqlServerProfile; sqlServer != nil {
		args.SqlServerProfile = &datastream.ConnectionProfileSqlServerProfileArgs{
			Hostname:                    pulumi.String(sqlServer.Hostname.GetValue()),
			Port:                        optionalPort(sqlServer.Port),
			Username:                    pulumi.String(sqlServer.Username),
			Password:                    secret(sqlServer.Password),
			SecretManagerStoredPassword: optionalString(sqlServer.SecretManagerStoredPassword.GetValue()),
			Database:                    pulumi.String(sqlServer.Database),
		}
	}

	if mongodb := spec.MongodbProfile; mongodb != nil {
		hosts := datastream.ConnectionProfileMongodbProfileHostAddressArray{}
		for _, host := range mongodb.HostAddresses {
			hosts = append(hosts, &datastream.ConnectionProfileMongodbProfileHostAddressArgs{
				Hostname: pulumi.String(host.Hostname),
				Port:     optionalPort(host.Port),
			})
		}
		mongodbArgs := &datastream.ConnectionProfileMongodbProfileArgs{
			HostAddresses:               hosts,
			Username:                    pulumi.String(mongodb.Username),
			Password:                    secret(mongodb.Password),
			SecretManagerStoredPassword: optionalString(mongodb.SecretManagerStoredPassword.GetValue()),
			ReplicaSet:                  optionalString(mongodb.ReplicaSet),
		}
		if len(mongodb.AdditionalOptions) > 0 {
			mongodbArgs.AdditionalOptions = pulumi.ToStringMap(mongodb.AdditionalOptions)
		}
		// Google's SRV block has no settings; the spec's bool emits the
		// empty marker block.
		if mongodb.SrvConnectionFormat {
			mongodbArgs.SrvConnectionFormat = &datastream.ConnectionProfileMongodbProfileSrvConnectionFormatArgs{}
		}
		if standard := mongodb.StandardConnectionFormat; standard != nil {
			standardArgs := &datastream.ConnectionProfileMongodbProfileStandardConnectionFormatArgs{}
			if standard.DirectConnection {
				standardArgs.DirectConnection = pulumi.BoolPtr(true)
			}
			mongodbArgs.StandardConnectionFormat = standardArgs
		}
		if ssl := mongodb.SslConfig; ssl != nil {
			mongodbArgs.SslConfig = &datastream.ConnectionProfileMongodbProfileSslConfigArgs{
				CaCertificate:                secret(ssl.CaCertificate),
				ClientCertificate:            secret(ssl.ClientCertificate),
				ClientKey:                    secret(ssl.ClientKey),
				SecretManagerStoredClientKey: optionalString(ssl.SecretManagerStoredClientKey.GetValue()),
			}
		}
		args.MongodbProfile = mongodbArgs
	}

	// The spec lifts the private-connectivity block's one leaf; the block is
	// sent only when a private connection is named.
	if privateConnection := spec.PrivateConnection.GetValue(); privateConnection != "" {
		args.PrivateConnectivity = &datastream.ConnectionProfilePrivateConnectivityArgs{
			PrivateConnection: pulumi.String(privateConnection),
		}
	}

	if ssh := spec.ForwardSshConnectivity; ssh != nil {
		args.ForwardSshConnectivity = &datastream.ConnectionProfileForwardSshConnectivityArgs{
			Hostname:   pulumi.String(ssh.Hostname),
			Port:       optionalPort(ssh.Port),
			Username:   pulumi.String(ssh.Username),
			Password:   secret(ssh.Password),
			PrivateKey: secret(ssh.PrivateKey),
		}
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := datastream.NewConnectionProfile(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create datastream connection profile")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpConnectionProfileId, created.ConnectionProfileId)
	return nil
}
