package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// firebaseAppProbe is what the three Firebase app verifiers share: the
// app's full resource name (projects/{p}/androidApps|iosApps|webApps/{id})
// is the handle the Firebase Management API addresses it by, and a GET on
// it returns the app's lifecycle state. The platform-specific verifiers
// below supply the GET; this file holds the judgment so the three kinds
// agree on what "exists" and "gone" mean.
//
// Exists: the app reads back ACTIVE and its app id matches the app_id
// output -- the output must be the API's own value, not a guess.
//
// Absent: an app destroyed under deletion_policy DELETE is removed with
// immediate=true, so it is gone for good -- but Firebase reports a removed
// app in two honest shapes: a 404 for the name, or the app itself in state
// DELETED (the shape a soft-deleted app carries during its 30-day window,
// which an immediate removal may still surface briefly). Either is the
// destroyed state; an ACTIVE app is the failure.
type firebaseAppState struct {
	AppId string
	State string
}

func verifyFirebaseAppExists(name string, outputs map[string]string, app *firebaseAppState, err error) error {
	if err != nil {
		return errors.Wrapf(err, "firebase app %s not readable after deploy", name)
	}
	if app.State != "ACTIVE" {
		return errors.Errorf("firebase app %s is in state %q, want ACTIVE", name, app.State)
	}
	if want := outputs["app_id"]; want != "" && app.AppId != want {
		return errors.Errorf("firebase app %s resolved to app id %q, want the app_id output %q", name, app.AppId, want)
	}
	return nil
}

func verifyFirebaseAppAbsent(name string, app *firebaseAppState, err error) error {
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing firebase app %s after destroy", name)
	}
	if app.State == "DELETED" {
		return nil
	}
	return errors.Errorf("firebase app %s still exists in state %q after destroy (deletion_policy DELETE removes it immediately)", name, app.State)
}

// firebaseAndroidAppVerifier probes an Android app registration through
// the Firebase Management API (projects/{p}/androidApps/{app_id}).
type firebaseAndroidAppVerifier struct{}

func (v *firebaseAndroidAppVerifier) IDOutputKey() string { return "name" }

func (v *firebaseAndroidAppVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	app, err := svc.Firebase.Projects.AndroidApps.Get(name).Context(ctx).Do()
	var state *firebaseAppState
	if err == nil {
		state = &firebaseAppState{AppId: app.AppId, State: app.State}
	}
	return verifyFirebaseAppExists(name, outputs, state, err)
}

func (v *firebaseAndroidAppVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	app, err := svc.Firebase.Projects.AndroidApps.Get(name).Context(ctx).Do()
	var state *firebaseAppState
	if err == nil {
		state = &firebaseAppState{AppId: app.AppId, State: app.State}
	}
	return verifyFirebaseAppAbsent(name, state, err)
}

// firebaseAppleAppVerifier probes an Apple app registration through the
// Firebase Management API (projects/{p}/iosApps/{app_id} -- the API path
// keeps the historical iosApps segment).
type firebaseAppleAppVerifier struct{}

func (v *firebaseAppleAppVerifier) IDOutputKey() string { return "name" }

func (v *firebaseAppleAppVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	app, err := svc.Firebase.Projects.IosApps.Get(name).Context(ctx).Do()
	var state *firebaseAppState
	if err == nil {
		state = &firebaseAppState{AppId: app.AppId, State: app.State}
	}
	return verifyFirebaseAppExists(name, outputs, state, err)
}

func (v *firebaseAppleAppVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	app, err := svc.Firebase.Projects.IosApps.Get(name).Context(ctx).Do()
	var state *firebaseAppState
	if err == nil {
		state = &firebaseAppState{AppId: app.AppId, State: app.State}
	}
	return verifyFirebaseAppAbsent(name, state, err)
}

// firebaseWebAppVerifier probes a web app registration through the
// Firebase Management API (projects/{p}/webApps/{app_id}).
type firebaseWebAppVerifier struct{}

func (v *firebaseWebAppVerifier) IDOutputKey() string { return "name" }

func (v *firebaseWebAppVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	app, err := svc.Firebase.Projects.WebApps.Get(name).Context(ctx).Do()
	var state *firebaseAppState
	if err == nil {
		state = &firebaseAppState{AppId: app.AppId, State: app.State}
	}
	return verifyFirebaseAppExists(name, outputs, state, err)
}

func (v *firebaseWebAppVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	app, err := svc.Firebase.Projects.WebApps.Get(name).Context(ctx).Do()
	var state *firebaseAppState
	if err == nil {
		state = &firebaseAppState{AppId: app.AppId, State: app.State}
	}
	return verifyFirebaseAppAbsent(name, state, err)
}
