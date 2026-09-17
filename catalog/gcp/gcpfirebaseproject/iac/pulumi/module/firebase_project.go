package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/firebase"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// firebaseProject enables Firebase on the project and provisions the
// project-level Firebase surface the spec declares: the default Cloud
// Storage for Firebase bucket and App Check enforcement.
//
// Firebase enablement is a ONE-WAY project singleton. The provider's create
// GETs the project first and ADOPTS an already-enabled project (no error,
// no second :addFirebase); its delete drops the resource from state and
// leaves the project enabled -- Google offers no way to remove Firebase
// from a project. That is why the enablement carries no deletion_policy:
// the spec's deletion_policy governs only the composed default bucket and
// App Check configurations below.
//
// API enablement is module plumbing: firebase.googleapis.com is what
// :addFirebase talks to, fcm.googleapis.com is what a control plane sends
// push through (declared explicitly, never assumed from :addFirebase's side
// effects), and the storage / App Check APIs are enabled exactly when the
// spec composes their resources. disable_on_destroy stays false everywhere:
// destroying this resource must never switch off APIs other resources in
// the project depend on.
func firebaseProject(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFirebaseProject.Spec

	enableService := func(service string) (*projects.Service, error) {
		serviceArgs := &projects.ServiceArgs{
			Service:                  pulumi.String(service),
			DisableDependentServices: pulumi.BoolPtr(true),
		}
		if spec.ProjectId.GetValue() != "" {
			serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		created, err := projects.NewService(ctx, "firebase-"+service, serviceArgs, pulumi.Provider(gcpProvider))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to enable %s api", service)
		}
		return created, nil
	}

	firebaseApi, err := enableService("firebase.googleapis.com")
	if err != nil {
		return err
	}
	fcmApi, err := enableService("fcm.googleapis.com")
	if err != nil {
		return err
	}

	// The enablement itself: one input of substance (the project).
	projectArgs := &firebase.ProjectArgs{}
	if spec.ProjectId.GetValue() != "" {
		projectArgs.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}
	createdProject, err := firebase.NewProject(ctx, "firebase", projectArgs,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{firebaseApi, fcmApi}))
	if err != nil {
		return errors.Wrap(err, "failed to enable firebase on the project")
	}

	// The default Cloud Storage for Firebase bucket, when the spec asks for
	// one. A project holds one at a time; destroy under DELETE unlinks and
	// deletes it, so it can be declared again later. Needs the pay-as-you-go
	// plan.
	// The App Check configurations below wait for the bucket when the spec
	// declares one (see appCheckDeps).
	var defaultBucket *firebase.StorageDefaultBucket
	if spec.DefaultStorageLocation != "" {
		storageApi, err := enableService("firebasestorage.googleapis.com")
		if err != nil {
			return err
		}
		bucketArgs := &firebase.StorageDefaultBucketArgs{
			Location: pulumi.String(spec.DefaultStorageLocation),
		}
		if spec.ProjectId.GetValue() != "" {
			bucketArgs.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
		}
		if spec.DeletionPolicy != "" {
			bucketArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		defaultBucket, err = firebase.NewStorageDefaultBucket(ctx, "default-bucket", bucketArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProject, storageApi}))
		if err != nil {
			return errors.Wrap(err, "failed to create the default storage bucket")
		}
	}

	// App Check: per-service enforcement and per-resource overrides. Each
	// configuration is its own resource keyed by the immutable service id
	// (and target resource), so plans stay stable as list order changes.
	if spec.AppCheck != nil && (len(spec.AppCheck.ServiceConfigs) > 0 || len(spec.AppCheck.ResourcePolicies) > 0) {
		appCheckApi, err := enableService("firebaseappcheck.googleapis.com")
		if err != nil {
			return err
		}
		// A service's enforcement can be configured only once that service is
		// set up on the project (live-verified: the API answers 400 "Cloud
		// Firestore is not yet set up" otherwise). The one setup this module
		// itself performs is the default bucket, so the configurations wait
		// for it -- a manifest that declares the bucket AND Storage
		// enforcement applies in one pass.
		appCheckDeps := []pulumi.Resource{createdProject, appCheckApi}
		if defaultBucket != nil {
			appCheckDeps = append(appCheckDeps, defaultBucket)
		}
		deps := pulumi.DependsOn(appCheckDeps)

		for _, sc := range spec.AppCheck.ServiceConfigs {
			args := &firebase.AppCheckServiceConfigArgs{
				ServiceId: pulumi.String(sc.ServiceId),
			}
			// An empty enforcement_mode is OFF -- Google's unset state -- and
			// is sent as unset so the API records exactly that.
			if sc.EnforcementMode != "" {
				args.EnforcementMode = pulumi.StringPtr(sc.EnforcementMode)
			}
			if spec.ProjectId.GetValue() != "" {
				args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
			}
			if spec.DeletionPolicy != "" {
				args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
			}
			if _, err := firebase.NewAppCheckServiceConfig(ctx, "app-check-"+sc.ServiceId, args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrapf(err, "failed to configure app check for %s", sc.ServiceId)
			}
		}

		for i, rp := range spec.AppCheck.ResourcePolicies {
			args := &firebase.AppCheckResourcePolicyArgs{
				ServiceId:      pulumi.String(rp.ServiceId),
				TargetResource: pulumi.String(rp.TargetResource),
			}
			if rp.EnforcementMode != "" {
				args.EnforcementMode = pulumi.StringPtr(rp.EnforcementMode)
			}
			if spec.ProjectId.GetValue() != "" {
				args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
			}
			if spec.DeletionPolicy != "" {
				args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
			}
			if _, err := firebase.NewAppCheckResourcePolicy(ctx, fmt.Sprintf("app-check-policy-%d", i), args,
				pulumi.Provider(gcpProvider), deps); err != nil {
				return errors.Wrapf(err, "failed to configure app check resource policy for %s", rp.TargetResource)
			}
		}
	}

	// The Admin SDK configuration -- the project-level values the Firebase
	// SDKs are initialised with. Read AFTER the enablement: the lookup takes
	// the created resource's project output, so the invoke waits on it
	// (and on an already-enabled project the values are simply read).
	// Every value is conditionally present on Google's side (RTDB, default
	// bucket, finalized location), so each degrades to "" rather than
	// failing the stack.
	adminSdk := firebase.GetAdminSdkConfigOutput(ctx, firebase.GetAdminSdkConfigOutputArgs{
		Project: createdProject.Project,
	}, pulumi.Provider(gcpProvider))

	ctx.Export(OpProjectId, createdProject.Project)
	ctx.Export(OpProjectNumber, createdProject.ProjectNumber)
	ctx.Export(OpDisplayName, createdProject.DisplayName)
	ctx.Export(OpDatabaseUrl, adminSdk.DatabaseUrl())
	ctx.Export(OpStorageBucket, adminSdk.StorageBucket())
	ctx.Export(OpLocationId, adminSdk.LocationId())

	return nil
}
