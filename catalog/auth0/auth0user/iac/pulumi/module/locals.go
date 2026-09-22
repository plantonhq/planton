package module

import (
	"github.com/pkg/errors"
	auth0userv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0user/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// Locals contains the computed values for the Auth0 User deployment: the spec
// read once into plain Go values, with every StringValueOrRef already
// flattened (the platform resolves references before the module runs, so
// GetValue() is the resolved string) and every JSON document rendered the way
// the provider takes it.
type Locals struct {
	Auth0User *auth0userv1alpha1.Auth0User

	// ResourceName is the stable Pulumi logical resource name (metadata.name).
	// It is never sent to Auth0: the email, username, or phone number is the
	// sign-in identifier, and Auth0 assigns or takes the user id.
	ResourceName string

	ConnectionName string

	// GeneratePassword is true exactly when the modules must mint the
	// password: a database user (not passwordless) declared without one. The
	// Terraform twin computes the same predicate in locals.tf.
	GeneratePassword bool

	// UserMetadataJSON and AppMetadataJSON are the two metadata documents
	// rendered as JSON strings, which is the only form the provider accepts.
	// Empty when the spec carries no document, so nothing is sent.
	UserMetadataJSON string
	AppMetadataJSON  string

	// RoleIds is the authoritative set of role ids, already flattened from
	// their references. Empty means the user holds no roles and the roles
	// resource is not created.
	RoleIds []string

	Permissions []resolvedPermission
}

// resolvedPermission is one direct permission with its resource server
// identifier flattened from its reference.
type resolvedPermission struct {
	Name                     string
	ResourceServerIdentifier string
}

// initializeLocals creates and populates the Locals struct from the stack input.
func initializeLocals(ctx *pulumi.Context, stackInput *auth0userv1alpha1.Auth0UserStackInput) (*Locals, error) {
	locals := &Locals{}

	locals.Auth0User = stackInput.Target

	spec := stackInput.Target.Spec
	metadata := stackInput.Target.Metadata

	locals.ResourceName = metadata.Name
	locals.ConnectionName = spec.ConnectionName.GetValue()

	// A declared password wins; a passwordless connection takes none; a
	// database user with neither gets one minted.
	locals.GeneratePassword = !spec.Passwordless && spec.Password == ""

	var err error
	if locals.UserMetadataJSON, err = renderMetadata(spec.UserMetadata); err != nil {
		return nil, errors.Wrap(err, "failed to render user_metadata as JSON")
	}
	if locals.AppMetadataJSON, err = renderMetadata(spec.AppMetadata); err != nil {
		return nil, errors.Wrap(err, "failed to render app_metadata as JSON")
	}

	for _, role := range spec.Roles {
		if role != nil && role.GetValue() != "" {
			locals.RoleIds = append(locals.RoleIds, role.GetValue())
		}
	}

	for _, permission := range spec.Permissions {
		if permission == nil {
			continue
		}
		locals.Permissions = append(locals.Permissions, resolvedPermission{
			Name:                     permission.Name,
			ResourceServerIdentifier: permission.ResourceServerIdentifier.GetValue(),
		})
	}

	return locals, nil
}

// renderMetadata turns a metadata Struct into the JSON string the provider
// takes. A nil or empty document renders as "" so the argument is left unset
// rather than sending "{}" and having the provider record an authored-looking
// empty object.
func renderMetadata(document *structpb.Struct) (string, error) {
	if document == nil || len(document.GetFields()) == 0 {
		return "", nil
	}
	rendered, err := protojson.Marshal(document)
	if err != nil {
		return "", err
	}
	return string(rendered), nil
}
