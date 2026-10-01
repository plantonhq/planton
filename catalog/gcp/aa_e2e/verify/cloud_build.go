package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// isGoogleNotFound reports whether a typed Google API read answered 404.
// Cloud Build and Cloud Deploy share it.
func isGoogleNotFound(err error) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == 404
}

// cloudBuildWorkerPoolVerifier probes a private pool through the Cloud
// Build API by the name output
// (projects/{p}/locations/{l}/workerPools/{id}). Posture assertion: Google's
// uid matches the exported one. Pools carry no labels, so the attribution
// canary does not apply.
type cloudBuildWorkerPoolVerifier struct{}

func (v *cloudBuildWorkerPoolVerifier) IDOutputKey() string { return "name" }

func (v *cloudBuildWorkerPoolVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	pool, err := svc.CloudBuild.Projects.Locations.WorkerPools.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "worker pool %s not found after deploy", name)
	}
	if uid := outputs["uid"]; uid != "" && pool.Uid != uid {
		return errors.Errorf("worker pool %s uid output %q does not match live uid %q", name, uid, pool.Uid)
	}
	return nil
}

func (v *cloudBuildWorkerPoolVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	pool, err := svc.CloudBuild.Projects.Locations.WorkerPools.Get(name).Context(ctx).Do()
	if err != nil {
		if isGoogleNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing worker pool %s after destroy", name)
	}
	if pool.State == "DELETING" || pool.State == "DELETED" || pool.DeleteTime != "" {
		return nil
	}
	return errors.Errorf("worker pool %s still exists after destroy (state %s)", name, pool.State)
}

// cloudBuildConnectionVerifier probes a repository connection through the
// Cloud Build v2 API by the name output
// (projects/{p}/locations/{l}/connections/{id}). Posture assertion: the
// installation is complete -- a connection still waiting on its GitHub App
// installation has not proven the credentials. Connections carry no
// labels, so the attribution canary does not apply.
type cloudBuildConnectionVerifier struct{}

func (v *cloudBuildConnectionVerifier) IDOutputKey() string { return "name" }

func (v *cloudBuildConnectionVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	connection, err := svc.CloudBuildV2.Projects.Locations.Connections.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "connection %s not found after deploy", name)
	}
	if connection.InstallationState == nil || connection.InstallationState.Stage != "COMPLETE" {
		stage := ""
		if connection.InstallationState != nil {
			stage = connection.InstallationState.Stage
		}
		return errors.Errorf("connection %s installation is not complete (stage %q)", name, stage)
	}
	return nil
}

func (v *cloudBuildConnectionVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CloudBuildV2.Projects.Locations.Connections.Get(name).Context(ctx).Do()
	if err == nil {
		return errors.Errorf("connection %s still exists after destroy", name)
	}
	if isGoogleNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing connection %s after destroy", name)
}

// cloudBuildRepositoryVerifier probes a repository link through the Cloud
// Build v2 API by the name output
// (projects/{p}/locations/{l}/connections/{c}/repositories/{id}). Posture
// assertion: the live remote URI matches the exported one.
type cloudBuildRepositoryVerifier struct{}

func (v *cloudBuildRepositoryVerifier) IDOutputKey() string { return "name" }

func (v *cloudBuildRepositoryVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	repository, err := svc.CloudBuildV2.Projects.Locations.Connections.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "repository %s not found after deploy", name)
	}
	if uri := outputs["remote_uri"]; uri != "" && repository.RemoteUri != uri {
		return errors.Errorf("repository %s remote_uri output %q does not match live %q", name, uri, repository.RemoteUri)
	}
	return nil
}

func (v *cloudBuildRepositoryVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CloudBuildV2.Projects.Locations.Connections.Repositories.Get(name).Context(ctx).Do()
	if err == nil {
		return errors.Errorf("repository %s still exists after destroy", name)
	}
	if isGoogleNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing repository %s after destroy", name)
}
