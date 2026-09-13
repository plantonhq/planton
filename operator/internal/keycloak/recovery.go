package keycloak

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// MasterRealm is the realm Keycloak's own administrators live in.
const MasterRealm = "master"

// ReestablishAdminInput is what re-establishing the master admin on a
// restored realm needs: the recovery admin Keycloak's recovery command
// created (a temporary account the operator made for exactly this), and the
// admin the operator authenticates as, with the password THIS install
// generated -- the one the restored realm does not know yet.
type ReestablishAdminInput struct {
	HTTPClient *http.Client
	ServerRoot string

	RecoveryUsername string
	RecoveryPassword string

	AdminUsername string
	AdminPassword string
}

// AdminNotFoundError is the restored master realm carrying no admin under
// the operator's username -- the source platform was not one of ours, or
// someone renamed the account. The operator does not invent an admin; the
// recovery credential stays so a person can.
type AdminNotFoundError struct {
	Username string
}

func (e *AdminNotFoundError) Error() string {
	return fmt.Sprintf("the restored master realm has no admin named %q", e.Username)
}

// ReestablishAdmin makes the restored realm's master admin the one this
// install knows. A database restored from an archive carries the admin
// password its SOURCE platform generated; the operator's freshly generated
// bootstrap admin Secret matches nothing, and Keycloak reads
// KC_BOOTSTRAP_ADMIN_* into an empty master realm only. So: sign in as the
// temporary recovery admin the recovery command created, give the real admin
// the new password, and remove the recovery admin -- in that order, so the
// realm never has a moment with no admin the operator can reach.
//
// Returns AdminNotFoundError (typed) when the realm has no admin under the
// operator's username; every other error is the server's own words. The
// caller retries on the reconcile cadence and, when the real admin already
// authenticates (a partial earlier run), never calls this at all.
func ReestablishAdmin(ctx context.Context, in ReestablishAdminInput) error {
	recovery := NewAdminClient(in.HTTPClient, in.ServerRoot)
	if err := recovery.Authenticate(ctx, in.RecoveryUsername, in.RecoveryPassword); err != nil {
		return fmt.Errorf("authenticating as the recovery admin %q: %w", in.RecoveryUsername, err)
	}

	admin, found, err := recovery.FindUserByUsername(ctx, MasterRealm, in.AdminUsername)
	if err != nil {
		return err
	}
	if !found {
		return &AdminNotFoundError{Username: in.AdminUsername}
	}
	adminID, _ := admin["id"].(string)
	if err := recovery.ResetUserPassword(ctx, MasterRealm, adminID, in.AdminPassword); err != nil {
		return fmt.Errorf("giving %q this install's password: %w", in.AdminUsername, err)
	}

	// The real admin now authenticates with the new password; only then is
	// the recovery admin removed, last, by itself.
	self, found, err := recovery.FindUserByUsername(ctx, MasterRealm, in.RecoveryUsername)
	if err != nil {
		return err
	}
	if found {
		selfID, _ := self["id"].(string)
		if err := recovery.DeleteUser(ctx, MasterRealm, selfID); err != nil {
			return fmt.Errorf("removing the recovery admin %q: %w", in.RecoveryUsername, err)
		}
	}
	return nil
}

// IsAdminNotFound reports whether err is the restored realm lacking the
// operator's admin.
func IsAdminNotFound(err error) bool {
	var notFound *AdminNotFoundError
	return errors.As(err, &notFound)
}
