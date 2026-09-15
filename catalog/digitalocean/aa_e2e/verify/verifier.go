// Package verify implements per-component resource verification for the
// DigitalOcean E2E harness. Each verifier answers two questions through the
// DigitalOcean REST API (godo): does the resource exist after deploy, and is
// it gone after destroy. A 404 from the API is the ONLY absence signal; every
// other error surfaces as a genuine failure, so flaky credentials or rate
// limits can never masquerade as a passing destroy verification.
//
// The one exception to godo is the Spaces bucket verifier: Spaces is an
// S3-compatible credential plane the API token cannot reach, so that verifier
// speaks the S3 API against the bucket's regional Spaces endpoint (see
// bucket.go).
package verify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// Verifier checks a single component's DigitalOcean resource for
// existence/absence. DigitalOcean lookups are account-scoped by id, so there
// is no region parameter (region is a property of the resource, not of the
// API endpoint).
type Verifier interface {
	// IDOutputKey is the stack-output key carrying the identifier used to
	// verify the resource (e.g. "vpc_id"). The key names come from each
	// kind's outputs.proto -- they are contract, not convention.
	IDOutputKey() string
	// VerifyExists returns an error unless the resource exists.
	VerifyExists(ctx context.Context, client *godo.Client, id string) error
	// VerifyAbsent returns an error unless the resource is gone.
	VerifyAbsent(ctx context.Context, client *godo.Client, id string) error
}

// OutputsVerifier inspects the full stack output map when a single string id
// is insufficient (e.g. a DNS record is addressed by domain + record id, and
// a Spaces bucket by region + name).
type OutputsVerifier interface {
	Verifier
	VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error
	VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error
}

// verifiers maps component slugs (the catalog directory names) to their
// verifiers. Every kind that appears in another kind's registry prerequisites
// MUST have an entry here, or composed scenarios fail at DEPENDENCIES-UP.
var verifiers = map[string]Verifier{
	"digitaloceanapp":                    &appVerifier{component: "digitaloceanapp", idOutputKey: "app_id"},
	"digitaloceanbucket":                 &bucketVerifier{},
	"digitaloceancdn":                    &cdnVerifier{},
	"digitaloceancertificate":            &certificateVerifier{},
	"digitaloceancontainerregistry":      &containerRegistryVerifier{},
	"digitaloceandatabasecluster":        &databaseClusterVerifier{},
	"digitaloceandatabaseconnectionpool": &databaseConnectionPoolVerifier{},
	"digitaloceandatabasedb":             &databaseDbVerifier{},
	"digitaloceandatabasefirewall":       &databaseFirewallVerifier{},
	"digitaloceandatabasekafkaschema":    &kafkaSchemaVerifier{},
	"digitaloceandatabasekafkatopic":     &kafkaTopicVerifier{},
	"digitaloceandatabasereplica":        &databaseReplicaVerifier{},
	"digitaloceandatabaseuser":           &databaseUserVerifier{},
	"digitaloceandnsrecord":              &dnsRecordVerifier{},
	"digitaloceandnszone":                &dnsZoneVerifier{},
	"digitaloceandroplet":                &dropletVerifier{},
	"digitaloceandropletautoscalepool":   &dropletAutoscalePoolVerifier{},
	"digitaloceanfirewall":               &firewallVerifier{},
	"digitaloceanfunction":               &appVerifier{component: "digitaloceanfunction", idOutputKey: "function_id"},
	"digitaloceankubernetescluster":      &kubernetesClusterVerifier{},
	"digitaloceankubernetesnodepool":     &kubernetesNodePoolVerifier{},
	"digitaloceanloadbalancer":           &loadBalancerVerifier{},
	"digitaloceanmonitoralert":           &monitorAlertVerifier{},
	"digitaloceanproject":                &projectVerifier{},
	"digitaloceanreservedip":             &reservedIpVerifier{},
	"digitaloceanspaceskey":              &spacesKeyVerifier{},
	"digitaloceansshkey":                 &sshKeyVerifier{},
	"digitaloceanuptimecheck":            &uptimeCheckVerifier{},
	"digitaloceanvolume":                 &volumeVerifier{},
	"digitaloceanvpc":                    &vpcVerifier{},
	"digitaloceanvpcpeering":             &vpcPeeringVerifier{},
}

