package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// discoveryEngineGet reads one Discovery Engine resource by its full name
// (the pinned client library carries no Discovery Engine client). The host
// is location-prefixed exactly as Google's provider builds it --
// https://{location}-discoveryengine.googleapis.com -- for every location
// including "global"; the location is the fourth segment of every
// Discovery Engine resource name (projects/{p}/locations/{l}/...). The
// decoded body is a generic map; the status code tells a 404 from any
// other error.
func discoveryEngineGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	location := resourceLocation(name)
	if location == "" {
		return nil, 0, errors.Errorf("discovery engine resource name %q carries no location segment", name)
	}
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "discovery engine resource",
		fmt.Sprintf("https://%s-discoveryengine.googleapis.com/v1/%s", location, name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// listOutput collects a list output as the outputs transformer flattens it
// -- one dot-indexed key per element (key.0, key.1, ...), in manifest
// order.
func listOutput(outputs map[string]string, key string) []string {
	var values []string
	for i := 0; ; i++ {
		value, ok := outputs[fmt.Sprintf("%s.%d", key, i)]
		if !ok {
			break
		}
		if strings.TrimSpace(value) != "" {
			values = append(values, value)
		}
	}
	return values
}

// stringField reads a top-level string field of a decoded resource.
func stringField(obj map[string]interface{}, key string) string {
	if value, ok := obj[key].(string); ok {
		return value
	}
	return ""
}
