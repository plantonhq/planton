package verify

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	pkgerrors "github.com/pkg/errors"
)

// serviceBusAPIVersion pins the Microsoft.ServiceBus RP version the
// verifiers read with. One constant for the whole family: the namespace,
// its entities, authorization rules, and disaster-recovery configs all
// live under the same RP and API version line.
const serviceBusAPIVersion = "2024-01-01"

// serviceBusResourceVerifier verifies any Service Bus family resource via
// the generic ARM resources GetByID (see armResourceExists), keyed on the
// resource's ARM ID output. The family shares one implementation because
// every kind's identity output IS a full ARM id under Microsoft.ServiceBus
// -- only the output key and the component name differ.
type serviceBusResourceVerifier struct {
	kind        string
	idOutputKey string
}

func (v *serviceBusResourceVerifier) IDOutputKey() string {
	return v.idOutputKey
}

func (v *serviceBusResourceVerifier) VerifyExists(ctx context.Context, cred azcore.TokenCredential, subscriptionID, id string) error {
	exists, err := armResourceExists(ctx, cred, subscriptionID, id, serviceBusAPIVersion)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-exists failed for %q", v.kind, id)
	}
	if !exists {
		return pkgerrors.Errorf("%s %q not found after deploy", v.kind, id)
	}
	return nil
}

func (v *serviceBusResourceVerifier) VerifyAbsent(ctx context.Context, cred azcore.TokenCredential, subscriptionID, id string) error {
	exists, err := armResourceExists(ctx, cred, subscriptionID, id, serviceBusAPIVersion)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-absent failed for %q", v.kind, id)
	}
	if exists {
		return pkgerrors.Errorf("%s %q still exists after destroy", v.kind, id)
	}
	return nil
}
