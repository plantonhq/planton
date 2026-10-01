package envsecrets

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/secretmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Ref is what one environment variable reads: the secret the resource owns
// and the exact version holding the current value. Pinning the version, not
// "latest", is what makes a changed value a change to the resource itself --
// rotation is a deploy the platform can see, never instances silently
// diverging.
type Ref struct {
	// Secret is the secret's short id.
	Secret pulumi.StringOutput
	// Version is the version number.
	Version pulumi.StringOutput
	// Project is the project the secret lives in, resolved even when the
	// placement left it to the provider's default.
	Project pulumi.StringOutput
	// Name is the version's full resource name,
	// projects/<project>/secrets/<id>/versions/<n> -- the pointer a runtime
	// without a native secret field receives as the variable's value.
	Name pulumi.StringOutput
}

// Stored is the outcome of Store: each variable's reference, and the grants
// the workload must be created after (a runtime checks or exercises its
// identity's access to every referenced secret when it deploys, so a deploy
// that races its grant fails).
type Stored struct {
	Refs   map[Key]Ref
	Grants []pulumi.Resource
}

// Store creates, for every variable, a Secret Manager secret replicated only
// in the placement's regions, a version holding the value, and a
// secretAccessor grant on that secret alone for the runtime identity. It
// enables the Secret Manager API first, never disabling it on destroy (one
// resource's teardown must not switch an API off for the project).
//
// A version replaced by a new value is created before the old one is
// deleted, and the workload is updated between the two, so the running
// workload keeps a readable version until its update lands.
//
// The Pulumi resource names are part of every existing stack's state: the
// API is "<kind>-secretmanager.googleapis.com", and the secret, its version,
// and its grant are each named by the secret id (the latter two parented to
// the secret). Renaming any of them makes the engine replace the secret,
// and a replacement under the same id fails.
func Store(ctx *pulumi.Context, placement Placement, variables []Variable, provider pulumi.ProviderResource) (*Stored, error) {
	stored := &Stored{Refs: map[Key]Ref{}}
	if len(variables) == 0 {
		return stored, nil
	}
	ids, err := SecretIDs(placement, variables)
	if err != nil {
		return nil, err
	}
	opts := []pulumi.ResourceOption{pulumi.Provider(provider)}

	var project pulumi.StringPtrInput
	if placement.Project != "" {
		project = pulumi.String(placement.Project)
	}

	projectNumber := ""
	if placement.RuntimeServiceAccount == "" {
		var projectID *string
		if placement.Project != "" {
			projectID = &placement.Project
		}
		found, err := organizations.LookupProject(ctx, &organizations.LookupProjectArgs{ProjectId: projectID}, pulumi.Provider(provider))
		if err != nil {
			return nil, errors.Wrap(err, "failed to look up the project number for the default runtime identity the secret grant names")
		}
		projectNumber = found.Number
	}
	member := RuntimeMember(placement, projectNumber)

	api, err := projects.NewService(ctx, APIResourceName(placement), &projects.ServiceArgs{
		Project:                  project,
		Service:                  pulumi.String("secretmanager.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(false),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable secretmanager.googleapis.com")
	}

	replicas := secretmanager.SecretReplicationUserManagedReplicaArray{}
	for _, region := range placement.ReplicaRegions {
		replicas = append(replicas, &secretmanager.SecretReplicationUserManagedReplicaArgs{
			Location: pulumi.String(region),
		})
	}

	for _, variable := range variables {
		key := KeyOf(placement, variable)
		id := ids[key]

		secret, err := secretmanager.NewSecret(ctx, id, &secretmanager.SecretArgs{
			Project:  project,
			SecretId: pulumi.String(id),
			Labels:   pulumi.ToStringMap(placement.Labels),
			Replication: &secretmanager.SecretReplicationArgs{
				UserManaged: &secretmanager.SecretReplicationUserManagedArgs{Replicas: replicas},
			},
		}, append(opts, pulumi.DependsOn([]pulumi.Resource{api}))...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create Secret Manager secret %s", id)
		}

		version, err := secretmanager.NewSecretVersion(ctx, id, &secretmanager.SecretVersionArgs{
			Secret:     secret.ID(),
			SecretData: pulumi.ToSecret(pulumi.String(variable.Value)).(pulumi.StringOutput),
		}, append(opts, pulumi.Parent(secret))...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to store the value of %s", id)
		}

		grant, err := secretmanager.NewSecretIamMember(ctx, id, &secretmanager.SecretIamMemberArgs{
			Project:  secret.Project,
			SecretId: secret.SecretId,
			Role:     pulumi.String("roles/secretmanager.secretAccessor"),
			Member:   pulumi.String(member),
		}, append(opts, pulumi.Parent(secret))...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to grant the runtime identity access to %s", id)
		}

		stored.Refs[key] = Ref{Secret: secret.SecretId, Version: version.Version, Project: secret.Project, Name: version.Name}
		stored.Grants = append(stored.Grants, grant)
	}
	return stored, nil
}

// APIResourceName is the Pulumi name of the resource that enables the
// Secret Manager API for one placement's kind.
func APIResourceName(placement Placement) string {
	return placement.Kind + "-secretmanager.googleapis.com"
}
