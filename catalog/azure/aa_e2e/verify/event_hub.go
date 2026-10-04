package verify

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	pkgerrors "github.com/pkg/errors"
)

// eventHubAPIVersion pins the Microsoft.EventHub RP version the verifiers
// read with. One constant for the whole family: the namespace, its hubs,
// consumer groups, authorization rules, schema groups, disaster-recovery
// configs, and dedicated clusters all live under the same RP and API
// version line.
const eventHubAPIVersion = "2024-01-01"

// eventHubResourceVerifier verifies any Event Hubs family resource via the
// generic ARM resources GetByID (see armResourceExists), keyed on the
// resource's ARM ID output. The family shares one implementation because
// every kind's identity output IS a full ARM id under Microsoft.EventHub
// -- only the output key and the component name differ. (The CMK kind's
// output is the parent namespace's id -- its absence check is therefore
// meaningless and that kind carries no verifier registration; CMK is
// add-only by Azure's own contract.)
type eventHubResourceVerifier struct {
	kind        string
	idOutputKey string
}

func (v *eventHubResourceVerifier) IDOutputKey() string {
	return v.idOutputKey
}

func (v *eventHubResourceVerifier) VerifyExists(ctx context.Context, cred azcore.TokenCredential, subscriptionID, id string) error {
	exists, err := armResourceExists(ctx, cred, subscriptionID, id, eventHubAPIVersion)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-exists failed for %q", v.kind, id)
	}
	if !exists {
		return pkgerrors.Errorf("%s %q not found after deploy", v.kind, id)
	}
	return nil
}

func (v *eventHubResourceVerifier) VerifyAbsent(ctx context.Context, cred azcore.TokenCredential, subscriptionID, id string) error {
	exists, err := armResourceExists(ctx, cred, subscriptionID, id, eventHubAPIVersion)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-absent failed for %q", v.kind, id)
	}
	if exists {
		return pkgerrors.Errorf("%s %q still exists after destroy", v.kind, id)
	}
	return nil
}
