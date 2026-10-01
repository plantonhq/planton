package verify

import (
	"context"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// dnsZoneVerifier verifies a DigitalOceanDnsZone via GET /v2/domains/{name}.
// DigitalOcean domains are identified by name, which the kind exports as
// zone_name.
type dnsZoneVerifier struct{}

func (*dnsZoneVerifier) IDOutputKey() string { return "zone_name" }

func (*dnsZoneVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := domainExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceandnszone verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceandnszone %q not found after deploy", id)
	}
	return nil
}

func (*dnsZoneVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := domainExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceandnszone verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceandnszone", ID: id}
	}
	return nil
}

func domainExists(ctx context.Context, client *godo.Client, name string) (bool, error) {
	_, _, err := client.Domains.Get(ctx, name)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// dnsRecordVerifier verifies a DigitalOceanDnsRecord. The API addresses
// records as /v2/domains/{domain}/records/{id}, so a single id is not enough:
// the verifier reads both record_id and domain from the stack outputs (the
// OutputsVerifier extension exists for exactly this shape).
type dnsRecordVerifier struct{}

func (*dnsRecordVerifier) IDOutputKey() string { return "record_id" }

func (v *dnsRecordVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandnsrecord requires the full outputs map (record_id + domain); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (v *dnsRecordVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandnsrecord requires the full outputs map (record_id + domain); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *dnsRecordVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.recordExistsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandnsrecord verify-exists failed")
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceandnsrecord %q not found after deploy", StringOutput(outputs, "record_id"))
	}
	return nil
}

func (v *dnsRecordVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.recordExistsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandnsrecord verify-absent failed")
	}
	if exists {
		return &StillExistsError{Component: "digitaloceandnsrecord", ID: StringOutput(outputs, "record_id")}
	}
	return nil
}

func (*dnsRecordVerifier) recordExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) (bool, error) {
	recordID := StringOutput(outputs, "record_id")
	domain := StringOutput(outputs, "domain")
	if recordID == "" || domain == "" {
		return false, pkgerrors.Errorf("outputs must carry record_id and domain (got record_id=%q, domain=%q)", recordID, domain)
	}
	numericID, err := strconv.Atoi(recordID)
	if err != nil {
		return false, pkgerrors.Wrapf(err, "record id %q is not the API's integer id", recordID)
	}
	// The domain vanishing is also a valid "record absent" signal: destroying
	// a zone-and-record scenario removes the domain, and the record with it.
	_, _, err = client.Domains.Record(ctx, domain, numericID)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
