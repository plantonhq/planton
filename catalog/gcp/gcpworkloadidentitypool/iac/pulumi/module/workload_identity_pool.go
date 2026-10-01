package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// workloadIdentityPool provisions the Workload Identity Pool — the trust
// boundary external identities federate into.
//
// workload_identity_pool_id, project, and mode are immutable (the API rejects
// mode updates even though a plan may show one): changing any of them
// destroys and recreates the pool, invalidating every principal built from
// the old pool name. display_name, description, disabled, and the inline
// certificate/trust configs all update in place.
//
// GCP soft-deletes pools: after destroy, the pool remains DELETED for ~30
// days and its ID cannot be reused until permanent deletion. Unlike custom
// roles there is NO undelete-on-create — recreating a pool with a
// soft-deleted ID fails outright — so treat pool IDs as long-lived and prefer
// `disabled` for temporary shutoffs.
func workloadIdentityPool(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpWorkloadIdentityPool.Spec

	args := &iam.WorkloadIdentityPoolArgs{
		WorkloadIdentityPoolId: pulumi.String(spec.WorkloadIdentityPoolId),
		// Disabled is always sent: it is the documented kill switch, and an
		// explicit false keeps re-enable flows (true -> false) working.
		Disabled: pulumi.Bool(spec.Disabled),
	}

	// Omitted optionals stay unset (matching the Terraform module's null)
	// rather than being sent as empty strings.
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.String(spec.DisplayName)
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	// Mode defaults to FEDERATION_ONLY (via the spec default) — the token-
	// exchange federation mode. The API treats an unset mode identically, so
	// sending the default explicitly is behavior-neutral and keeps the diff
	// honest if a pool was created outside Planton in another mode.
	if spec.GetMode() != "" {
		args.Mode = pulumi.String(spec.GetMode())
	}

	// Honor the spec contract: an empty project_id falls back to the provider's
	// default project. Leaving Project unset lets the gcp provider resolve its
	// own project (configuration or the GOOGLE_PROJECT / GOOGLE_CLOUD_PROJECT
	// environment chain); an empty string would be sent verbatim and rejected.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	// The mTLS trust-domain surface: certificate issuance and foreign trust
	// bundles. Both optional; FEDERATION_ONLY pools normally set neither.
	if cert := spec.InlineCertificateIssuanceConfig; cert != nil {
		certArgs := &iam.WorkloadIdentityPoolInlineCertificateIssuanceConfigArgs{}
		// The provider declares ca_pools and use_default_shared_ca as a
		// two-sided ExactlyOneOf, and a bool set to FALSE still counts as
		// "set" — so each is sent only when it is the chosen source.
		if len(cert.CaPools) > 0 {
			caPools := pulumi.StringMap{}
			for region, caPool := range cert.CaPools {
				caPools[region] = pulumi.String(caPool.GetValue())
			}
			certArgs.CaPools = caPools
		}
		if cert.UseDefaultSharedCa {
			certArgs.UseDefaultSharedCa = pulumi.BoolPtr(true)
		}
		if cert.GetKeyAlgorithm() != "" {
			certArgs.KeyAlgorithm = pulumi.String(cert.GetKeyAlgorithm())
		}
		if cert.GetLifetime() != "" {
			certArgs.Lifetime = pulumi.String(cert.GetLifetime())
		}
		// 0 is not a legal percentage (the API floor is 50), so a zero value can
		// only mean "unset" — let the API default to 50.
		if cert.GetRotationWindowPercentage() != 0 {
			certArgs.RotationWindowPercentage = pulumi.Int(int(cert.GetRotationWindowPercentage()))
		}
		args.InlineCertificateIssuanceConfig = certArgs
	}

	if trust := spec.InlineTrustConfig; trust != nil && len(trust.AdditionalTrustBundles) > 0 {
		bundles := iam.WorkloadIdentityPoolInlineTrustConfigAdditionalTrustBundleArray{}
		for _, bundle := range trust.AdditionalTrustBundles {
			anchors := iam.WorkloadIdentityPoolInlineTrustConfigAdditionalTrustBundleTrustAnchorArray{}
			for _, anchor := range bundle.TrustAnchors {
				anchors = append(anchors, &iam.WorkloadIdentityPoolInlineTrustConfigAdditionalTrustBundleTrustAnchorArgs{
					PemCertificate: pulumi.String(anchor.PemCertificate),
				})
			}
			bundleArgs := &iam.WorkloadIdentityPoolInlineTrustConfigAdditionalTrustBundleArgs{
				TrustDomain:  pulumi.String(bundle.TrustDomain),
				TrustAnchors: anchors,
			}
			// Additionally trust the GCP-managed regional roots (managed
			// identity trust domains only). Sent only when true so plain
			// PEM-anchor bundles stay byte-identical to their pre-flag
			// shape — matching the Terraform module.
			if bundle.TrustDefaultSharedCa {
				bundleArgs.TrustDefaultSharedCa = pulumi.BoolPtr(true)
			}
			bundles = append(bundles, bundleArgs)
		}
		args.InlineTrustConfig = &iam.WorkloadIdentityPoolInlineTrustConfigArgs{
			AdditionalTrustBundles: bundles,
		}
	}

	// Which workloads may receive a managed identity. GCP applies these
	// through a second API call after the pool create; a failed apply can
	// leave a pool without its rules — re-apply converges.
	if len(spec.AttestationRules) > 0 {
		rules := iam.WorkloadIdentityPoolAttestationRuleArray{}
		for _, rule := range spec.AttestationRules {
			rules = append(rules, &iam.WorkloadIdentityPoolAttestationRuleArgs{
				GoogleCloudResource: pulumi.String(rule.GoogleCloudResource),
			})
		}
		args.AttestationRules = rules
	}

	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdPool, err := iam.NewWorkloadIdentityPool(ctx, "workload-identity-pool", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create workload identity pool")
	}

	ctx.Export(OpName, createdPool.Name)
	ctx.Export(OpWorkloadIdentityPoolId, createdPool.WorkloadIdentityPoolId)
	ctx.Export(OpState, createdPool.State)

	return nil
}
