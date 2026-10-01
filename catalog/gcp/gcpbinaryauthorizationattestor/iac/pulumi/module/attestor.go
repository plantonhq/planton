package module

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	gcpbinaryauthorizationattestorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbinaryauthorizationattestor/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/binaryauthorization"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/containeranalysis"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/kms"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var keyVersionPattern = regexp.MustCompile(`^(.+)/cryptoKeyVersions/([0-9]+)$`)

// attestor creates the attestor, its own Artifact Analysis note when the
// spec asks for one, and the grant Google requires for the attestor to
// read attestations under that note.
func attestor(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	target := locals.GcpBinaryAuthorizationAttestor
	spec := target.Spec

	// An empty project means the provider's default project -- the
	// Terraform module's google_client_config twin.
	project := strings.TrimPrefix(spec.GetProjectId().GetValue(), "projects/")
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the provider's default project for the attestor")
		}
		if clientConfig.Project == "" {
			return errors.New("the attestor names no project and the provider has no default project -- set spec.project_id or configure a project")
		}
		project = clientConfig.Project
	}

	// Binary Authorization serves the attestor; Artifact Analysis stores
	// the note and the attestations signed against it. disable_on_destroy
	// is false: policies and other attestors in the project depend on both.
	var apis []pulumi.Resource
	for _, service := range []string{"binaryauthorization.googleapis.com", "containeranalysis.googleapis.com"} {
		api, err := projects.NewService(ctx, "gcpbaatt-"+service, &projects.ServiceArgs{
			Project:                  pulumi.String(project),
			Service:                  pulumi.String(service),
			DisableDependentServices: pulumi.BoolPtr(true),
			DisableOnDestroy:         pulumi.BoolPtr(false),
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to enable %s api", service)
		}
		apis = append(apis, api)
	}

	attestorName := spec.AttestorName
	if attestorName == "" {
		attestorName = target.Metadata.Name
	}
	var deletionPolicy pulumi.StringPtrInput
	if spec.DeletionPolicy != "" {
		deletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	// The attestor's note: created here (Google's one-note-per-attestor
	// shape) or an existing one by reference.
	var noteReference pulumi.StringInput = pulumi.String(spec.GetAttestationAuthorityNote().GetNoteReference())
	var createdNote *containeranalysis.Note
	if n := spec.Note; n != nil {
		noteName := n.NoteName
		if noteName == "" {
			noteName = attestorName + "-note"
		}
		noteArgs := &containeranalysis.NoteArgs{
			Project: pulumi.StringPtr(project),
			Name:    pulumi.StringPtr(noteName),
			AttestationAuthority: &containeranalysis.NoteAttestationAuthorityArgs{
				Hint: &containeranalysis.NoteAttestationAuthorityHintArgs{
					HumanReadableName: pulumi.String(n.HumanReadableName),
				},
			},
			DeletionPolicy: deletionPolicy,
		}
		if n.ShortDescription != "" {
			noteArgs.ShortDescription = pulumi.StringPtr(n.ShortDescription)
		}
		if n.LongDescription != "" {
			noteArgs.LongDescription = pulumi.StringPtr(n.LongDescription)
		}
		if n.ExpirationTime != "" {
			noteArgs.ExpirationTime = pulumi.StringPtr(n.ExpirationTime)
		}
		if len(n.RelatedNoteNames) > 0 {
			noteArgs.RelatedNoteNames = pulumi.ToStringArray(n.RelatedNoteNames)
		}
		if len(n.RelatedUrl) > 0 {
			urls := containeranalysis.NoteRelatedUrlArray{}
			for _, u := range n.RelatedUrl {
				url := &containeranalysis.NoteRelatedUrlArgs{Url: pulumi.String(u.Url)}
				if u.Label != "" {
					url.Label = pulumi.StringPtr(u.Label)
				}
				urls = append(urls, url)
			}
			noteArgs.RelatedUrls = urls
		}
		var err error
		createdNote, err = containeranalysis.NewNote(ctx, target.Metadata.Name+"-note", noteArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn(apis))
		if err != nil {
			return errors.Wrap(err, "failed to create the attestor's note")
		}
		noteReference = createdNote.ID().ToStringOutput()
	}

	publicKeys, err := buildPublicKeys(ctx, spec.GetAttestationAuthorityNote().GetPublicKeys(), gcpProvider)
	if err != nil {
		return err
	}

	args := &binaryauthorization.AttestorArgs{
		Project:        pulumi.StringPtr(project),
		Name:           pulumi.StringPtr(attestorName),
		DeletionPolicy: deletionPolicy,
		AttestationAuthorityNote: &binaryauthorization.AttestorAttestationAuthorityNoteArgs{
			NoteReference: noteReference,
			PublicKeys:    publicKeys,
		},
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	created, err := binaryauthorization.NewAttestor(ctx, target.Metadata.Name, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn(apis))
	if err != nil {
		return errors.Wrap(err, "failed to create attestor")
	}

	delegationEmail := created.AttestationAuthorityNote.DelegationServiceAccountEmail().Elem()

	// Google requires the attestor's service account to read occurrences
	// of its note before it can verify attestations; for a note this block
	// created, the module grants it. A referenced note's owner grants it
	// there.
	if createdNote != nil {
		_, err := containeranalysis.NewNoteIamMember(ctx, target.Metadata.Name+"-reads-note", &containeranalysis.NoteIamMemberArgs{
			Project: pulumi.StringPtr(project),
			Note:    createdNote.Name,
			Role:    pulumi.String("roles/containeranalysis.notes.occurrences.viewer"),
			Member:  pulumi.Sprintf("serviceAccount:%s", delegationEmail),
		}, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to grant the attestor read access to its note")
		}
		ctx.Export(OpNoteReference, createdNote.ID())
	} else {
		ctx.Export(OpNoteReference, created.AttestationAuthorityNote.NoteReference())
	}

	ctx.Export(OpAttestorId, created.ID())
	ctx.Export(OpAttestorName, created.Name)
	ctx.Export(OpDelegationServiceAccountEmail, delegationEmail)
	return nil
}

// buildPublicKeys renders the spec's keys. A PKIX key held in Cloud KMS
// reads its public key and algorithm from the key version, the way
// Google's own example wires an attestor to Cloud KMS, and takes the key
// version's full name as its ID unless one is given -- identical to the
// Terraform module's google_kms_crypto_key_version reads.
func buildPublicKeys(ctx *pulumi.Context, keys []*gcpbinaryauthorizationattestorv1alpha1.GcpBinaryAuthorizationAttestorPublicKey, gcpProvider *gcp.Provider) (binaryauthorization.AttestorAttestationAuthorityNotePublicKeyArray, error) {
	result := binaryauthorization.AttestorAttestationAuthorityNotePublicKeyArray{}
	for _, key := range keys {
		args := &binaryauthorization.AttestorAttestationAuthorityNotePublicKeyArgs{}
		if key.Id != "" {
			args.Id = pulumi.StringPtr(key.Id)
		}
		if key.Comment != "" {
			args.Comment = pulumi.StringPtr(key.Comment)
		}
		if key.AsciiArmoredPgpPublicKey != "" {
			args.AsciiArmoredPgpPublicKey = pulumi.StringPtr(key.AsciiArmoredPgpPublicKey)
		}
		if pkix := key.PkixPublicKey; pkix != nil {
			if version := pkix.GetKmsKeyVersion().GetValue(); version != "" {
				match := keyVersionPattern.FindStringSubmatch(version)
				if match == nil {
					return nil, errors.Errorf("kms_key_version %q is not a Cloud KMS key version name", version)
				}
				number, err := strconv.Atoi(match[2])
				if err != nil {
					return nil, errors.Wrapf(err, "kms_key_version %q has no version number", version)
				}
				read, err := kms.GetKMSCryptoKeyVersion(ctx, &kms.GetKMSCryptoKeyVersionArgs{
					CryptoKey: match[1],
					Version:   &number,
				}, pulumi.Provider(gcpProvider))
				if err != nil {
					return nil, errors.Wrapf(err, "failed to read the public key of %s", version)
				}
				if len(read.PublicKeys) == 0 {
					return nil, errors.Errorf("%s has no public key -- the key's purpose must be ASYMMETRIC_SIGN", version)
				}
				args.PkixPublicKey = &binaryauthorization.AttestorAttestationAuthorityNotePublicKeyPkixPublicKeyArgs{
					PublicKeyPem:       pulumi.StringPtr(read.PublicKeys[0].Pem),
					SignatureAlgorithm: pulumi.StringPtr(read.PublicKeys[0].Algorithm),
				}
				if key.Id == "" {
					args.Id = pulumi.StringPtr(read.Id)
				}
			} else {
				args.PkixPublicKey = &binaryauthorization.AttestorAttestationAuthorityNotePublicKeyPkixPublicKeyArgs{
					PublicKeyPem:       pulumi.StringPtr(pkix.PublicKeyPem),
					SignatureAlgorithm: pulumi.StringPtr(pkix.SignatureAlgorithm),
				}
			}
		}
		result = append(result, args)
	}
	return result, nil
}