// StillExistsError is the ONE error a VerifyAbsent returns when the API still
// answers for a destroyed resource. It is a type, not a string, so the
// harness can tell "not gone yet" from every other failure (auth, rate
// limit, a broken lookup) and poll only the former: DigitalOcean's
// read-after-delete is eventually consistent -- a GET issued a second after
// a successful DELETE can still answer 200 (measured live on volumes:
// destroy returned in ~2s, the probe 1s later saw the volume, and it read
// 404 seconds after that) -- while a genuine API error must fail the phase
// immediately rather than be retried into a timeout.
type StillExistsError struct {
	// Component is the catalog slug of the kind whose resource lingers.
	Component string
	// ID is the identifier the probe used (a UUID, a name, or a composite).
	ID string
	// Detail optionally replaces the default "still exists after destroy"
	// clause for verifiers whose absence is a state, not a 404 (a firewall
	// whose rule set must be empty).
	Detail string
}

func (e *StillExistsError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s %q %s", e.Component, e.ID, e.Detail)
	}
	return fmt.Sprintf("%s %q still exists after destroy", e.Component, e.ID)
}

// IsStillExists reports whether err (anywhere in its chain) is a
// StillExistsError -- the harness's signal to keep polling.
func IsStillExists(err error) bool {
	var target *StillExistsError
	return errors.As(err, &target)
}

// GetVerifier returns the verifier for a component, or an error if none is registered.
func GetVerifier(component string) (Verifier, error) {
	v, ok := verifiers[component]
	if !ok {
		return nil, pkgerrors.Errorf("no DigitalOcean verifier registered for component %q", component)
	}
	return v, nil
}

// isNotFound reports whether err is the DigitalOcean API's 404. godo wraps
// every non-2xx response in *godo.ErrorResponse carrying the raw
// *http.Response, so the status code is checked typed, never by matching
// error strings.
func isNotFound(err error) bool {
	var errResp *godo.ErrorResponse
	if errors.As(err, &errResp) && errResp.Response != nil {
		return errResp.Response.StatusCode == http.StatusNotFound
	}
	return false
}

// StringOutput reads a string-valued stack output, tolerating non-string
// scalars: DigitalOcean's numeric ids (droplets, DNS records) may decode as
// float64 or json.Number depending on the engine's JSON path, and a float64
// rendered with %v would turn 12345678 into "1.2345678e+07". Exported because
// the harness reads outputs with the same care.
func StringOutput(outputs map[string]interface{}, key string) string {
	if outputs == nil {
		return ""
	}
	v, ok := outputs[key]
	if !ok {
		return ""
	}
	return scalarString(v)
}

// StringSliceOutput reads a list-valued stack output (a `repeated string`
// in the outputs contract) as []string, tolerating a missing or empty list
// and applying StringOutput's scalar care to every element. Order is the
// engine's; callers that compare sets must not rely on it.
func StringSliceOutput(outputs map[string]interface{}, key string) []string {
	if outputs == nil {
		return nil
	}
	raw, ok := outputs[key].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, scalarString(v))
	}
	return out
}

// StringMapOutput reads a map-valued stack output (a `map<string, string>`
// in the outputs contract -- the per-instance id maps that keyed blind
// imports derive from) as map[string]string, tolerating a missing or empty
// map and applying StringOutput's scalar care to every value.
func StringMapOutput(outputs map[string]interface{}, key string) map[string]string {
	out := map[string]string{}
	if outputs == nil {
		return out
	}
	raw, ok := outputs[key].(map[string]interface{})
	if !ok {
		return out
	}
	for k, v := range raw {
		out[k] = scalarString(v)
	}
	return out
}

// scalarString renders one decoded JSON scalar the way StringOutput does.
func scalarString(v interface{}) string {
	switch n := v.(type) {
	case string:
		return n
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case json.Number:
		return n.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
