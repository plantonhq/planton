package verify

import (
	"context"
	"strings"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// reservedIpVerifier verifies a DigitalOceanReservedIp. The address itself
// is the resource identity; the API splits IPv4 and IPv6 reservations into
// separate endpoints (GET /v2/reserved_ips/{ip} vs
// GET /v2/reserved_ipv6/{ip}), so the verifier routes by address family --
// an IPv6 address always contains a colon. The live absence check after
// destroy carries extra weight for the v6 family: the provider's delete
// swallows non-404 errors at the pin, so a "successful" destroy is only
// proven here.
type reservedIpVerifier struct{}

func (*reservedIpVerifier) IDOutputKey() string { return "reserved_ip_address" }

func (*reservedIpVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	exists, err := reservedIpExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanreservedip verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceanreservedip %q not found after deploy", id)
	}
	return nil
}

func (*reservedIpVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	exists, err := reservedIpExists(ctx, client, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "digitaloceanreservedip verify-absent failed for %q", id)
	}
	if exists {
		return &StillExistsError{Component: "digitaloceanreservedip", ID: id}
	}
	return nil
}

func reservedIpExists(ctx context.Context, client *godo.Client, address string) (bool, error) {
	var err error
	if strings.Contains(address, ":") {
		_, _, err = client.ReservedIPV6s.Get(ctx, address)
	} else {
		_, _, err = client.ReservedIPs.Get(ctx, address)
	}
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
