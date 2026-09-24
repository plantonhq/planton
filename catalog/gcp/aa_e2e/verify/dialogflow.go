package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// dialogflowGet reads one Dialogflow CX resource by its full name (the
// pinned client library carries no Dialogflow CX client). The host is
// location-prefixed exactly as Google's provider builds it --
// https://{location}-dialogflow.googleapis.com/v3/ -- for every location
// including "global". A generative-settings name carries its language as a
// query string ({agent}/generativeSettings?languageCode=en), which the GET
// passes through unchanged. The decoded body is a generic map; the status
// code tells a 404 from any other error.
func dialogflowGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	location := resourceLocation(name)
	if location == "" {
		return nil, 0, errors.Errorf("dialogflow resource name %q carries no location segment", name)
	}
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "dialogflow resource",
		fmt.Sprintf("https://%s-dialogflow.googleapis.com/v3/%s", location, name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}
