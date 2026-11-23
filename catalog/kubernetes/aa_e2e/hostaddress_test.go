package aa_e2e

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeCommands answers the two commands ResolveHostAddress runs.
func fakeCommands(t *testing.T, getent, gateways string, getentErr, gatewayErr error) {
	t.Helper()
	previous := commandOutput
	t.Cleanup(func() { commandOutput = previous })
	commandOutput = func(_ context.Context, name string, args ...string) (string, error) {
		joined := name + " " + strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "getent ahostsv4 host.docker.internal"):
			return getent, getentErr
		case strings.Contains(joined, "network inspect kind"):
			return gateways, gatewayErr
		}
		t.Fatalf("unexpected command %q", joined)
		return "", nil
	}
}

func TestHostAddressPrefersHostDockerInternal(t *testing.T) {
	fakeCommands(t, "192.168.65.254  STREAM host.docker.internal\n192.168.65.254  DGRAM\n", "172.18.0.1 ", nil, nil)
	got, err := ResolveHostAddress(context.Background(), "stigmer-e2e")
	if err != nil {
		t.Fatal(err)
	}
	if got != (HostAddress{Source: "host.docker.internal", IP: "192.168.65.254"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestHostAddressFallsBackToTheKindGatewayIPv4(t *testing.T) {
	fakeCommands(t, "", "fc00:f853:ccd:e793::1 172.18.0.1 ", errors.New("exit status 2"), nil)
	got, err := ResolveHostAddress(context.Background(), "stigmer-e2e")
	if err != nil {
		t.Fatal(err)
	}
	if got != (HostAddress{Source: "the kind network's gateway", IP: "172.18.0.1"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestHostAddressNamesBothAttemptsWhenNeitherWorks(t *testing.T) {
	fakeCommands(t, "", "fc00::1 ", errors.New("exit status 2"), nil)
	_, err := ResolveHostAddress(context.Background(), "stigmer-e2e")
	if err == nil || !strings.Contains(err.Error(), "host.docker.internal does not resolve on node stigmer-e2e-control-plane") || !strings.Contains(err.Error(), "no IPv4 gateway") {
		t.Fatalf("err %v, want both attempts named", err)
	}
}
