package verify

import (
	"context"
	"strings"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// appVerifier verifies components backed by an App Platform app via
// GET /v2/apps/{id}. Two kinds share it: DigitalOceanApp (the app is the
// component) and DigitalOceanFunction (the provider has no standalone
// Functions resource, so the component deploys an app carrying a functions
// section and its function_id output IS the app id). The struct is
// parameterized the same way the AWS harness reuses one verifier across
// load-balancer kinds.
//
// Beyond existence, the outputs form asserts the live app has an ACTIVE
// deployment (read live, never from an output -- an apply-time snapshot goes
// stale) and checks the URL-shaped outputs the module CLAIMS against the live
// app. Outputs are contractually identical across both engines, so one
// assertion protects both; an absent output means "not claimed" and is
// skipped. The two kinds name the same URL differently (live_url on the App,
// https_endpoint on the Function), so the key is a field, like the id key.
type appVerifier struct {
	component    string
	idOutputKey  string
	urlOutputKey string
}

func (v *appVerifier) IDOutputKey() string { return v.idOutputKey }

func (v *appVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	_, err := getApp(ctx, client, v.component, id)
	return err
}

func (v *appVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	_, _, err := client.Apps.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return pkgerrors.Wrapf(err, "%s verify-absent failed for %q", v.component, id)
	}
	return &StillExistsError{Component: v.component, ID: id}
}

func (v *appVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, v.idOutputKey)
	if id == "" {
		return pkgerrors.Errorf("%s output missing after deploy", v.idOutputKey)
	}

	app, err := getApp(ctx, client, v.component, id)
	if err != nil {
		return err
	}

	// The provider's create waits for the first deployment to reach ACTIVE,
	// so anything else here means the module returned before the app served
	// traffic -- or the deployment failed after the fact.
	if app.ActiveDeployment == nil {
		return pkgerrors.Errorf("%s %q has no active deployment after deploy", v.component, id)
	}
	if app.ActiveDeployment.Phase != godo.DeploymentPhase_Active {
		return pkgerrors.Errorf("%s %q active deployment phase is %q, want %q",
			v.component, id, app.ActiveDeployment.Phase, godo.DeploymentPhase_Active)
	}

	if url := StringOutput(outputs, v.urlOutputKey); url != "" && app.LiveURL != url {
		return pkgerrors.Errorf("%s %q %s mismatch: output %q, live %q",
			v.component, id, v.urlOutputKey, url, app.LiveURL)
	}
	// Both modules export default_hostname as the default ingress with its
	// scheme stripped; compare against the same normalization of the live
	// value so the assertion tests the contract, not the formatting.
	if host := StringOutput(outputs, "default_hostname"); host != "" && stripScheme(app.DefaultIngress) != host {
		return pkgerrors.Errorf("%s %q default_hostname mismatch: output %q, live %q",
			v.component, id, host, stripScheme(app.DefaultIngress))
	}
	if domain := StringOutput(outputs, "live_domain"); domain != "" && app.LiveDomain != domain {
		return pkgerrors.Errorf("%s %q live_domain mismatch: output %q, live %q",
			v.component, id, domain, app.LiveDomain)
	}

	return nil
}

func (v *appVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	id := StringOutput(outputs, v.idOutputKey)
	if id == "" {
		return pkgerrors.Errorf("%s output missing for destroy verification", v.idOutputKey)
	}
	return v.VerifyAbsent(ctx, client, id)
}

func getApp(ctx context.Context, client *godo.Client, component, id string) (*godo.App, error) {
	app, _, err := client.Apps.Get(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, pkgerrors.Errorf("%s %q not found after deploy", component, id)
		}
		return nil, pkgerrors.Wrapf(err, "%s verify-exists failed for %q", component, id)
	}
	return app, nil
}

func stripScheme(u string) string {
	u = strings.TrimPrefix(u, "https://")
	return strings.TrimPrefix(u, "http://")
}
